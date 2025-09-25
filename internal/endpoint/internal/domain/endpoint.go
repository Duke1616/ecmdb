package domain

type Endpoint struct {
	Id           int64  // ID 信息
	Path         string // 路径
	Method       string // 方法
	Resource     string // 资源服务
	Desc         string // 注释
	IsAuth       bool   // 是否登陆
	IsAudit      bool   // 是否审计
	IsPermission bool   // 是否鉴权
}
