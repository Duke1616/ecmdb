package domain

type CreateType uint8

func (s CreateType) ToUint8() uint8 {
	return uint8(s)
}

const (
	// LDAP LDAP创建
	LDAP CreateType = 1
	// SYSTEM 系统创建
	SYSTEM CreateType = 2
)

type Status uint8

func (s Status) ToUint8() uint8 {
	return uint8(s)
}

const (
	// ENABLED 启用
	ENABLED Status = 1
	// DISABLED 禁用
	DISABLED Status = 2
)

type User struct {
	Id           int64      `json:"id"`
	DepartmentId int64      `json:"department_id"`
	Username     string     `json:"username"`
	Password     string     `json:"password"`
	Email        string     `json:"email"`
	Title        string     `json:"title"`
	DisplayName  string     `json:"display_name"`
	Status       Status     `json:"status"`
	CreateType   CreateType `json:"create_type"`
	RoleCodes    []string   `json:"role_codes"`
	FeishuInfo   FeishuInfo `json:"feishu_info"`
	WechatInfo   WechatInfo `json:"wechat_info"`
}

type UserCombination struct {
	DepartMentId int64
	Total        int
	Users        []User
}
