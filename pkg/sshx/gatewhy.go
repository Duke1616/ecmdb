package sshx

import (
	"sort"
	"strconv"
)

type GatewayConfig struct {
	AuthType   string
	Host       string
	Port       int
	Username   string
	Password   string
	PrivateKey string
	Passphrase string
	Sort       int
}

type MultiGatewayManager struct {
	Gateways []*GatewayConfig
}

func NewMultiGatewayManager(config []*GatewayConfig) *MultiGatewayManager {
	sort.Slice(config, func(i, j int) bool {
		return config[i].Sort < config[j].Sort
	})

	return &MultiGatewayManager{
		Gateways: config,
	}
}

// GetStringField 获取字段值，若不存在则返回默认值
func GetStringField(data map[string]interface{}, key string, defaultValue string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return defaultValue
}

// GetIntField 获取字段值，若不存在则返回默认值，返回值为 int
func GetIntField(data map[string]interface{}, key string, defaultValue int) int {
	if value, ok := data[key].(string); ok {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
