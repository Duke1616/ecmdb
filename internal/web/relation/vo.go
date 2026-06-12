package web

// CreateRelationTypeReq 关联关系类型
type CreateRelationTypeReq struct {
	Name           string `json:"name"`
	UID            string `json:"uid"`
	SourceDescribe string `json:"source_describe"`
	TargetDescribe string `json:"target_describe"`
}

type UpdateRelationTypeReq struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	SourceDescribe string `json:"source_describe"`
	TargetDescribe string `json:"target_describe"`
}

type DeleteRelationTypeReq struct {
	Id int64 `json:"id"`
}

type Page struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

type RetrieveRelationType struct {
	Total         int64          `json:"total,omitempty"`
	RelationTypes []RelationType `json:"relation_types,omitempty"`
}
