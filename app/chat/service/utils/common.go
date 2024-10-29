package utils

import (
	"fmt"
	"github.com/go-kratos/kratos/v2/errors"
	"im-service/app/chat/service/internal/consts"
	"net"
	"os"
)

// GetLocalIP 获取服务器的本地IP地址
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no valid IP address found")
}

func PathExists(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err == nil {
		if fi.IsDir() {
			return true, nil
		}
		return false, errors.New(consts.HTTP_ERROR, "存在同名文件", "存在同名文件")
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
