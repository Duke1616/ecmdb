package ioc

import (
	"github.com/Duke1616/ecmdb/internal/endpoint"
	"github.com/Duke1616/ecmdb/internal/event/service/easyflow"
	grpcpkg "github.com/Duke1616/ework-runner/pkg/grpc"
	"github.com/gotomicro/ego/server/egin"
	"github.com/gotomicro/ego/task/ecron"
)

type App struct {
	Web    *egin.Component
	Server *grpcpkg.Server
	Event  *easyflow.ProcessEvent
	Jobs   []*ecron.Component
	Svc    endpoint.Service
}
