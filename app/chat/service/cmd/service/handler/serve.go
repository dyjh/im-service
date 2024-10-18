package handler

import (
	"encoding/json"
	"fmt"
	_ "fmt"
	"github.com/go-kratos/kratos/v2/log"
	"im-service/app/chat/service/cmd/service/consts"
	"im-service/app/chat/service/internal/conf"

	//"github.com/gin-gonic/gin/binding"

	"github.com/gorilla/websocket"
)

type InitMsg struct {
	Code     int    `json:"code"`
	Type     string `json:"type"`
	Content  string `json:"content"`
	ClientId string `json:"client_id"`
}

type ClientEvent struct {
	Uid   string `json:"uid"`
	Event string `json:"event"` // register 上线事件 un_register 离线时间
}

var wsConf *conf.Websocket

func (manager *ClientManager) WebSocketStart(c *conf.Websocket, logger log.Logger) {
	logHelper := log.NewHelper(logger)
	wsConf = c
	for {
		//global.GVA_LOG.Info("<---监听管道通信--->")
		select {
		case conn := <-Manager.Register: // 建立连接
			logHelper.Info("建立新连接:" + "ClientID - " + conn.Uid)
			Manager.Clients[conn.Uid] = conn
			replyMsg := &InitMsg{
				Code:     consts.WsSuccess,
				Type:     "init",
				Content:  "已连接至服务器",
				ClientId: conn.Uid,
			}
			msg, _ := json.Marshal(replyMsg)
			_ = conn.sendByte(websocket.TextMessage, msg)

		case conn := <-Manager.Unregister: // 断开连接
			logHelper.Info("注销客户端")
			fmt.Println("注销客户端")
			if conn.MemberId > 0 {
				if ClientId, ok := Manager.MapUserIdToClientId[conn.MemberId]; ok && ClientId == conn.Uid {
					delete(Manager.MapUserIdToClientId, conn.MemberId)
				}
			}

			if _, ok := Manager.Clients[conn.Uid]; ok {
				logHelper.Info("删除链接")
				close(conn.Send)
				delete(Manager.Clients, conn.Uid)
			}
		}
		//
	}
}
