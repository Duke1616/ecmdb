package domain

type Status uint8

func (s Status) ToUint8() uint8 {
	return uint8(s)
}

const (
	// RUNNING 启用
	RUNNING Status = 1
	// STOPPING 停止
	STOPPING Status = 2
	// OFFLINE 离线
	OFFLINE Status = 3
)

type Worker struct {
	Id     int64
	Key    string
	Name   string
	Desc   string
	Topic  string
	Status Status
}

type Execute struct {
	Topic    string
	Name     string // 执行名称
	UUID     string // 唯一标识
	Language string // 语言
	Code     string // 代码
}
