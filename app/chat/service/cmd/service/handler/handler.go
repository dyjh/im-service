package handler

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"im-service/app/chat/service/cmd/service/consts"
	"im-service/app/chat/service/utils"
	"net/http"
	"strconv"
	"sync"
	"time"
	//"strconv"
	//"time"
)

type SendMsg struct {
	//ChatType    string `json:"chat_type"`
	//To          uint   `json:"to"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
}

type SenderInfo struct {
	Avatar   string `json:"avatar"`
	Nickname string `json:"nickname"`
}

type ChatContent struct {
	MessageId   string     `json:"message_id"`
	ContentType string     `json:"content_type"`
	Content     string     `json:"content"`
	CreateAt    time.Time  `json:"create_at"`
	SenderInfo  SenderInfo `json:"sender_info"`
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
	State              int
	Send               chan ChatContent `json:"-"`
	HeartbeatFailTimes int              `json:"-"`
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
}

var Manager = ClientManager{
	Clients:             make(map[string]*Client), // 参与连接的用户，出于性能的考虑，需要设置最大连接数
	MapUserIdToClientId: make(map[uint]string),
	Register:            make(chan *Client),
	Reply:               make(chan *Client),
	Unregister:          make(chan *Client),
}

// 消息发送类型
const (
	SendRes      = "SendRes"      // 消息发送结果
	ChatResponse = "ChatResponse" // 消息发送结果
	Heart        = "Ping"         // 心跳
	MsgReceive   = "MsgReceive"   // 消息接收
	Refresh      = "Refresh"      // 刷新消息列表
)

type Handler struct {
	redisClient *redis.Client
	log         *log.Helper
	producer    rocketmq.Producer
}

func NewHandler(redisClient *redis.Client, producer rocketmq.Producer, logger log.Logger) *Handler {
	return &Handler{
		redisClient: redisClient,
		log:         log.NewHelper(logger),
		producer:    producer,
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
		Send:          make(chan ChatContent),
		State:         1,
		ReadDeadline:  time.Duration(wsConf.ReadDeadline) * time.Second,
		WriteDeadline: time.Duration(wsConf.WriteDeadline) * time.Second,
	}
	// 用户会话注册到用户管理上
	Manager.Register <- newClient

	// ---------------------------------------- //
	go newClient.read(ctx, h)
	go newClient.write()

	// 启动心跳服务
	newClient.HeartBeat()
}

