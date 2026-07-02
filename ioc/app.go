package ioc

import (
	"context"

	"github.com/gotomicro/ego/server/egin"
)

type Task interface {
	Start(ctx context.Context)
}

type App struct {
	Web   *egin.Component
	Tasks []Task
}

func (a *App) StartBackgroundTasks(ctx context.Context) {
	for _, t := range a.Tasks {
		go func(t Task) {
			t.Start(ctx)
		}(t)
	}
}
