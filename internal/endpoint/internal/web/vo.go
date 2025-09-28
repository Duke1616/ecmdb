package web

type RegisterEndpointReq struct {
	Path         string `bson:"path"`
	Method       string `bson:"method"`
	Resource     string `bson:"resource"`
	Desc         string `bson:"desc"`
	IsAuth       bool   `bson:"is_auth"`
	IsAudit      bool   `bson:"is_audit"`
	IsPermission bool   `bson:"is_permission"`
}

type RegisterEndpointsReq struct {
	Resource        string                `json:"resource"`         // 资源名称
	RegisterEndpoint []RegisterEndpointReq `json:"register_endpoint"`
}

type Endpoint struct {
	Id           int64  `json:"id"`
	Path         string `json:"path"`
	Method       string `json:"method"`
	Resource     string `json:"resource"`
	Desc         string `json:"desc"`
	IsAuth       bool   `json:"is_auth"`
	IsAudit      bool   `json:"is_audit"`
	IsPermission bool   `json:"is_permission"`
}

type Page struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

type FilterPathReq struct {
	Page
	Path string `json:"path"`
}

type RetrieveEndpoints struct {
	Endpoints []Endpoint `json:"endpoints"`
	Total     int64      `json:"total"`
}
