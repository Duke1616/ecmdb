package ioc

import (
	"fmt"
	"net"

	"github.com/spf13/viper"
)

// InitListener 初始化 HTTP 监听器
func InitListener() net.Listener {
	port := viper.GetInt("web.port")
	if port == 0 {
		port = 8000
	}
	host := viper.GetString("web.host")

	addr := fmt.Sprintf("%s:%d", host, port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	return l
}
