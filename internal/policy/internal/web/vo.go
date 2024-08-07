package web

type Effect string

const (
	// ALLOW 同意
	ALLOW Effect = "allow"
	// DENY 拒绝
	DENY Effect = "deny"
)

type PolicyReq struct {
	RoleName string   `json:"role_name"`
	Policies []Policy `json:"policies"`
}

type Policy struct {
	Path   string `json:"path"`
	Method string `json:"method"`
	Effect Effect `json:"effect"`
}

type AddGroupingPolicyReq struct {
	UserId   string `json:"user_id"`
	RoleName string `json:"role_name"`
}

type AuthorizeReq struct {
	UserId string `json:"user_id"`
	Path   string `json:"path"`
	Method string `json:"method"`
}
