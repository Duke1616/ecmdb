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
	Id          int64
	Username    string
	Password    string
	Email       string
	Title       string
	DisplayName string
	Status      Status
	CreateType  CreateType
	RoleCodes   []string
	FeishuInfo  FeishuInfo
}
