package ioc

import (
	"github.com/Duke1616/ecmdb/cmd/initial/version"
	"github.com/Duke1616/ecmdb/internal/role"
	"github.com/Duke1616/ecmdb/internal/user"
)

type App struct {
	UserSvc user.Service
	RoleSvc role.Service
	VerSvc  version.Service
}
