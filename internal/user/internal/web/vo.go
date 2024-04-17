package web

type LoginLdapReq struct {
	User     string `json:"user"`
	Password string `json:"password"`
}
