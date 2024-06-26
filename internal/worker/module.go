package worker

import (
	"github.com/Duke1616/ecmdb/internal/worker/internal/event/watch"
)

type Module struct {
	Svc Service
	w   *watch.TaskWorkerWatch
	Hdl *Handler
}