func SendMsgByUserId(userId uint, msg []byte) error {
	if ClientId, ok := Manager.MapUserIdToClientId[userId]; ok {
		UserClient := Manager.Clients[ClientId]
		err := UserClient.sendByte(websocket.TextMessage, msg)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("客户端不存在")
	}

	return nil
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
				//global.GVA_LOG.Error(consts.WebsocketReadFailMsg, zap.Error(val))
				fmt.Println("程序异常：" + val.Error())
			}
		}
		_ = c.Socket.Close()
	}()

	for {

		// 使用 ReadMessage() 读取原始消息字节
		_, rawMsg, err := c.Socket.ReadMessage()
		if string(rawMsg) == "" {
			continue
		}
		// 将原始消息打印为字符串
		//global.GVA_LOG.Info("收到的原始请求: " + string(rawMsg))

		sendMsg := new(SendMsg)
		err = c.Socket.ReadJSON(&sendMsg)
		isClose := false

		if err != nil {
			//global.GVA_LOG.Error("格式错误", zap.Error(err))
			if err.Error() == "websocket: close 1005 (no status)" {
				//global.GVA_LOG.Info(consts.WebsocketClientLogoutMsg, zap.String("ClientID", c.Uid))
				isClose = true
			}
			sendErr(c, "数据格式不正确", isClose)
			if isClose {
				return
			} else {
				continue
			}
		}

		if sendMsg.ContentType == "pong" {
			continue
		}

		if sendMsg.ContentType != "text" {
			sendErr(c, "消息格式错误", false)
			continue
		}

		if sendMsg.Content == "" {
			sendErr(c, "不能发送空消息", false)
			continue
		}

		if c.MemberId == 0 {
			sendErr(c, "请先绑定用户", false)
			continue
		}

		if c.GroupId == "" {
			sendErr(c, "未指定消息发送对象", false)
			continue
		}

		// 获取唯一消息编号
		var MessageNo = utils.GetMessageNo()

		MsgContent := ChatContent{
			MessageId:   MessageNo,
			ContentType: sendMsg.ContentType,
			Content:     sendMsg.Content,
			CreateAt:    time.Now(),
		}

		ContentBytes, _ := json.Marshal(sendMsg.Content)

		// 消息发送id
		var SendIds = make([]uint64, 10, 50)
		for _, client := range Manager.Clients {
			if client.GroupId == c.GroupId && client.MemberId != c.MemberId {
				client.Send <- MsgContent
				SendIds = append(SendIds, uint64(client.MemberId))
			}
		}

		// 根据GroupId获取组成员
		GroupClients, err := utils.GetUsersInGroup(ctx, h.redisClient, c.GroupId)
		if err != nil {
			sendErr(c, "聊天组数据错误，请联系管理员", isClose)
			return
		}
		GroupOnlineMemberIds := make([]int, 0, 4)
		for ClientMemberId, clientInfo := range GroupClients {
			mId, _ := strconv.Atoi(ClientMemberId)
			GroupOnlineMemberIds = append(GroupOnlineMemberIds, mId)
			err = utils.SendMqMsg(h.producer, "chatMessage", clientInfo.MQTag, ContentBytes)
			if err != nil {
				sendErr(c, "消息发送失败，请联系管理员", isClose)
				return
			}
		}

		replyMsg := ReplyMsg{
			Code:    consts.WsSuccess,
			Type:    SendRes,
			Content: "success",
		}
		msg, _ := json.Marshal(replyMsg)
		// 回复数据至前端用户
		_ = c.sendByte(websocket.TextMessage, msg)

		// TODO 对Group内未在线用户发送消息记录
	}
}

// 消息发出
func (c *Client) write() {
	defer func() {
		err := recover()
		if err != nil {
			if val, ok := err.(error); ok {
				//global.GVA_LOG.Error(consts.WebsocketSendMessageFailMsg, zap.Error(val))
				fmt.Println("程序异常" + val.Error())
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
			//global.GVA_LOG.Info("接收消息：", zap.Uint("memberId", c.MemberId), zap.Any("message", message))
			replyMsg := ReplyMsg{
				Code:    consts.WsSuccess,
				Type:    MsgReceive,
				Content: message,
			}

			msg, _ := json.Marshal(replyMsg)
			_ = c.sendByte(websocket.TextMessage, msg)
		case <-c.ticker.C:
			// 心跳计时器
			if c.State == 1 {
				heartMsg, _ := json.Marshal(ReplyMsg{
					Code:    consts.WsSuccess,
					Type:    Heart,
					Content: fmt.Sprintf("%d", c.MemberId),
				})
				err := c.sendByte(websocket.TextMessage, heartMsg)

				if err != nil {
					fmt.Println(c.Uid)
					fmt.Println("心跳包发送失败：" + err.Error())

					data1, _ := json.Marshal(Manager.Clients)
					fmt.Println("客户端数据打印：" + string(data1))

					data2, _ := json.Marshal(Manager.MapUserIdToClientId)
					fmt.Println("用户id映射打印：" + string(data2))

					c.HeartbeatFailTimes++
					if c.HeartbeatFailTimes > int(wsConf.HeartbeatFailMaxTimes) {
						c.State = 0
						Manager.Unregister <- c
						return
					}
				} else {

					// ping通则清空失败次数
					c.HeartbeatFailTimes = 0
				}
			} else {
				return
			}
		}
	}
}

// 错误信息返回
func sendErr(c *Client, msg string, isClose bool) {
	// 向前端返回数据格式不正确的状态码和消息
	errMsg, _ := json.Marshal(ReplyMsg{
		Code:   consts.WsError,
		Type:   SendRes,
		ErrMsg: msg,
	})
	_ = c.sendByte(websocket.TextMessage, errMsg)

	if isClose {
		Manager.Unregister <- c
		_ = c.Socket.Close()
	}
}

// SendMessage 发送字符串消息
func (c *Client) SendMessage(messageType int, message string) error {
	return c.sendByte(messageType, []byte(message))
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
