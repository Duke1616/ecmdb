package domain

type Status uint8

func (s Status) ToUint8() uint8 {
	return uint8(s)
}

const (
	// START 等待开始
	START Status = 1
	// RUNNING 运行中
	RUNNING Status = 2
	// END 完成
	END Status = 3
	// RETRY 重试
	RETRY Status = 4
)

type Order struct {
	Id         int64
	TemplateId int64
	FlowId     int64
	Data       map[string]interface{}
	Status     Status
	CreateBy   string
}
