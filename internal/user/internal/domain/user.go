package domain

const (
	Ldap   = iota + 1 // LDAP 创建
	System            // 系统 创建
)

const (
	LdapSync = iota + 1
	UserRegistry
)

type User struct {
	ID         int64
	Username   string
	Password   string
	Email      string
	Title      string
	SourceType int64
	CreateType int64
}
