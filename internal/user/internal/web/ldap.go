package web

type LdapUser struct {
	Username      string `json:"username"`
	Email         string `json:"email"`
	Title         string `json:"title"`
	DisplayName   string `json:"display_name"`
	IsSystemExist bool   `json:"is_system_exist"`
}

type RetrieveLdapUsers struct {
	Users []LdapUser `json:"users"`
	Total int64      `json:"total"`
}
