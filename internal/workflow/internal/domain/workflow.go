package domain

type Workflow struct {
	Id         int64
	TemplateId int64
	Name       string
	Icon       string
	Owner      string
	Desc       string
	FlowData   LogicFlow
}

type LogicFlow struct {
	Edges []map[string]interface{} `json:"edges"`
	Nodes []map[string]interface{} `json:"nodes"`
}
