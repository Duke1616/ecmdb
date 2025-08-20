package provider

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/event/domain"
)

type Provider interface {
	// Send 发送消息
	Send(ctx context.Context, notification domain.Notification) (bool, error)
}

// Selector 供应商选择器接口
type Selector interface {
	// Next 获取下一个供应商，无可用供应商时返回错误
	Next(ctx context.Context, notification domain.Notification) (Provider, error)
}

// SelectorBuilder 供应商选择器的构造器
type SelectorBuilder interface {
	// Build 构造选择器，可以在Build方法上添加参数来构建更复杂的选择器
	Build() (Selector, error)
}
