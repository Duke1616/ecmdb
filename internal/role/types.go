package role

import (
	"github.com/Duke1616/ecmdb/internal/role/internal/domain"
	"github.com/Duke1616/ecmdb/internal/role/internal/service"
	"github.com/Duke1616/ecmdb/internal/role/internal/web"
)

type Handler = web.Handler

type Service = service.Service

type Role = domain.Role
