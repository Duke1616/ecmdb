package domain

const (
	Ldap   = iota + 1 // LDAP 创建
	System            // 系统 创建
)

const (
	LdapSync = iota + 1
	UserRegistry
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
	ID         int64
	Username   string
	Password   string
	Email      string
	Title      string
	Status     Status
	SourceType int64
	CreateType int64
	RoleCodes  []string
}
