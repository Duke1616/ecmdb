package web

type CreateDepartmentReq struct {
	Pid        int64   `json:"pid"`
	Name       string  `json:"name"`
	Sort       int64   `json:"sort"`
	Enabled    bool    `json:"enabled"`
	Leaders    []int64 `json:"leaders"`
	MainLeader int64   `json:"main_leader"`
}

type UpdateDepartmentReq struct {
	Id         int64   `json:"id"`
	Pid        int64   `json:"pid"`
	Name       string  `json:"name"`
	Sort       int64   `json:"sort"`
	Enabled    bool    `json:"enabled"`
	Leaders    []int64 `json:"leaders"`
	MainLeader int64   `json:"main_leader"`
}

type Department struct {
	Id         int64         `json:"id"`
	Pid        int64         `json:"pid"`
	Name       string        `json:"name"`
	Sort       int64         `json:"sort"`
	Enabled    bool          `json:"enabled"`
	Leaders    []int64       `json:"leaders"`
	MainLeader int64         `json:"main_leader"`
	Children   []*Department `json:"children"`
}
