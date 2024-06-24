package web

type Status uint8

func (s Status) ToUint8() uint8 {
	return uint8(s)
}

const (
	// RUNNING 启用
	RUNNING Status = 1
	// STOPPING 停止
	STOPPING Status = 2
)

type Page struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

type ListWorkerReq struct {
	Page
}

type Worker struct {
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	Topic  string `json:"topic"`
	Status Status `json:"status"`
}

type RetrieveWorkers struct {
	Total   int64    `json:"total"`
	Workers []Worker `json:"workers"`
}
