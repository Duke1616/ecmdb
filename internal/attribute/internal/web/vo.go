package web

type CreateAttributeReq struct {
	Name      string `json:"name"`
	ModelUID  string `json:"model_uid"`
	FieldName string `json:"field_name"`
	FieldType string `json:"field_type"`
	Required  bool   `json:"required"`
}

type DetailAttributeReq struct {
	ModelUid string `json:"model_uid"`
}

type ListAttributeReq struct {
	ModelUID string `json:"model_uid"`
}
