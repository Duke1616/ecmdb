package department

import (
	"github.com/Duke1616/ecmdb/internal/department/internal/domain"
	"github.com/Duke1616/ecmdb/internal/department/internal/service"
	"github.com/Duke1616/ecmdb/internal/department/internal/web"
)

type Handler = web.Handler

type Department = domain.Department

type Service = service.Service
