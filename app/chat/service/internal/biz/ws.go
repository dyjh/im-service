package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"

	v1 "im-service/api/chat/service/v1"

	"github.com/go-kratos/kratos/v2/errors"
)

var (
	// ErrUserNotFound is user not found.
	ErrUserNotFound = errors.NotFound(v1.ErrorReason_USER_NOT_FOUND.String(), "user not found")
)

// WsUsecase is a Greeter usecase.
type WsUsecase struct {
	repo WsRepo
	log  *log.Helper
}

func NewWsUseCase(repo WsRepo, logger log.Logger) *WsUsecase {
	return &WsUsecase{repo: repo, log: log.NewHelper(log.With(logger, "module", "usecase/cart"))}
}

// GreeterRepo is a Greater repo.
type WsRepo interface {
	SayHello(context.Context, *v1.HelloRequest) (*v1.HelloReply, error)
}
