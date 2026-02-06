package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/codebook"
	"github.com/Duke1616/ecmdb/internal/task/internal/domain"
	"github.com/Duke1616/ecmdb/internal/task/internal/service"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/enotify/notify"
	"github.com/Duke1616/enotify/notify/feishu"
	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
	lark "github.com/larksuite/oapi-sdk-go/v3"
)

type ExecuteResultConsumer struct {
	consumer    mq.Consumer
	handler     notify.Handler
	codebookSvc codebook.Service
	userSvc     user.Service
	svc         service.Service
	logger      *elog.Component
}

func NewExecuteResultConsumer(q mq.MQ, svc service.Service, codebookSvc codebook.Service,
	userSvc user.Service, lark *lark.Client) (
	*ExecuteResultConsumer, error) {
	groupID := "task_receive_execute"
	consumer, err := q.Consumer(ExecuteResultEventName, groupID)
	if err != nil {
		return nil, err
	}

	handler, err := feishu.NewHandler(lark)
	if err != nil {
		return nil, err
	}

	return &ExecuteResultConsumer{
		consumer:    consumer,
		svc:         svc,
		codebookSvc: codebookSvc,
		userSvc:     userSvc,
		handler:     handler,
		logger:      elog.DefaultLogger,
	}, nil
}

func (c *ExecuteResultConsumer) Start(ctx context.Context) {
	go func() {
		for {
			err := c.Consume(ctx)
			if err != nil {
				c.logger.Error("同步修改任务执行状态失败", elog.Any("错误信息", err))
			}
		}
	}()
}

func (c *ExecuteResultConsumer) Consume(ctx context.Context) error {
	cm, err := c.consumer.Consume(ctx)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}
	var evt ExecuteResultEvent
	if err = json.Unmarshal(cm.Value, &evt); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}

	_, err = c.svc.UpdateTaskStatus(ctx, domain.TaskResult{
		Id:         evt.TaskId,
		Result:     evt.Result,
		WantResult: evt.WantResult,
		Status:     domain.Status(evt.Status),
	})

	if domain.Status(evt.Status) == domain.FAILED {
		return c.failedNotify(ctx, evt.TaskId)
	}

	return err
}

// failedNotify 发送消息通知给自动化任务模版的管理者
func (c *ExecuteResultConsumer) failedNotify(ctx context.Context, id int64) error {
	t, err := c.svc.Detail(ctx, id)
	if err != nil {
		return err
	}

	code, err := c.codebookSvc.FindByUid(ctx, t.CodebookUid)
	if err != nil {
		return err
	}

	u, err := c.userSvc.FindByUsername(ctx, code.Owner)
	if err != nil {
		return err
	}

	content := fmt.Sprintf(`{"text": "任务执行失败, 请通过平台进行查看，任务ID: %d, 工作节点: %s"}`,
		id, t.WorkerName)

	msg := feishu.NewCreateBuilder(u.FeishuInfo.UserId).SetReceiveIDType(feishu.ReceiveIDTypeUserID).
		SetContent(feishu.NewFeishuCustom("text", content)).Build()

	if err = c.handler.Send(ctx, msg); err != nil {
		return fmt.Errorf("任务执行失败，触发发送信息失败: %w, 工单ID: %d", err, id)
	}

	return nil
}

func (c *ExecuteResultConsumer) Stop(_ context.Context) error {
	return c.consumer.Close()
}
