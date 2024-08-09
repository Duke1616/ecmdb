package web

type CreateMenuReq struct {
	Pid           int64      `json:"pid"`
	Path          string     `json:"path"`
	Name          string     `json:"name"`
	Component     string     `json:"component"`
	ComponentPath string     `json:"component_path"`
	Redirect      string     `json:"redirect"`
	Sort          int64      `json:"sort"`
	Type          uint8      `json:"type"`
	Status        uint8      `json:"status"`
	Meta          Meta       `json:"meta"`
	Endpoints     []Endpoint `json:"endpoints"`
}

type UpdateMenuReq struct {
	Id            int64      `json:"id"`
	Pid           int64      `json:"pid"`
	Name          string     `json:"name"`
	Path          string     `json:"path"`
	Component     string     `json:"component"`
	ComponentPath string     `json:"component_path"`
	Redirect      string     `json:"redirect"`
	Sort          int64      `json:"sort"`
	Type          uint8      `json:"type"`
	Status        uint8      `json:"status"`
	Meta          Meta       `json:"meta"`
	Endpoints     []Endpoint `json:"endpoints"`
}

type Endpoint struct {
	Id     int64  `json:"id"`
	Path   string `json:"path"`
	Method string `json:"method"`
	Desc   string `json:"desc"`
}

type Meta struct {
	Title       string `json:"title"`        // 展示名称
	IsHidden    bool   `json:"is_hidden"`    // 是否展示
	IsAffix     bool   `json:"is_affix"`     // 是否固定
	IsKeepAlive bool   `json:"is_keepalive"` // 是否缓存
	Icon        string `json:"icon"`         // Icon图标
}

type Menu struct {
	Id            int64      `json:"id"`
	Pid           int64      `json:"pid"`
	Name          string     `json:"name"`
	Path          string     `json:"path"`
	Redirect      string     `json:"redirect"`
	Sort          int64      `json:"sort"`
	Component     string     `json:"component"`
	ComponentPath string     `json:"component_path"`
	Status        uint8      `json:"status"`
	Type          uint8      `json:"type"`
	Meta          Meta       `json:"meta"`
	Endpoints     []Endpoint `json:"endpoints"`
	Children      []*Menu    `json:"children"`
}
