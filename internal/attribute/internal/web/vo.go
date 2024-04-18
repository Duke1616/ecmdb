package web

type CreateAttributeReq struct {
	Name      string `json:"name"`
	ModelUID  string `json:"model_uid"`
	UID       string `json:"uid"`
	FieldType string `json:"field_type"`
	Required  bool   `json:"required"`
}

type DetailAttributeReq struct {
	ModelUid string `json:"model_uid"`
}
