package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"

	pb "im-service/api/chat/service/v1"

	"github.com/go-kratos/kratos/v2/errors"
)

var (
	// ErrUserNotFound is user not found.
	ErrUserNotFound = errors.NotFound(pb.ErrorReason_USER_NOT_FOUND.String(), "user not found")
)

// WsUseCase is a Greeter useCase.
type WsUseCase struct {
	repo WsRepo
	log  *log.Helper
}

func NewWsUseCase(repo WsRepo, logger log.Logger) *WsUseCase {
	return &WsUseCase{repo: repo, log: log.NewHelper(log.With(logger, "module", "useCase/chat"))}
}

// WsRepo is a Greater repo.
type WsRepo interface {
	BindMember(context.Context, *pb.BindMemberRequest) (*pb.BindMemberReply, error)
	BindGroup(context.Context, *pb.BindGroupRequest) (*pb.BindGroupReply, error)
	CancelGroup(context.Context, *pb.CancelGroupRequest) (*pb.CancelGroupReply, error)
	SendMsg(context.Context, *pb.SendMsgRequest) (*pb.SendMsgReply, error)
	DelGroupColl(context.Context, *pb.DelGroupCollRequest) (*pb.DelGroupCollReply, error)
	GetGroupHistory(context.Context, *pb.GetGroupHistoryRequest) (*pb.GetGroupHistoryReply, error)
}

func (wc *WsUseCase) BindMember(ctx context.Context, req *pb.BindMemberRequest) (*pb.BindMemberReply, error) {
	return wc.repo.BindMember(ctx, req)
}

func (wc *WsUseCase) BindGroup(ctx context.Context, req *pb.BindGroupRequest) (*pb.BindGroupReply, error) {
	return wc.repo.BindGroup(ctx, req)
}

func (wc *WsUseCase) CancelGroup(ctx context.Context, req *pb.CancelGroupRequest) (*pb.CancelGroupReply, error) {
	return wc.repo.CancelGroup(ctx, req)
}

func (wc *WsUseCase) SendMsg(ctx context.Context, req *pb.SendMsgRequest) (*pb.SendMsgReply, error) {
	return wc.repo.SendMsg(ctx, req)
}

func (wc *WsUseCase) DelGroupColl(ctx context.Context, req *pb.DelGroupCollRequest) (*pb.DelGroupCollReply, error) {
	return wc.repo.DelGroupColl(ctx, req)
}

func (wc *WsUseCase) GetGroupHistory(ctx context.Context, req *pb.GetGroupHistoryRequest) (*pb.GetGroupHistoryReply, error) {
	return wc.repo.GetGroupHistory(ctx, req)
}
