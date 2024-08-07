package web

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

type CreateMenuReq struct {
	Pid         int64   `json:"pid"`
	Name        string  `json:"name"`
	Path        string  `json:"path"`
	Sort        int64   `json:"sort"`
	IsRoot      bool    `json:"is_root"`
	Type        Type    `json:"type"`
	Meta        Meta    `json:"meta"`
	EndpointIds []int64 `json:"endpoint_ids"`
}

type Meta struct {
	Title       string `json:"title"`        // 展示名称
	IsHidden    bool   `json:"is_hidden"`    // 是否展示
	IsAffix     bool   `json:"is_affix"`     // 是否固定
	IsKeepAlive bool   `json:"is_keepalive"` // 是否缓存
	Icon        string `json:"icon"`         // Icon图标
}
