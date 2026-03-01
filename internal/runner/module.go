package runner

import (
	"github.com/Duke1616/ecmdb/internal/runner/internal/service"
	"github.com/Duke1616/ecmdb/internal/runner/internal/web"
)

type Module struct {
	Svc service.Service
	Hdl *web.Handler
}
