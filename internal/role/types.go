package role

import (
	"github.com/Duke1616/ecmdb/internal/role/internal/domain"
	"github.com/Duke1616/ecmdb/internal/role/internal/service"
	"github.com/Duke1616/ecmdb/internal/role/internal/web"
)

type Handler = web.Handler

type Service = service.Service

const (
	AdminRole = domain.AdminRole
)

type Role = domain.Role
