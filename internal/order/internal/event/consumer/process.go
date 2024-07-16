package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Duke1616/ecmdb/internal/order/internal/event"
	"github.com/Duke1616/ecmdb/internal/order/internal/service"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
	"reflect"
)

type ProcessEventConsumer struct {
	workFlowSvc workflow.Service
	svc         service.Service
	consumer    mq.Consumer
	logger      *elog.Component
}

func NewProcessEventConsumer(q mq.MQ, workFlowSvc workflow.Service, svc service.Service) (*ProcessEventConsumer, error) {
	groupID := "process_order"
	consumer, err := q.Consumer(event.CreateProcessEventName, groupID)
	if err != nil {
		return nil, err
	}

	return &ProcessEventConsumer{
		consumer:    consumer,
		workFlowSvc: workFlowSvc,
		svc:         svc,
		logger:      elog.DefaultLogger,
	}, nil
}

func (c *ProcessEventConsumer) Start(ctx context.Context) {
	go func() {
		for {
			err := c.Consume(ctx)
			if err != nil {
				c.logger.Error("同步事件失败", elog.Any("err", err))
			}
		}
	}()
}

func (c *ProcessEventConsumer) Consume(ctx context.Context) error {
	cm, err := c.consumer.Consume(ctx)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}
	var evt event.OrderEvent
	if err = json.Unmarshal(cm.Value, &evt); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}
	flow, err := c.workFlowSvc.Find(ctx, evt.WorkflowId)
	if err != nil {
		return err
	}

	// 启动流程引擎，配置工单与流程引擎关系ID
	engineId, err := engine.InstanceStart(flow.ProcessId, "业务申请", flow.Name, string(c.Variables(evt)))
	if err != nil {
		return err
	}

	return c.svc.RegisterProcessInstanceId(ctx, evt.Id, engineId)
}

func (c *ProcessEventConsumer) Stop(_ context.Context) error {
	return c.consumer.Close()
}

func (c *ProcessEventConsumer) Variables(evt event.OrderEvent) []byte {
	var vars []event.Variables
	for key, value := range evt.Data {
		// 判断如果浮点数类型，转换成string
		strValue := value
		valueType := reflect.TypeOf(value)
		if valueType.Kind() == reflect.Float64 {
			strValue = fmt.Sprintf("%f", value)
		}

		vars = append(vars, event.Variables{
			Key:   key,
			Value: strValue,
		})
	}
	VariablesJson, _ := json.Marshal(vars)

	return VariablesJson
}