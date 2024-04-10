package web

type CreateAttributeReq struct {
	Name       string `json:"name"`
	ModelID    int64  `json:"model_id"`
	Identifies string `json:"identifies"`
	FieldType  string `json:"field_type"`
	Required   bool   `json:"required"`
}
