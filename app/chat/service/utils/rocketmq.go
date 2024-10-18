package utils

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

func SendMqMsg(p rocketmq.Producer, Topic string, Tag string, sendData []byte) error {
	msg := &primitive.Message{
		Topic: Topic, // 替换为您的 Topic
		Body:  sendData,
	}
	msg.WithTag(Tag) // 设置指定的 Tag

	res, err := p.SendSync(context.Background(), msg)
	if err != nil {
		fmt.Printf("发送消息失败: %s\n", err)
		return err
	}
	fmt.Printf("发送消息成功: result=%s\n", res.String())
	return nil
}
