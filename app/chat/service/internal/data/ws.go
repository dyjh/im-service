package data

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/errors"
	pb "im-service/api/chat/service/v1"
	"im-service/app/chat/service/cmd/service/handler"
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
	clientLen := len(handler.Manager.Clients)
	log.Info(fmt.Sprintf("现存客户端数量为%d", clientLen))
	//TODO implement me
	return nil, nil
}

func (w wsRepo) BindGroup(ctx context.Context, req *pb.BindGroupRequest) (*pb.BindGroupReply, error) {
	clientId := handler.Manager.MapUserIdToClientId[uint(req.MemberId)]
	if clientId == "" {
		return &pb.BindGroupReply{}, errors.New(-1, "用户不存在", "用户不存在")
	}

	handler.Manager.Clients[clientId].GroupId = req.GroupId
	mId := strconv.Itoa(int(req.MemberId))
	ClientInfo := utils.WSClientInfo{
		IP:    w.data.IP,
		Port:  w.data.Port,
		MQTag: w.data.mqTag,
	}
	err := utils.AddUserToGroup(ctx, w.data.r, handler.Manager.Clients[clientId].GroupId, mId, ClientInfo)

	return &pb.BindGroupReply{}, err
}

func (w wsRepo) CancelGroup(ctx context.Context, req *pb.CancelGroupRequest) (*pb.CancelGroupReply, error) {
	clientId := handler.Manager.MapUserIdToClientId[uint(req.MemberId)]
	if handler.Manager.Clients[clientId].GroupId != "" {
		mId := strconv.Itoa(int(req.MemberId))
		_ = utils.RemoveUserFromGroup(ctx, w.data.r, handler.Manager.Clients[clientId].GroupId, mId)
		handler.Manager.Clients[clientId].GroupId = ""
	}

	return &pb.CancelGroupReply{}, nil
}

func (w wsRepo) SendMsg(ctx context.Context, req *pb.SendMsgRequest) (*pb.SendMsgReply, error) {
	//TODO implement me
	return nil, nil
}

// NewWsRepo .
func NewWsRepo(data *Data, logger log.Logger) biz.WsRepo {
	return &wsRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}
