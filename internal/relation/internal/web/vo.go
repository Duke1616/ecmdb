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
