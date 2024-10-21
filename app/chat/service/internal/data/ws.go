package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	pb "im-service/api/chat/service/v1"
	"im-service/app/chat/service/cmd/service/handler"
	"im-service/app/chat/service/internal/biz"
	"im-service/app/chat/service/internal/consts"
	"im-service/app/chat/service/internal/model"
	"im-service/app/chat/service/utils"
	"strconv"
)

type wsRepo struct {
	data          *Data
	groupChatColl *mongo.Collection
	log           *log.Helper
}

func (w *wsRepo) BindMember(ctx context.Context, req *pb.BindMemberRequest) (*pb.BindMemberReply, error) {
	var err error
	if c, IsExist := w.data.clientManager.Clients[req.ClientId]; IsExist && w.data.clientManager.Clients[req.ClientId].MemberId == 0 {

		c.Lock()
		defer func() {
			c.Unlock()
		}()

		mId := strconv.Itoa(int(req.MemberId))
		ClientInfo := utils.WSClientInfo{
			IP:    w.data.IP,
			Port:  w.data.Port,
			MQTag: w.data.mqTag,
		}
		err = utils.AddOrUpdateUserInfo(ctx, w.data.r, mId, ClientInfo)
		if err != nil {
			return nil, err
		}

		w.data.clientManager.Clients[req.ClientId].MemberId = uint(req.MemberId)
		w.data.clientManager.MapUserIdToClientId[uint(req.MemberId)] = req.ClientId
	} else {
		err = errors.New(consts.GRPC_ERROR, "客户端不存在或者已被绑定", "客户端不存在或者已被绑定")
	}
	return &pb.BindMemberReply{}, err
}

func (w *wsRepo) BindGroup(ctx context.Context, req *pb.BindGroupRequest) (*pb.BindGroupReply, error) {
	clientId := w.data.clientManager.MapUserIdToClientId[uint(req.MemberId)]
	if c, IsExist := w.data.clientManager.Clients[clientId]; IsExist {
		c.Lock()
		defer func() {
			c.Unlock()
		}()

		BindGroup := model.GroupBind{
			GroupId:  req.GroupId,
			MemberId: req.MemberId,
		}

		_ = w.data.mysqlClient.FirstOrCreate(&BindGroup)

		err := utils.BindUserToGroup(ctx, w.data.r, strconv.Itoa(int(req.MemberId)), req.GroupId)
		if err != nil {
			return &pb.BindGroupReply{}, err
		}

		w.data.clientManager.Clients[clientId].GroupId = req.GroupId

		return &pb.BindGroupReply{}, nil
	} else {
		return &pb.BindGroupReply{}, errors.New(consts.GRPC_ERROR, "客户端不存在", "客户端不存在")
	}

}

func (w *wsRepo) CancelGroup(ctx context.Context, req *pb.CancelGroupRequest) (*pb.CancelGroupReply, error) {
	clientId := w.data.clientManager.MapUserIdToClientId[uint(req.MemberId)]
	if c := w.data.clientManager.Clients[clientId]; c.GroupId != "" {

		c.Lock()
		defer func() {
			c.Unlock()
		}()

		mId := strconv.Itoa(int(req.MemberId))
		err := utils.UnbindUserFromGroup(ctx, w.data.r, w.data.clientManager.Clients[clientId].GroupId, mId)
		if err != nil {
			return &pb.CancelGroupReply{}, err
		}
		w.data.clientManager.Clients[clientId].GroupId = ""
	}

	return &pb.CancelGroupReply{}, nil
}

