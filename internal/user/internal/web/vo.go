package web

type LoginLdapReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginSystemReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type User struct {
	ID         int64    `json:"id"`
	Username   string   `json:"username"`
	Password   string   `json:"password"`
	Email      string   `json:"email"`
	Title      string   `json:"title"`
	SourceType int64    `json:"source_type"`
	CreateType int64    `json:"create_type"`
	RoleCodes  []string `json:"role_codes"`
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
