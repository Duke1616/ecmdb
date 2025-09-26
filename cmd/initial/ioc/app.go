package ioc

import (
	"github.com/Duke1616/ecmdb/cmd/initial/version"
	"github.com/Duke1616/ecmdb/internal/menu"
	"github.com/Duke1616/ecmdb/internal/permission"
	"github.com/Duke1616/ecmdb/internal/policy"
	"github.com/Duke1616/ecmdb/internal/role"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"gorm.io/gorm"
)

type App struct {
	UserSvc       user.Service
	RoleSvc       role.Service
	MenuSvc       menu.Service
	PermissionSvc permission.Service
	policySvc     policy.Service
	VerSvc        version.Service
	GormDB        *gorm.DB
	DB            *mongox.Mongo
}
