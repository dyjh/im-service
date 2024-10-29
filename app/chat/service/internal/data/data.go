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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"im-service/app/chat/service/internal/conf"
	"im-service/app/chat/service/internal/consts"
	"im-service/app/chat/service/internal/handler"
	"im-service/app/chat/service/internal/model"
	"im-service/app/chat/service/utils"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewWsRepo, NewMongo, NewMysql, NewMqProducer, NewMqConsumer)

// Data .
type Data struct {
	// TODO wrapped database client
	mongo         *mongo.Database
	mysqlClient   *gorm.DB
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

func NewMongo(conf *conf.Data) *mongo.Database {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(conf.Mongodb.Uri))
	if err != nil {
		panic(err)
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		panic(err)
	}
	return client.Database(conf.Mongodb.Database)
}

func NewMysql(conf *conf.Data, logger log.Logger) *gorm.DB {
	log := log.NewHelper(log.With(logger, "module", "order-service/data/gorm"))

	db, err := gorm.Open(mysql.Open(conf.Mysql.Source), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed opening connection to mysql: %v", err)
	}

	if err := db.AutoMigrate(&model.GroupBind{}); err != nil {
		log.Fatal(err)
	}
	return db
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

func NewMqConsumer(conf *conf.RocketMq, logService log.Logger, logLevel string, h *handler.Handler) (rocketmq.PushConsumer, string) {

	rlog.SetLogLevel(logLevel)
	logger := log.NewHelper(logService)
	c, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{conf.Addr}),              // 替换为您的 NameServer 地址
		consumer.WithGroupName(conf.GroupPrefix+"_chat_consumer"), // 替换为您的消费者组名
		consumer.WithConsumeFromWhere(consumer.ConsumeFromFirstOffset),
		consumer.WithMaxReconsumeTimes(5), // 设置最大重试次数
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
			// 解析外层消息
			err = json.Unmarshal(msg.Body, &ReceiveMsg)
			if err != nil {
				logger.Errorf("外层消息解析失败%v", err)
				// 返回重试结果
				continue
			}

			// 解析内层消息
			err = json.Unmarshal(ReceiveMsg.Body, &msgBody)
			if err != nil {
				logger.Errorf("内层消息解析失败%v", err)
				// 返回重试结果
				continue
			}

			switch ReceiveMsg.Type {
			case consts.MQ_MSG_TYEP_MEMBER:
				mId, err := strconv.Atoi(ReceiveMsg.SendTo)
				if err != nil {
					logger.Errorf("发送人格式转化失败%v", err)
					// 返回重试结果
					return consumer.ConsumeRetryLater, err
				}

				clientId, isExist := h.ClientManager.MapUserIdToClientId[uint(mId)]
				if isExist {
					sendMsg := handler.ReplyMsg{
						Code:    consts.WsSuccess,
						Type:    consts.MsgReceive,
						Content: msgBody,
					}
					client := h.ClientManager.Clients[clientId]
					client.Send <- sendMsg
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
func NewData(database *mongo.Database, mysql *gorm.DB, r *redis.Client, p rocketmq.Producer, c rocketmq.PushConsumer, mqTag string, logger log.Logger, confData *conf.Server, clientManager *handler.ClientManager) (*Data, func(), error) {
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
		mongo:         database,
		mysqlClient:   mysql,
		log:           logHelper,
		p:             p,
		r:             r,
		IP:            IP,
		mqTag:         mqTag,
		Port:          port,
		clientManager: clientManager,
	}, cleanup, nil
}
