package web

type Effect string

const (
	// ALLOW 同意
	ALLOW Effect = "allow"
	// DENY 拒绝
	DENY Effect = "deny"
)

type PolicyReq struct {
	RoleCode string   `json:"role_code"`
	Policies []Policy `json:"policies"`
}

type Policy struct {
	Path   string `json:"path"`
	Method string `json:"method"`
	Effect Effect `json:"effect"`
}

type AddGroupingPolicyReq struct {
	UserId   string `json:"user_id"`
	RoleCode string `json:"role_code"`
}

type GetPermissionsForUserReq struct {
	UserId string `json:"user_id"`
}

type GetPermissionsForRoleReq struct {
	RoleCode string `json:"role_code"`
}

type AuthorizeReq struct {
	UserId string `json:"user_id"`
	Path   string `json:"path"`
	Method string `json:"method"`
}

type RetrievePolicies struct {
	Policies []Policy `json:"policies"`
}
