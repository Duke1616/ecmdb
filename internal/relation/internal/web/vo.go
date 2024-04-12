package web

type CreateRelationTypeReq struct {
	name string
}

type CreateModelRelationReq struct {
	SourceModelIdentifies  string `json:"source_model_identifies"`
	TargetModelIdentifies  string `json:"target_model_identifies"`
	RelationTypeIdentifies string `json:"relation_type_identifies"`
	Mapping                string `json:"mapping"`
}

type CreateResourceRelationReq struct {
	SourceModelIdentifies  string `json:"source_model_identifies"`
	TargetModelIdentifies  string `json:"target_model_identifies"`
	SourceResourceID       int64  `json:"source_resource_id"`
	TargetResourceID       int64  `json:"target_resource_id"`
	RelationTypeIdentifies string `json:"relation_type_identifies"`
}

type Page struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}
