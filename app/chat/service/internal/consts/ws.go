package consts

const (
	WsSuccess = 0
	WsError   = 1
	WsEnd     = 101
)

const (
	// websocket相关错误
	WebsocketOnOpenFailMsg      string = "websocket连接失败"
	WebsocketUpgradeFailMsg     string = "websocket协议升级失败"
	WebsocketSendMessageFailMsg string = "websocket发送消息失败"
	WebsocketReadFailMsg        string = "websocket Read 协程读取消息错误"
	WebsocketClientLogoutMsg    string = "websocket客户端已下线"
	WebsocketHeartFailMsg       string = "websocket心跳包检测协程错误"
)
