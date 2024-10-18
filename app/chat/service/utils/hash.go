package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

func GetMessageNo() string {
	// 获取当前时间戳（毫秒）
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	// 将时间戳转换为字符串
	timestampStr := strconv.FormatInt(timestamp, 10)

	// 计算字符串的MD5哈希值
	hash := md5.Sum([]byte(timestampStr))

	// 将哈希值转换为十六进制字符串
	return fmt.Sprintf("MSN%s", hex.EncodeToString(hash[:]))
}
