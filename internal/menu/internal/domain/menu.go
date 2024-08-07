package domain

type Type uint8

func (s Type) ToUint8() uint8 {
	return uint8(s)
}

const (
	// MENU 菜单
	MENU Type = 1
	// BUTTON 按钮
	BUTTON Type = 2
)

type Menu struct {
	Id          int64
	Pid         int64
	Name        string
	Path        string
	Sort        int64
	IsRoot      bool
	Type        Type
	Meta        Meta
	EndpointIds []int64
}

type Meta struct {
	Title       string // 展示名称
	IsHidden    bool   // 是否展示
	IsAffix     bool   // 是否固定
	IsKeepAlive bool   // 是否缓存
	Icon        string // Icon图标
}
