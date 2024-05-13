package web

type LoginLdapReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Email        string `json:"email"`
	Title        string `json:"title"`
	SourceType   int64  `json:"source_type"`
	CreateType   int64  `json:"create_type"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
