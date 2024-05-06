package web

type LoginLdapReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
