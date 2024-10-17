package service

import (
	"context"
	"im-service/app/chat/service/internal/biz"

	pb "im-service/api/chat/service/v1"
)

type WsService struct {
	pb.UnimplementedWsServer

	uc *biz.WsUsecase
}

func NewWsService(uc *biz.WsUsecase) *WsService {
	return &WsService{uc: uc}
}

func (s *WsService) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{}, nil
}
