package handler

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
	"im-service/app/chat/service/internal/consts"
	"im-service/app/chat/service/internal/model"
	"im-service/app/chat/service/utils"
	"net/http"
	"strconv"
	"sync"
	"time"
	//"strconv"
	//"time"
)

type SendMsg struct {
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
}

type ChatContent struct {
	MessageId   string    `json:"message_id"`
	MemberId    uint64    `json:"member_id"`
	ContentType string    `json:"content_type"`
	Content     string    `json:"content"`
	CreateAt    time.Time `json:"create_at"`
}

type ReplyMsg struct {
	Code    int         `json:"code"`
	Type    string      `json:"type"` // 消息回复类型 ping 心跳
	Content interface{} `json:"content"`
	ErrMsg  string      `json:"err_msg"`
}

type Client struct {
	Uid                string
	MemberId           uint
	GroupId            string
	GroupNo            string
	Socket             *websocket.Conn `json:"-"`
	state              int
	Send               chan ReplyMsg `json:"-"`
	heartbeatFailTimes int
	ticker             *time.Ticker
	ReadDeadline       time.Duration `json:"-"`
	WriteDeadline      time.Duration `json:"-"`
	sync.RWMutex
}

type ClientManager struct {
	Clients             map[string]*Client
	MapUserIdToClientId map[uint]string
	Mutex               sync.Mutex
	Reply               chan *Client
	Register            chan *Client
	Unregister          chan *Client
	RedisClient         *redis.Client
	MongoDb             *mongo.Database
	Mysql               *gorm.DB
}

type Handler struct {
	log           *log.Helper
	producer      rocketmq.Producer
	ClientManager *ClientManager
}

type MessageHistory struct {
	MemberId    uint64    `json:"member_id"`
	Content     string    `json:"content"`
	ContentType string    `json:"content_type"`
	SendTime    time.Time `json:"send_time"`
}

func NewHandler(producer rocketmq.Producer, logger log.Logger, manager *ClientManager) *Handler {
	return &Handler{
		log:           log.NewHelper(logger),
		producer:      producer,
		ClientManager: manager,
	}
}

func (h *Handler) WsHandler(ctx *gin.Context) {

	// 升级成ws协议
	conn, err := (&websocket.Upgrader{
		ReadBufferSize:  int(wsConf.WriteReadBufferSize),
		WriteBufferSize: int(wsConf.WriteReadBufferSize),
		CheckOrigin: func(r *http.Request) bool { // CheckOrigin解决跨域问题
			return true
		}}).
		Upgrade(ctx.Writer, ctx.Request, nil)

	if err != nil {
		h.log.Error(consts.WebsocketUpgradeFailMsg)
		return
	}

	// 创建一个用户客户端会话实例
	newClient := &Client{
		Uid:           getClientUid(),
		Socket:        conn,
		Send:          make(chan ReplyMsg),
		state:         1,
		ReadDeadline:  time.Duration(wsConf.ReadDeadline) * time.Second,
		WriteDeadline: time.Duration(wsConf.WriteDeadline) * time.Second,
	}
	// 用户会话注册到用户管理上
	h.ClientManager.Register <- newClient

	// ---------------------------------------- //
	go newClient.read(ctx, h)
	go newClient.write(h)

	// 启动心跳服务
	newClient.HeartBeat()
}

func getClientUid() string {
	// 获取当前时间戳（毫秒）
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	// 将时间戳转换为字符串
	timestampStr := strconv.FormatInt(timestamp, 10)

	// 计算字符串的MD5哈希值
	hash := md5.Sum([]byte(timestampStr))

	// 将哈希值转换为十六进制字符串
	return hex.EncodeToString(hash[:])
}

