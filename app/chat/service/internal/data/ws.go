package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kratos/kratos/v2/errors"
	pb "im-service/api/chat/service/v1"
	"im-service/app/chat/service/cmd/service/handler"
	"im-service/app/chat/service/internal/consts"
	"im-service/app/chat/service/utils"
	"strconv"

	"im-service/app/chat/service/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type wsRepo struct {
	data *Data
	log  *log.Helper
}

func (w wsRepo) BindMember(ctx context.Context, req *pb.BindMemberRequest) (*pb.BindMemberReply, error) {
	/*clientLen := len(w.data.clientManager.Clients)
	log.Info(fmt.Sprintf("现存客户端数量为%d", clientLen))
	log.Info(w.data.clientManager.Clients)*/
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
		err = errors.New(consts.GRPC_ERROR, "客户端不存在", "客户端不存在")
	}
	return &pb.BindMemberReply{}, err
}

func (w wsRepo) BindGroup(ctx context.Context, req *pb.BindGroupRequest) (*pb.BindGroupReply, error) {
	clientId := w.data.clientManager.MapUserIdToClientId[uint(req.MemberId)]
	if c, IsExist := w.data.clientManager.Clients[clientId]; IsExist {
		c.Lock()
		defer func() {
			c.Unlock()
		}()

		err := utils.BindUserToGroup(ctx, w.data.r, strconv.Itoa(int(req.MemberId)), req.GroupId)
		if err != nil {
			return &pb.BindGroupReply{}, err
		}

		w.data.clientManager.Clients[clientId].GroupId = req.GroupId
		return &pb.BindGroupReply{}, nil
	} else {
		return &pb.BindGroupReply{}, errors.New(-1, "用户不存在", "用户不存在")
	}

}

func (w wsRepo) CancelGroup(ctx context.Context, req *pb.CancelGroupRequest) (*pb.CancelGroupReply, error) {
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

func (w wsRepo) SendMsg(ctx context.Context, req *pb.SendMsgRequest) (*pb.SendMsgReply, error) {
	userLinkInfo, err := utils.GetUserInfo(ctx, w.data.r, strconv.Itoa(int(req.MemberId)))
	if err != nil {
		return &pb.SendMsgReply{}, err
	}

	Ip, _ := utils.GetLocalIP()

	msg := handler.ReplyMsg{
		Code:    consts.WsSuccess,
		Type:    handler.SysMsg,
		Content: req.Content,
	}

	if Ip == userLinkInfo.IP {
		clientId := w.data.clientManager.MapUserIdToClientId[uint(req.MemberId)]
		if clientId != "" {
			client := w.data.clientManager.Clients[clientId]

			err := client.SendMessage(msg)
			if err != nil {
				return nil, err
			}
		}
	} else {
		byteData, _ := json.Marshal(msg)
		MqMsg := utils.MqMsg{
			Type:   consts.MQ_MSG_TYEP_MEMBER,
			SendTo: strconv.Itoa(int(req.MemberId)),
			Body:   byteData,
		}

		JsonData, _ := json.Marshal(MqMsg)
		err = utils.SendMqMsg(w.data.p, "chatMessage", userLinkInfo.MQTag, JsonData)
		fmt.Println(fmt.Sprintf("当前推送tag:%s", userLinkInfo.MQTag))
		if err != nil {
			return &pb.SendMsgReply{}, err
		}
	}

	// TODO 消息存入数据库

	return &pb.SendMsgReply{}, nil
}

// NewWsRepo .
func NewWsRepo(data *Data, logger log.Logger) biz.WsRepo {
	return &wsRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}
