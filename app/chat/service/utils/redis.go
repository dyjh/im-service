package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
)

type WSClientInfo struct {
	IP    string `json:"ip"`
	Port  string `json:"port"`
	MQTag string `json:"mq_tag"`
}

func AddOrUpdateUserInfo(ctx context.Context, rdb *redis.Client, userId string, info WSClientInfo) error {
	userKey := "user:" + userId + ":info"

	// 序列化用户信息为 JSON 字符串
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	// 将用户信息保存到 Redis
	err = rdb.Set(ctx, userKey, data, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func DeleteUserInfo(ctx context.Context, rdb *redis.Client, userId string) error {
	userKey := "user:" + userId + ":info"

	// 删除用户信息键
	err := rdb.Del(ctx, userKey).Err()
	if err != nil {
		return err
	}

	return nil
}

func GetUserInfo(ctx context.Context, rdb *redis.Client, userId string) (WSClientInfo, error) {
	userKey := "user:" + userId + ":info"

	// 从 Redis 获取用户信息
	data, err := rdb.Get(ctx, userKey).Result()
	if err != nil {
		if err == redis.Nil {
			// 键不存在，返回自定义的错误或空值
			return WSClientInfo{}, fmt.Errorf("user %s not found", userId)
		}
		return WSClientInfo{}, err
	}

	var info WSClientInfo
	// 反序列化 JSON 字符串到结构体
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		return WSClientInfo{}, err
	}

	return info, nil
}

func GetUsersInGroup(ctx context.Context, rdb *redis.Client, groupId string) (map[string]WSClientInfo, error) {
	groupKey := "group:" + groupId + ":users"

	// 获取组内用户ID列表
	userIds, err := rdb.SMembers(ctx, groupKey).Result()
	if err != nil {
		return nil, err
	}

	// 使用 Pipeline 批量获取用户信息
	pipe := rdb.Pipeline()
	defer pipe.Close()

	cmds := make([]*redis.StringCmd, len(userIds))
	for i, userId := range userIds {
		userKey := "user:" + userId + ":info"
		cmds[i] = pipe.Get(ctx, userKey)
	}

	// 执行 Pipeline
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, err
	}

	// 解析结果
	users := make(map[string]WSClientInfo)
	for i, cmd := range cmds {
		data, err := cmd.Result()
		if err != nil && err != redis.Nil {
			return nil, err
		}
		var info WSClientInfo
		if err := json.Unmarshal([]byte(data), &info); err != nil {
			return nil, err
		}
		users[userIds[i]] = info
	}

	return users, nil
}

func BindUserToGroup(ctx context.Context, rdb *redis.Client, userId string, groupId string) error {
	groupKey := "group:" + groupId + ":users"

	// 将用户ID添加到组的集合中
	err := rdb.SAdd(ctx, groupKey, userId).Err()
	if err != nil {
		return err
	}

	return nil
}

func UnbindUserFromGroup(ctx context.Context, rdb *redis.Client, userId string, groupId string) error {
	groupKey := "group:" + groupId + ":users"

	// 从组的集合中移除用户ID
	err := rdb.SRem(ctx, groupKey, userId).Err()
	if err != nil {
		return err
	}

	// 检查组是否为空，若为空则删除键
	size, err := rdb.SCard(ctx, groupKey).Result()
	if err != nil {
		return err
	}
	if size == 0 {
		if err := rdb.Del(ctx, groupKey).Err(); err != nil {
			return err
		}
	}

	return nil
}
