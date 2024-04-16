package ldapx

type Profile struct {
	DN          string
	Email       string
	Username    string
	Position    string
	DisplayName string
	Groups      []string
}

type Config struct {
	Url                  string
	BindDN               string
	BindPassword         string
	UsernameAttribute    string
	MailAttribute        string
	DisplayNameAttribute string
	BaseDN               string
	UserFilter           string
	GroupFilter          string
	GroupNameAttribute   string
}
