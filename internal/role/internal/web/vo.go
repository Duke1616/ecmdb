package web

type CreateRoleReq struct {
	Name   string `json:"name"`
	Code   string `json:"code"`
	Desc   string `json:"desc"`
	Status bool   `json:"status"`
}

type UpdateRoleReq struct {
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
	Desc   string `json:"desc"`
	Status bool   `json:"status"`
}

type Role struct {
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
	Desc   string `json:"desc"`
	Status bool   `json:"status"`
}

type Page struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

type RetrieveRoles struct {
	Roles []Role `json:"roles"`
	Total int64  `json:"total"`
}

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

type RetrieveRolePermission struct {
	AuthzIds []int64 `json:"authz_ids"`
	Menu     []*Menu `json:"menus"`
}

type AddPermissionForRoleReq struct {
	RoleCode string  `json:"role_code"`
	MenuIds  []int64 `json:"menu_ids"`
}

type UserRole struct {
	Page
	Codes []string `json:"codes"`
}

type RetrieveUserDoesNotHaveRoles struct {
	Total int64  `json:"total"`
	Roles []Role `json:"roles"`
}

type RetrieveUserHaveRoles struct {
	Roles []Role `json:"roles"`
}
