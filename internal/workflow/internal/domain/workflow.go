package domain

type NotifyMethod uint8

func (s NotifyMethod) ToUint8() uint8 {
	return uint8(s)
}

const (
	// Feishu 飞书
	Feishu NotifyMethod = 1
	// Wechat 企业微信
	Wechat NotifyMethod = 2
)

type Workflow struct {
	Id           int64
	TemplateId   int64
	Name         string
	Icon         string
	Owner        string
	Desc         string
	IsNotify     bool
	NotifyMethod NotifyMethod
	FlowData     LogicFlow // 前端数据传递Flow数据
	ProcessId    int       // 绑定对应的后端引擎 ID
}

type LogicFlow struct {
	Edges []map[string]interface{} `json:"edges"`
	Nodes []map[string]interface{} `json:"nodes"`
}
