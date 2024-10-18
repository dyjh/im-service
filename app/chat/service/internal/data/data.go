package data

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/go-redis/redis/v8"
	"im-service/app/chat/service/internal/conf"
	"im-service/app/chat/service/utils"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewWsRepo)

// Data .
type Data struct {
	// TODO wrapped database client
	log   *log.Helper
	p     rocketmq.Producer
	r     *redis.Client
	IP    string
	mqTag string
	Port  string
}

// NewRedisClient 初始化 Redis 客户端
func NewRedisClient(c *conf.Data) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     c.Redis.Addr,
		Password: c.Redis.Password,
		DB:       int(c.Redis.Db),
	})
}

func NewMqProducer(conf *conf.RocketMq) rocketmq.Producer {
	p, err := rocketmq.NewProducer(
		producer.WithNameServer([]string{conf.Addr}),              // 替换为您的 NameServer 地址
		producer.WithGroupName(conf.GroupPrefix+"_chat_producer"), // 替换为您的生产者组名
	)
	if err != nil {
		panic(err)
	}
	err = p.Start()
	if err != nil {
		panic(err)
	}
	return p
}

func NewMqConsumer(conf *conf.RocketMq) (rocketmq.PushConsumer, string) {
	c, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{conf.Addr}),              // 替换为您的 NameServer 地址
		consumer.WithGroupName(conf.GroupPrefix+"_chat_consumer"), // 替换为您的消费者组名
	)
	if err != nil {
		panic(err)
	}
	tag := fmt.Sprintf("%s_chat_context", conf.Addr)
	// 订阅带有指定 Tag 的消息
	err = c.Subscribe("chatMessage", consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: tag, // 指定要订阅的 Tag
	}, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range msgs {
			fmt.Printf("接收到消息: %s\n", string(msg.Body))
		}
		return consumer.ConsumeSuccess, nil
	})
	if err != nil {
		panic(err)
	}

	err = c.Start()
	if err != nil {
		panic(err)
	}
	return c, tag
}

// NewData .
func NewData(r *redis.Client, p rocketmq.Producer, c rocketmq.PushConsumer, mqTag string, logger log.Logger, confData *conf.Server) (*Data, func(), error) {
	logHelper := log.NewHelper(logger)
	cleanup := func() {
		logHelper.Info("closing the data resources")
		_ = p.Shutdown()
		_ = c.Shutdown()
		_ = r.Shutdown(context.Background())
	}
	IP, _ := utils.GetLocalIP()

	parts := strings.Split(confData.Http.Addr, ":")
	// 检查是否有足够的部分（确保分割后的长度大于1）
	var port string
	if len(parts) > 1 && parts[1] != "" {
		port := parts[1]
		fmt.Println("Port:", port)
	} else {
		panic("端口信息错误")
	}
	return &Data{
		log:   logHelper,
		p:     p,
		r:     r,
		IP:    IP,
		mqTag: mqTag,
		Port:  port,
	}, cleanup, nil
}
