package domain

type Department struct {
	Id      int64
	Pid     int64
	Name    string
	Sort    int64
	Enabled bool
	// 部分负责人，可能有多个人
	Leaders []int64
	// 分管领导，部门最大的领导
	MainLeader int64
}
