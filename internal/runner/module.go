package runner

import (
	"github.com/Duke1616/ecmdb/internal/runner/service"
	"github.com/Duke1616/ecmdb/internal/runner/web"
)

type Module struct {
	Svc service.Service
	Hdl *web.Handler
}
