package full

import "github.com/Duke1616/ecmdb/cmd/initial/ioc"

// 用户相关
const (
	UserName    = "admin"
	Password    = "123456"
	DisPlayName = "超级管理员"
)

// 角色相关
const (
	RoleCode = "admin"
	Desc     = ""
)

type InitialFull interface {
	InitUser() (int64, error)
	InitRole() error
	InitMenu() error
	InitPermission(userID int64) error
}

type fullInitial struct {
	App *ioc.App
}

func NewInitial(app *ioc.App) InitialFull {
	return &fullInitial{
		App: app,
	}
}
