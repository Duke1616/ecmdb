package registry

import (
	"context"
	"io"
)

type Registry interface {
	Register(ctx context.Context, si Instance) error
	UnRegister(ctx context.Context, si Instance) error

	ListWorkers(ctx context.Context, name string) ([]Instance, error)
	Subscribe(name string) <-chan Event

	io.Closer
}

type Instance struct {
	Name  string `yaml:"name" json:"name"`   // 实例名称
	Desc  string `yaml:"desc" json:"desc"`   // 注解
	Topic string `yaml:"topic" json:"topic"` // 建立 Topic 通道
}

type EventType int

const (
	EventTypeUnknown EventType = iota
	EventTypeAdd
	EventTypeDelete
)

type Event struct {
	Type     EventType
	Key      string
	Instance Instance
}
