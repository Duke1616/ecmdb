package ldapx

type Profile struct {
	DN          string
	Email       string
	Username    string
	Title       string
	DisplayName string
	Groups      []string
}

type Config struct {
	Url                  string `mapstructure:"url" json:"url,omitempty"`
	BaseDN               string `mapstructure:"base_dn" json:"base_dn,omitempty"`
	BindDN               string `mapstructure:"bind_dn" json:"bind_dn,omitempty"`
	BindPassword         string `mapstructure:"bind_password" json:"bind_password,omitempty"`
	UsernameAttribute    string `mapstructure:"username_attribute" json:"username_attribute,omitempty"`
	MailAttribute        string `mapstructure:"mail_attribute" json:"mail_attribute,omitempty"`
	DisplayNameAttribute string `mapstructure:"display_name_attribute" json:"display_name_attribute,omitempty"`
	TitleAttribute       string `mapstructure:"title_attribute" json:"title_attribute,omitempty"`
	UserFilter           string `mapstructure:"user_filter" json:"user_filter,omitempty"`
	GroupFilter          string `mapstructure:"group_filter" json:"group_filter"`
	GroupNameAttribute   string `mapstructure:"group_name_attribute" json:"group_name_attribute"`
}
