package service

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	pb "im-service/api/chat/service/v1"
	"im-service/app/chat/service/internal/biz"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewWsService)

type WsService struct {
	pb.UnimplementedWsServer

	wc  *biz.WsUseCase
	log *log.Helper
}

func NewWsService(wc *biz.WsUseCase, logger log.Logger) *WsService {
	return &WsService{
		wc:  wc,
		log: log.NewHelper(log.With(logger, "module", "service/ws")),
	}
}
