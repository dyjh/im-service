package data

import (
	"context"
	"fmt"
	pb "im-service/api/chat/service/v1"
	"im-service/app/chat/service/cmd/service/handler"

	"im-service/app/chat/service/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type wsRepo struct {
	data *Data
	log  *log.Helper
}

func (w wsRepo) BindMember(ctx context.Context, request *pb.BindMemberRequest) (*pb.BindMemberReply, error) {
	clientLen := len(handler.Manager.Clients)
	log.Info(fmt.Sprintf("现存客户端数量为%d", clientLen))
	//TODO implement me
	return nil, nil
}

func (w wsRepo) BindGroup(ctx context.Context, request *pb.BindGroupRequest) (*pb.BindGroupReply, error) {
	//TODO implement me
	return nil, nil
}

func (w wsRepo) CancelGroup(ctx context.Context, request *pb.CancelGroupRequest) (*pb.CancelGroupReply, error) {
	//TODO implement me
	return nil, nil
}

func (w wsRepo) SendMsg(ctx context.Context, request *pb.SendMsgRequest) (*pb.SendMsgReply, error) {
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
