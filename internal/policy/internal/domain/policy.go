package domain

type Effect string

func (s Effect) ToString() string {
	return string(s)
}

const (
	// ALLOW 同意
	ALLOW Effect = "allow"
	// DENY 拒绝
	DENY Effect = "deny"
)

type Policies struct {
	RoleCode string
	Policies []Policy
}

type Policy struct {
	Path   string
	Method string
	Effect Effect
}

type AddGroupingPolicy struct {
	UserId   string
	RoleCode string
}
