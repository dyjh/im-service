package data

import (
	"context"
	pb "im-service/api/chat/service/v1"

	"im-service/app/chat/service/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type wsRepo struct {
	data *Data
	log  *log.Helper
}

func (w wsRepo) SayHello(ctx context.Context, request *pb.HelloRequest) (*pb.HelloReply, error) {
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