// 从websocket读取客户端用户的消息，然后服务器回应前端一个消息
func (c *Client) read(ctx *gin.Context, h *Handler) {
	defer func() { // 避免忘记关闭，所以要加上close
		err := recover()
		if err != nil {
			if val, ok := err.(error); ok {
				h.log.Error("程序异常：" + val.Error())
			}
		}
		_ = c.Socket.Close()
	}()

	for {
		// 使用 ReadMessage() 读取原始消息字节
		_, rawMsg, err := c.Socket.ReadMessage()
		if err != nil {
			// 处理关闭错误和其他类型的错误
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				h.log.Info("客户端已断开连接:", c.Uid)
				return
			}
			h.log.Error("读取消息错误:", err)
			return
		}
		if len(rawMsg) == 0 {
			continue
		}
		h.log.Debug("收到的原始请求: " + string(rawMsg))

		sendMsg := new(SendMsg)
		err = json.Unmarshal(rawMsg, &sendMsg)
		if err != nil {
			sendErr(h, c, "数据格式不正确", false)
			continue
		}

		if sendMsg.ContentType == "pong" {
			continue
		}

		if sendMsg.ContentType != "text" {
			sendErr(h, c, "消息格式错误", false)
			continue
		}

		if sendMsg.Content == "" {
			sendErr(h, c, "不能发送空消息", false)
			continue
		}

		if c.MemberId == 0 {
			sendErr(h, c, "请先绑定用户", false)
			continue
		}

		if c.GroupId == "" {
			sendErr(h, c, "未指定消息发送对象", false)
			continue
		}

		// 获取唯一消息编号
		var MessageNo = utils.GetMessageNo()

		MsgContent := ChatContent{
			MessageId:   MessageNo,
			MemberId:    uint64(c.MemberId),
			ContentType: sendMsg.ContentType,
			Content:     sendMsg.Content,
			CreateAt:    time.Now(),
		}

		ContentBytes, _ := json.Marshal(MsgContent)

		// 根据GroupId获取组成员
		GroupClients, err := utils.GetUsersInGroup(ctx, h.ClientManager.RedisClient, c.GroupId)
		if err != nil {
			h.log.Errorf("聊天组数据错误:%s", err.Error())
			sendErr(h, c, "聊天组数据错误，请联系管理员", false)
			return
		}

		Ip, _ := utils.GetLocalIP()

		for ClientMemberId, clientInfo := range GroupClients {
			mId, _ := strconv.Atoi(ClientMemberId)
			if mId == int(c.MemberId) {
				continue
			}
			if Ip == clientInfo.IP {

				clientId := h.ClientManager.MapUserIdToClientId[uint(mId)]
				if clientId != "" {
					client := h.ClientManager.Clients[clientId]
					replyMsg := ReplyMsg{
						Code:    consts.WsSuccess,
						Type:    consts.MsgReceive,
						Content: MsgContent,
					}
					//err := client.sendByte(websocket.TextMessage, ContentBytes)
					client.Send <- replyMsg
					if err != nil {
						h.log.Errorf("消息发送失败:%s", err.Error())
						sendErr(h, c, "消息发送失败，请联系管理员", false)
						return
					}
				}
			} else {
				MqMsg := utils.MqMsg{
					Type:   consts.MQ_MSG_TYEP_MEMBER,
					SendTo: ClientMemberId,
					Body:   ContentBytes,
				}
				JsonData, _ := json.Marshal(MqMsg)

				err = utils.SendMqMsg(h.producer, "chatMessage", clientInfo.MQTag, JsonData)
				if err != nil {
					h.log.Errorf("消息发送失败:%s", err.Error())
					sendErr(h, c, "消息发送失败，请联系管理员", false)
					return
				}
			}

		}

		replyMsg := ReplyMsg{
			Code:    consts.WsSuccess,
			Type:    consts.SendRes,
			Content: "success",
		}
		msg, _ := json.Marshal(replyMsg)
		// 回复数据至前端用户
		_ = c.sendByte(websocket.TextMessage, msg)

		// TODO 对Group内未在线用户发送消息记录
		_ = c.AddMessageToGroup(h, ctx, c.GroupId, MsgContent)
	}
}

