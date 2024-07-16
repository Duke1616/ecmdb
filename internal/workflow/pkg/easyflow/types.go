package easyflow

type ProcessEngineConvert interface {
	Deploy(workflow Workflow) (int, error)
}

type Workflow struct {
	Id       int64
	Name     string
	Owner    string
	FlowData LogicFlow
}

type LogicFlow struct {
	Edges []map[string]interface{} `json:"edges"`
	Nodes []map[string]interface{} `json:"nodes"`
}

// Edge 定义线字段
type Edge struct {
	Type         string      `json:"type"`
	SourceNodeId string      `json:"sourceNodeId"`
	TargetNodeId string      `json:"targetNodeId"`
	Properties   interface{} `json:"properties"`
	ID           string      `json:"id"`
}

// Node 节点定义
type Node struct {
	Type       string      `json:"type"`
	Properties interface{} `json:"properties"`
	ID         string      `json:"id"`
}

type EdgeProperty struct {
	Expression string `json:"expression"`
}

type UserProperty struct {
	Name     string `json:"name"`
	Approved string `json:"approved"`
}

type StartProperty struct {
	Name string `json:"name"`
}

type EndProperty struct {
	Name string `json:"name"`
}

type ConditionProperty struct {
	Name string `json:"name"`
}
