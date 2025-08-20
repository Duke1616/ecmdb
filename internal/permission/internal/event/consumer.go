package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/menu"
	"github.com/Duke1616/ecmdb/internal/permission/internal/domain"
	"github.com/Duke1616/ecmdb/internal/permission/internal/service"
	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
)

type MenuChangeEventConsumer struct {
	svc      service.Service
	consumer mq.Consumer
	logger   *elog.Component
}

func NewMenuChangeEventConsumer(q mq.MQ, svc service.Service) (*MenuChangeEventConsumer, error) {
	groupID := "menu_change"
	consumer, err := q.Consumer(MenuChangeEventName, groupID)
	if err != nil {
		return nil, err
	}

	return &MenuChangeEventConsumer{
		consumer: consumer,
		svc:      svc,
		logger:   elog.DefaultLogger,
	}, nil
}

func (c *MenuChangeEventConsumer) Start(ctx context.Context) {
	go func() {
		for {
			err := c.Consume(ctx)
			if err != nil {
				c.logger.Error("同步事件失败", elog.Any("err", err))
			}
		}
	}()
}

func (c *MenuChangeEventConsumer) Consume(ctx context.Context) error {
	cm, err := c.consumer.Consume(ctx)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}
	var evt menu.EventMenuQueue
	if err = json.Unmarshal(cm.Value, &evt); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}

	return c.svc.MenuChangeTriggerRoleAndPolicy(ctx, evt.Action.ToUint8(), c.toDomainMenu(evt.Menu))
}

func (c *MenuChangeEventConsumer) Stop(_ context.Context) error {
	return c.consumer.Close()
}

func (c *MenuChangeEventConsumer) toDomainMenu(req menu.EventMenu) domain.Menu {
	return domain.Menu{
		Id: req.Id,
		Endpoints: slice.Map(req.Endpoints, func(idx int, src menu.EventEndpoint) domain.Endpoint {
			return domain.Endpoint{
				Path:   src.Path,
				Method: src.Method,
			}
		}),
	}
}
