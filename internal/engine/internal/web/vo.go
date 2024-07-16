package web

type Pass struct {
	TaskId    int        `json:"task_id"`
	Comment   string     `json:"comment"`
	Variables []Variable `json:"variables"`
}

type Variable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
