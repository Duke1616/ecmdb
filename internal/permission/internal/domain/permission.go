package domain

type Action uint8

func (s Action) ToUint8() uint8 {
	return uint8(s)
}

const (
	// CREATE 创建动作，比如 ADMIN 超级管理员自动权限录入
	// TODO 拥有根菜单权限录入
	CREATE Action = 1
	// WRITE 写入动作
	WRITE Action = 2
	// DELETE 全部删除、重新录入数据
	DELETE Action = 3
)

type MenuEvent struct {
	Menu   Menu   `json:"menu"`
	Action Action `json:"action"`
}

type Menu struct {
	Id        int64      `json:"id"`
	Endpoints []Endpoint `json:"endpoints"`
}

type Endpoint struct {
	Path     string `json:"path"`
	Method   string `json:"method"`
	Resource string `json:"resource"`
}
