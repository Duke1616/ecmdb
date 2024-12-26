package notification

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/workflow"
)

type NotifyParams struct {
	InstanceId   int
	UserIDs      []string
	NodeId       string
	WantResult   map[string]interface{}
	NotifyMethod string
}

type Notification interface {
	SendAction[any]
}

type SendAction[T any] interface {
	Send(ctx context.Context, nOrder order.Order, params NotifyParams) (bool, error)
	// UnmarshalProperty(ctx context.Context, wf workflow.Workflow, nodeId string) (T, error)
	IsNotification(ctx context.Context, wf workflow.Workflow, instanceId int,
		nodeId string) (bool, map[string]interface{}, error)
}

type Notifications interface {
	GetAutomation() User
	GetUser() Automation
}

type Automation interface {
	Send(ctx context.Context, nOrder order.Order) (bool, error)
	IsNotify(ctx context.Context, wf workflow.Workflow, nodeId string) (bool, error)
	WantResult(ctx context.Context) (map[string]interface{}, error)
}

type User interface {
	Send(ctx context.Context, nOrder order.Order, params NotifyParams) (bool, error)
	IsNotify(ctx context.Context, wf workflow.Workflow) (bool, error)
	WantResult(ctx context.Context) (map[string]interface{}, error)
}
