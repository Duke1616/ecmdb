package web

type CreateAttributeReq struct {
	Name            string `json:"name"`
	ModelIdentifies string `json:"model_identifies"`
	Identifies      string `json:"identifies"`
	FieldType       string `json:"field_type"`
	Required        bool   `json:"required"`
}
