package ioc

import (
	"github.com/Duke1616/ecmdb/internal/event/easyflow"
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/task/ecron"
)

type App struct {
	Web   *gin.Engine
	Event *easyflow.ProcessEvent
	Jobs  []*ecron.Component
}
