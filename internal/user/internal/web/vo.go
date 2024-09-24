package web

type LoginLdapReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginSystemReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterUserReq struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	RePassword  string `json:"re_password"`
	DisplayName string `json:"display_name"`
}

type User struct {
	Id           int64      `json:"id"`
	DepartmentId int64      `json:"department_id"`
	Username     string     `json:"username"`
	Password     string     `json:"password"`
	Email        string     `json:"email"`
	Title        string     `json:"title"`
	DisplayName  string     `json:"display_name"`
	CreateType   uint8      `json:"create_type"`
	RoleCodes    []string   `json:"role_codes"`
	FeishuInfo   FeishuInfo `json:"feishu_info"`
}

type UpdateUserReq struct {
	Id           int64      `json:"id"`
	DepartmentId int64      `json:"department_id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	Title        string     `json:"title"`
	DisplayName  string     `json:"display_name"`
	FeishuInfo   FeishuInfo `json:"feishu_info"`
}

type FeishuInfo struct {
	UserId string `json:"user_id"`
}

type FindByUsernameRegexReq struct {
	Page
	Username string `json:"username"`
}

type FindByUserNamesReq struct {
	Usernames []string `json:"usernames"`
}

type FindByUserNameReq struct {
	Username string `json:"username"`
}

type FindUsersByDepartmentIdReq struct {
	Page
	DepartmentId int64 `json:"department_id"`
}

type UserBindRoleReq struct {
	Id        int64    `json:"id"`
	RoleCodes []string `json:"role_codes"`
}

type Page struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

type RetrieveUsers struct {
	Total int64  `json:"total"`
	Users []User `json:"users"`
}

// UserDepartmentCombination 一组数据
type UserDepartmentCombination struct {
	Id          int64                        `json:"id"`
	Type        string                       `json:"type"`
	DisplayName string                       `json:"display_name"`
	Name        string                       `json:"name"`
	Disabled    bool                         `json:"disabled"`
	Sort        int64                        `json:"sort"`
	Children    []*UserDepartmentCombination `json:"children"`
}
