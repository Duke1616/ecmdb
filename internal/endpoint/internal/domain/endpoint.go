package domain

type Endpoint struct {
	Id           int64
	Path         string
	Method       string
	Resource     string
	Desc         string
	IsAuth       bool
	IsAudit      bool
	IsPermission bool
}
