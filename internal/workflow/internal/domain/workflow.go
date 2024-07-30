package domain

type Workflow struct {
	Id         int64
	TemplateId int64
	Name       string
	Icon       string
	Owner      string
	Desc       string
	FlowData   LogicFlow // 前端数据传递Flow数据
	ProcessId  int       // 绑定对应的后端引擎 ID
}

type LogicFlow struct {
	Edges []map[string]interface{} `json:"edges"`
	Nodes []map[string]interface{} `json:"nodes"`
}
