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
