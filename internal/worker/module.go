package worker

import "github.com/Duke1616/ecmdb/internal/worker/internal/event"

type Module struct {
	Svc Service
	c   *event.TaskWorkerConsumer
	Hdl *Handler
}
