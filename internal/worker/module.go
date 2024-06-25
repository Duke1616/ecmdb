package worker

import "github.com/Duke1616/ecmdb/internal/worker/internal/event"

type Module struct {
	Svc Service
	w   *event.TaskWorkerWatch
	Hdl *Handler
}