func (w *wsRepo) SendMsg(ctx context.Context, req *pb.SendMsgRequest) (*pb.SendMsgReply, error) {
	userLinkInfo, err := utils.GetUserInfo(ctx, w.data.r, strconv.Itoa(int(req.MemberId)))
	if err != nil {
		return &pb.SendMsgReply{}, err
	}

	Ip, _ := utils.GetLocalIP()

	msg := handler.ReplyMsg{
		Code:    consts.WsSuccess,
		Type:    consts.SysMsg,
		Content: req.Content,
	}

	if Ip == userLinkInfo.IP {
		clientId := w.data.clientManager.MapUserIdToClientId[uint(req.MemberId)]
		if clientId != "" {
			client := w.data.clientManager.Clients[clientId]
			client.Send <- msg
		}
	} else {
		byteData, _ := json.Marshal(req.Content)
		MqMsg := utils.MqMsg{
			Type:   consts.MQ_MSG_TYEP_MEMBER,
			SendTo: strconv.Itoa(int(req.MemberId)),
			Body:   byteData,
		}

		JsonData, _ := json.Marshal(MqMsg)
		err = utils.SendMqMsg(w.data.p, "chatMessage", userLinkInfo.MQTag, JsonData)
		if err != nil {
			return &pb.SendMsgReply{}, err
		}
	}

	// TODO 系统消息暂不入库

	return &pb.SendMsgReply{}, nil
}

func (w *wsRepo) DelGroupColl(ctx context.Context, req *pb.DelGroupCollRequest) (*pb.DelGroupCollReply, error) {
	coll := w.data.mongo.Collection("group_messages")
	// 创建过滤条件
	filter := bson.M{
		"group_id": req.GroupId,
		"memberId": req.MemberId,
	}

	// 删除文档
	_, err := coll.DeleteMany(ctx, filter)
	if err != nil {
		return nil, errors.New(consts.GRPC_ERROR, fmt.Sprintf("%w", err), "聊天记录删除失败")
	}

	return &pb.DelGroupCollReply{}, nil
}

func (w *wsRepo) GetGroupHistory(ctx context.Context, req *pb.GetGroupHistoryRequest) (*pb.GetGroupHistoryReply, error) {
	var messages []handler.MongoMsgData
	coll := w.data.mongo.Collection("group_messages")

	filter := bson.M{
		"group_id": req.GroupId,
		"memberId": req.MemberId,
	}

	// 如果提供了 lastID，则从该消息之后开始查询
	if req.LastId != "" {
		lastObjectId, err := primitive.ObjectIDFromHex(req.LastId)
		if err != nil {
			w.log.Errorf("Invalid lastID: %v", err)
			return nil, errors.New(consts.GRPC_ERROR, fmt.Sprintf("%w", err), "非法ID")
		}
		// 在提供了 lastID 的情况下，查询比 lastID 更早的消息
		filter["_id"] = bson.M{"$lt": lastObjectId}
	}

	opts := options.Find().
		SetSort(bson.D{{"_id", -1}}). // 按 _id 倒序排列
		SetLimit(int64(req.PageSize)) // 设置分页大小

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		w.log.Errorf("Find 操作失败: %v", err)
		return nil, errors.New(consts.GRPC_ERROR, fmt.Sprintf("%w", err), "记录查询失败")
	}

	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &messages); err != nil {
		w.log.Errorf("Cursor All 操作失败: %v", err)
		return nil, errors.New(consts.GRPC_ERROR, fmt.Sprintf("%w", err), "Cursor All 操作失败")
	}

	var groupMsg []*pb.MsgHistory

	for _, msg := range messages {
		groupMsg = append(groupMsg, &pb.MsgHistory{
			Id:          msg.ID.Hex(),
			MessageId:   msg.MessageBody.MessageId,
			MemberId:    msg.MessageBody.MemberId,
			ContentType: msg.MessageBody.ContentType,
			Content:     msg.MessageBody.Content,
			CreateAt:    msg.MessageBody.CreateAt.Format("2006-01-02 15:04:05"),
		})
	}
	return &pb.GetGroupHistoryReply{MsgList: groupMsg}, nil
}

// NewWsRepo .
func NewWsRepo(data *Data, logger log.Logger) biz.WsRepo {
	return &wsRepo{
		data:          data,
		groupChatColl: data.mongo.Collection("group_chat"),
		log:           log.NewHelper(logger),
	}
}
