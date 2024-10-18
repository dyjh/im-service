package utils

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
)

type WSClientInfo struct {
	IP    string `json:"ip"`
	Port  string `json:"port"`
	MQTag string `json:"mq_tag"`
}

func GetUsersInGroup(ctx context.Context, rdb *redis.Client, groupId string) (map[string]WSClientInfo, error) {
	key := "group:" + groupId + ":users"
	result, err := rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	users := make(map[string]WSClientInfo)
	for userid, data := range result {
		var info WSClientInfo
		if err := json.Unmarshal([]byte(data), &info); err != nil {
			return nil, err
		}
		users[userid] = info
	}
	return users, nil
}

func AddUserToGroup(ctx context.Context, rdb *redis.Client, groupId string, userid string, info WSClientInfo) error {
	key := "group:" + groupId + ":users"
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return rdb.HSet(ctx, key, userid, data).Err()
}

func RemoveUserFromGroup(ctx context.Context, rdb *redis.Client, groupId string, userid string) error {
	key := "group:" + groupId + ":users"

	// 开始事务
	txf := func(tx *redis.Tx) error {
		// 事务函数
		// 删除用户字段
		if err := tx.HDel(ctx, key, userid).Err(); err != nil {
			return err
		}

		// 检查哈希表长度
		length, err := tx.HLen(ctx, key).Result()
		if err != nil {
			return err
		}

		// 如果长度为 0，删除整个键
		if length == 0 {
			if err := tx.Del(ctx, key).Err(); err != nil {
				return err
			}
		}

		return nil
	}

	// 执行事务
	for {
		err := rdb.Watch(ctx, txf, key)
		if err == nil {
			// 成功
			return nil
		}
		if err == redis.TxFailedErr {
			// 重试
			continue
		}
		// 其他错误
		return err
	}
}
