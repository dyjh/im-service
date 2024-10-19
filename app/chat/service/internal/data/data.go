package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/apache/rocketmq-client-go/v2/rlog"
	"github.com/go-redis/redis/v8"
	"im-service/app/chat/service/cmd/service/handler"
	"im-service/app/chat/service/internal/conf"
	"im-service/app/chat/service/internal/consts"
	"im-service/app/chat/service/utils"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewWsRepo)

// Data .
type Data struct {
	// TODO wrapped database client
	log           *log.Helper
	p             rocketmq.Producer
	r             *redis.Client
	IP            string
	mqTag         string
	Port          string
	clientManager *handler.ClientManager
}

// NewRedisClient 初始化 Redis 客户端
func NewRedisClient(c *conf.Data) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     c.Redis.Addr,
		Password: c.Redis.Password,
		DB:       int(c.Redis.Db),
	})
}

func NewMqProducer(conf *conf.RocketMq, logLevel string) rocketmq.Producer {
	rlog.SetLogLevel(logLevel)
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

func NewMqConsumer(conf *conf.RocketMq, logLevel string, h *handler.Handler) (rocketmq.PushConsumer, string) {

	rlog.SetLogLevel(logLevel)

	c, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{conf.Addr}),              // 替换为您的 NameServer 地址
		consumer.WithGroupName(conf.GroupPrefix+"_chat_consumer"), // 替换为您的消费者组名
		consumer.WithConsumeFromWhere(consumer.ConsumeFromFirstOffset),
	)
	if err != nil {
		panic(err)
	}
	tag := fmt.Sprintf("%s_chat_context", conf.Addr)
	err = c.Subscribe("chatMessage", consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: tag, // 指定要订阅的 Tag
	}, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range msgs {
			var (
				ReceiveMsg utils.MqMsg
				msgBody    handler.ReplyMsg
			)
			err = json.Unmarshal(msg.Body, &ReceiveMsg)
			if err != nil {
				fmt.Printf("消息解析失败")
				continue
			}

			err = json.Unmarshal(ReceiveMsg.Body, &msgBody)
			if err != nil {
				fmt.Printf("消息解析失败")
				continue
			}

			switch ReceiveMsg.Type {
			case consts.MQ_MSG_TYEP_MEMBER:
				mId, err := strconv.Atoi(ReceiveMsg.SendTo)
				if err != nil {
					fmt.Printf("消息解析失败")
					continue
				}
				err = handler.SendMsgByMemberId(h, uint(mId), msgBody)
				if err != nil {
					fmt.Println(err)
					continue
				}
				break
			}

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
func NewData(r *redis.Client, p rocketmq.Producer, c rocketmq.PushConsumer, mqTag string, logger log.Logger, confData *conf.Server, clientManager *handler.ClientManager) (*Data, func(), error) {
	logHelper := log.NewHelper(logger)
	cleanup := func() {
		logHelper.Info("closing the data resources")
		_ = p.Shutdown()
		_ = c.Shutdown()
		_ = r.Close()
		//_ = r.Shutdown(context.Background())
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
		log:           logHelper,
		p:             p,
		r:             r,
		IP:            IP,
		mqTag:         mqTag,
		Port:          port,
		clientManager: clientManager,
	}, cleanup, nil
}
