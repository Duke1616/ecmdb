package domain

type Type uint8

func (s Type) ToUint8() uint8 {
	return uint8(s)
}

const (
	// DIR 目录
	DIR Type = 1
	// MENU 菜单
	MENU Type = 2
	// BUTTON 按钮
	BUTTON Type = 3
)

type Status uint8

func (s Status) ToUint8() uint8 {
	return uint8(s)
}

const (
	// ENABLED 启用
	ENABLED Status = 1
	// DISABLED 禁用
	DISABLED Status = 2
)

type Menu struct {
	Id        int64
	Pid       int64
	Path      string
	Name      string
	Sort      int64
	Component string
	Redirect  string
	Status    Status
	Type      Type
	Meta      Meta
	Endpoints []Endpoint
}

type Endpoint struct {
	Path   string
	Method string
	Desc   string
}

type Meta struct {
	Title       string // 展示名称
	IsHidden    bool   // 是否展示
	IsAffix     bool   // 是否固定
	IsKeepAlive bool   // 是否缓存
	Icon        string // Icon图标
}
