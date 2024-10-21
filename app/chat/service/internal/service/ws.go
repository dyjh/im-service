package service

import (
	"context"

	pb "im-service/api/chat/service/v1"
)

func (ws *WsService) BindMember(ctx context.Context, req *pb.BindMemberRequest) (*pb.BindMemberReply, error) {
	return ws.wc.BindMember(ctx, req)
}
func (ws *WsService) BindGroup(ctx context.Context, req *pb.BindGroupRequest) (*pb.BindGroupReply, error) {
	return ws.wc.BindGroup(ctx, req)
}
func (ws *WsService) CancelGroup(ctx context.Context, req *pb.CancelGroupRequest) (*pb.CancelGroupReply, error) {
	return ws.wc.CancelGroup(ctx, req)
}
func (ws *WsService) SendMsg(ctx context.Context, req *pb.SendMsgRequest) (*pb.SendMsgReply, error) {
	return ws.wc.SendMsg(ctx, req)
}

func (ws *WsService) DelGroupColl(ctx context.Context, req *pb.DelGroupCollRequest) (*pb.DelGroupCollReply, error) {
	return ws.wc.DelGroupColl(ctx, req)
}

func (ws *WsService) GetGroupHistory(ctx context.Context, req *pb.GetGroupHistoryRequest) (*pb.GetGroupHistoryReply, error) {
	return ws.wc.GetGroupHistory(ctx, req)
}
