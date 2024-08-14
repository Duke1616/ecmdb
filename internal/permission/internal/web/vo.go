package web

type RolePermissionReq struct {
	RoleCode string `json:"role_code"`
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

type ChangePermissionForRoleReq struct {
	RoleCode string  `json:"role_code"`
	MenuIds  []int64 `json:"menu_ids"`
}

type FindUserPermissionMenus struct {
	UserId int64 `json:"user_id"`
}

type RetrieveRolePermission struct {
	AuthzIds []int64 `json:"authz_ids"`
	Menu     []*Menu `json:"menus"`
}

type RetrieveUserPermission struct {
	Menu []*Menu `json:"menus"`
}
