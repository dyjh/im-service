package consts

const (
	WsSuccess = 0
	WsError   = 1
	WsEnd     = 101
)

// 消息发送类型
const (
	SendRes    = "SendRes"    // 消息发送结果
	Heart      = "Ping"       // 心跳
	MsgReceive = "MsgReceive" // 消息接收
	SysMsg     = "SysMsg"     // 消息接收
	Refresh    = "Refresh"    // 刷新消息列表
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
