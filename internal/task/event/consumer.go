package event

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
	"reflect"
)

type TaskEventConsumer struct {
	workFlowSvc workflow.Service
	consumer    mq.Consumer
	logger      *elog.Component
}

func NewTaskEventConsumer(q mq.MQ, workFlowSvc workflow.Service) (*TaskEventConsumer, error) {
	groupID := "task"
	consumer, err := q.Consumer(CreateFLowEventName, groupID)
	if err != nil {
		return nil, err
	}
	return &TaskEventConsumer{
		consumer:    consumer,
		workFlowSvc: workFlowSvc,
		logger:      elog.DefaultLogger,
	}, nil
}

func (c *TaskEventConsumer) Start(ctx context.Context) {
	go func() {
		for {
			err := c.Consume(ctx)
			if err != nil {
				c.logger.Error("同步事件失败", elog.Any("err", err))
			}
		}
	}()
}

func (c *TaskEventConsumer) Consume(ctx context.Context) error {
	cm, err := c.consumer.Consume(ctx)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}
	var evt OrderEvent
	if err = json.Unmarshal(cm.Value, &evt); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}
	flow, err := c.workFlowSvc.Find(ctx, evt.FlowId)
	if err != nil {
		return err
	}

	_, err = engine.InstanceStart(flow.ProcessId, "业务申请", flow.Name, string(c.Variables(evt)))
	return err
}

func (c *TaskEventConsumer) Stop(_ context.Context) error {
	return c.consumer.Close()
}

func (c *TaskEventConsumer) Variables(evt OrderEvent) []byte {
	var vars []Variables
	for key, value := range evt.Data {
		// 判断如果浮点数类型，转换成string
		strValue := value
		valueType := reflect.TypeOf(value)
		if valueType.Kind() == reflect.Float64 {
			strValue = fmt.Sprintf("%f", value)
		}

		vars = append(vars, Variables{
			Key:   key,
			Value: strValue,
		})
	}
	VariablesJson, _ := json.Marshal(vars)

	return VariablesJson
}
