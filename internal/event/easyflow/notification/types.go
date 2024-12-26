package notification

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/workflow"
)

type Notification interface {
	Send(ctx context.Context, nOrder order.Order, wf workflow.Workflow, instanceId int, nodeId string,
		userIds []string) (bool, error)
}
