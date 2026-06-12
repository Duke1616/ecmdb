package domain

type FieldSecureAttrChange struct {
	ModelUid   string `json:"model_uid"` // 模型唯一标识
	FieldUid   string `json:"field_uid"` // 模型字段唯一标识
	Secure     bool   `json:"secure"`    // 安全字段修改状态
	TiggerTime int64  `json:"ctime"`     // 触发时间
}

type FieldDelete struct {
	ModelUid    string `json:"model_uid"`    // 模型唯一标识
	FieldUid    string `json:"field_uid"`    // 字段唯一标识
	TriggerTime int64  `json:"trigger_time"` // 触发时间
}
