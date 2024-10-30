package notification

import (
	"context"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/event/easyflow/notification/method"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/enotify/notify"
)

type NotifyParams struct {
	InstanceId   int
	UserIDs      []string
	NodeId       string
	WantResult   map[string]interface{}
	NotifyMethod string
}

type Notification interface {
	SendAction
}

type SendAction interface {
	Send(ctx context.Context, nOrder order.Order, params NotifyParams) (bool, error)
	IsNotification(ctx context.Context, wf workflow.Workflow, instanceId int,
		nodeId string) (bool, map[string]interface{}, error)
}

type TaskFetcher interface {
	FetchTasksWithRetry(ctx context.Context, instanceId int, userIDs []string) ([]model.Task, error)
}

type BuilderNotification interface {
	BuildMessages(rules []method.Rule, nOrder order.Order, startUser string, users []user.User,
		tasks []model.Task, notifyMethod string) []notify.NotifierWrap
}
