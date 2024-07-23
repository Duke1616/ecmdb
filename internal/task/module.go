package task

import "github.com/Duke1616/ecmdb/internal/task/internal/event"

type Module struct {
	Svc Service
	c   *event.ExecuteResultConsumer
}
