package department

import (
	"github.com/Duke1616/ecmdb/internal/department/internal/service"
	"github.com/Duke1616/ecmdb/internal/department/internal/web"
)

type Module struct {
	Hdl *web.Handler
	Svc service.Service
}
