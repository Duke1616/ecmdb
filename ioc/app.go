package ioc

import (
	"github.com/Duke1616/ecmdb/internal/endpoint"
	"github.com/Duke1616/ecmdb/internal/event/service/easyflow"
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/task/ecron"
)

type App struct {
	Web   *gin.Engine
	Event *easyflow.ProcessEvent
	Jobs  []*ecron.Component
	Svc   endpoint.Service
}