// 消息发出
func (c *Client) write(h *Handler) {
	defer func() {
		err := recover()
		if err != nil {
			if val, ok := err.(error); ok {
				h.log.Errorw(consts.WebsocketSendMessageFailMsg, val)
			}
		}
		c.ticker.Stop()
		_ = c.Socket.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				_ = c.sendByte(websocket.CloseMessage, []byte{})
				return
			}
			h.log.Infow("接收消息：", "memberId", c.MemberId, "message", message)

			msg, _ := json.Marshal(message)
			_ = c.sendByte(websocket.TextMessage, msg)
		case <-c.ticker.C:
			// 心跳计时器
			if c.state == 1 {
				heartMsg, _ := json.Marshal(ReplyMsg{
					Code:    consts.WsSuccess,
					Type:    consts.Heart,
					Content: fmt.Sprintf("%d", c.MemberId),
				})
				err := c.sendByte(websocket.TextMessage, heartMsg)

				if err != nil {
					c.heartbeatFailTimes++
					if c.heartbeatFailTimes > int(wsConf.HeartbeatFailMaxTimes) {
						c.state = 0
						h.ClientManager.Unregister <- c
						return
					}
				} else {

					// ping通则清空失败次数
					c.heartbeatFailTimes = 0
				}
			} else {
				return
			}
		}
	}
}

// 错误信息返回
func sendErr(h *Handler, c *Client, msg string, isClose bool) {
	// 向前端返回数据格式不正确的状态码和消息
	errMsg, _ := json.Marshal(ReplyMsg{
		Code:   consts.WsError,
		Type:   consts.SendRes,
		ErrMsg: msg,
	})
	_ = c.sendByte(websocket.TextMessage, errMsg)

	if isClose {
		h.ClientManager.Unregister <- c
		_ = c.Socket.Close()
	}
}

// 发送消息基础方法
func (c *Client) sendByte(messageType int, message []byte) error {
	c.Lock()
	defer func() {
		c.Unlock()
	}()
	if err := c.Socket.SetReadDeadline(time.Now().Add(c.WriteDeadline)); err != nil {
		return err
	}

	if err := c.Socket.WriteMessage(messageType, message); err != nil {
		return err
	}
	return nil
}

type MongoMsgData struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"` // 添加 _id 字段
	MemberId    uint64             `json:"memberId" bson:"memberId"`
	GroupId     string             `json:"group_id" bson:"group_id"`
	MessageBody ChatContent        `json:"message_body" bson:"message_body"`
	CreateAt    time.Time          `json:"create_at" bson:"create_at"`
}

// AddMessageToGroup 向指定 group_id 的所有成员添加一条新的消息
func (c *Client) AddMessageToGroup(h *Handler, ctx *gin.Context, groupID string, messageBody ChatContent) error {
	var GroupMembers []*model.GroupBind
	res := h.ClientManager.Mysql.Model(model.GroupBind{}).Where("group_id = ?", groupID).
		Find(&GroupMembers)
	coll := h.ClientManager.MongoDb.Collection("group_messages")
	if res.RowsAffected > 0 {
		for _, v := range GroupMembers {
			messageSave := MongoMsgData{
				MemberId:    v.MemberId,
				GroupId:     v.GroupId,
				MessageBody: messageBody,
				CreateAt:    messageBody.CreateAt,
			}
			// 插入消息文档
			_, err := coll.InsertOne(ctx, messageSave)
			if err != nil {
				h.log.Errorf("InsertOne 操作失败: %v", err)
				return errors.New(consts.HTTP_ERROR, fmt.Sprintf("%w", err), "消息添加失败")
			}
			// 记录日志
			h.log.Infof("已插入消息: %v\n", messageSave)

		}
	}
	return nil
}

// HeartBeat 心跳包处理方法
func (c *Client) HeartBeat() {
	c.ticker = time.NewTicker(time.Duration(wsConf.PingPeriod) * time.Second)
	if c.ReadDeadline == 0 {
		_ = c.Socket.SetReadDeadline(time.Time{})
	} else {
		_ = c.Socket.SetReadDeadline(time.Now().Add(c.ReadDeadline))
	}
	c.Socket.SetPongHandler(func(receivedPong string) error {
		if c.ReadDeadline > time.Nanosecond {
			_ = c.Socket.SetReadDeadline(time.Now().Add(c.ReadDeadline))
		} else {
			_ = c.Socket.SetReadDeadline(time.Time{})
		}
		return nil
	})

}
