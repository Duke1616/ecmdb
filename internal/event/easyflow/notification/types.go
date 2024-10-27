package notification

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/workflow"
)

type Notification interface {
	SendAction
}

type SendAction interface {
	Send(ctx context.Context, instanceId int, userIDs []string) (bool, error)
	IsNotification(ctx context.Context, InstanceId int) (order.Order, workflow.NotifyMethod, bool, error)
}
