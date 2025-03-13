package task

import (
	"github.com/Duke1616/ecmdb/internal/task/internal/event"
	"github.com/Duke1616/ecmdb/internal/task/internal/web"
)

type Module struct {
	Svc                Service
	Hdl                *web.Handler
	c                  *event.ExecuteResultConsumer
	StartTaskJob       *StartTaskJob
	PassProcessTaskJob *PassProcessTaskJob
	RecoveryTaskJob    *RecoveryTaskJob
}
