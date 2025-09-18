package event

const FieldSecureAttrChangeName = "field_secure_attr_change"

type FieldSecureAttrChange struct {
	ModelUid   string `json:"model_uid"` // 模型唯一标识
	FieldUid   string `json:"field_uid"` // 模型字段唯一标识
	Secure     bool   `json:"secure"`    // 安全字段修改状态
	TiggerTime int64  `json:"ctime"`     // 触发时间, 我可以根据这个时间进行查询
}
