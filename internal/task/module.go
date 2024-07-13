package task

import (
	"github.com/Duke1616/ecmdb/internal/task/event"
	"github.com/Duke1616/ecmdb/internal/task/web"
)

type Module struct {
	Hdl *web.Handler
	c   *event.TaskEventConsumer
}
