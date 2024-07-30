package workflow

import (
	"github.com/Duke1616/ecmdb/internal/workflow/internal/service"
	"github.com/Duke1616/ecmdb/internal/workflow/internal/web"
)

type Module struct {
	Hdl *web.Handler
	Svc service.Service
}
