package domain

type Strategy struct {
	Key   string
	Value []string
}

// Condition 过滤条件
type Condition struct {
	Key   string
	Value string
	Op    string
}

// Strategies 自动触发器规则
type Strategies struct {
	TemplateId         int64       // 模板ID
	CodebookIdentifier string      // 绑定任务脚本
	RunnerTag          string      // 执行节点标签
	Conditions         []Condition // 匹配条件
}
