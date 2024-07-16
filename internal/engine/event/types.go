package event

const (
	OrderStatusModifyEventName = "order_status_modify_events"
)

type Status uint8

func (s Status) ToUint8() uint8 {
	return uint8(s)
}

const (
	// START 等待开始
	START Status = 1
	// PROCESS 流程运行中
	PROCESS Status = 2
	// END 完成
	END Status = 3
	// RETRY 重试
	RETRY Status = 4
)

type OrderStatusModifyEvent struct {
	ProcessInstanceId int    `json:"process_instance_id"`
	Status            Status `json:"status"`
}
