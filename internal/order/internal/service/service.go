package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/order/internal/domain"
	"github.com/Duke1616/ecmdb/internal/order/internal/event"
	"github.com/Duke1616/ecmdb/internal/order/internal/repository"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
	"github.com/xen0n/go-workwx"
	"golang.org/x/sync/errgroup"
	"reflect"
	"strconv"
	"strings"
)

type Service interface {
	CreateOrder(ctx context.Context, req domain.Order) error
	DetailByProcessInstId(ctx context.Context, instanceId int) (domain.Order, error)
	UpdateStatusByInstanceId(ctx context.Context, instanceId int, status uint8) error
	// RegisterProcessInstanceId 注册流程引擎ID
	RegisterProcessInstanceId(ctx context.Context, id int64, instanceId int) error
	// ListOrderByProcessInstanceIds 获取代办流程
	ListOrderByProcessInstanceIds(ctx context.Context, instanceIds []int) ([]domain.Order, error)

	// ListHistoryOrder 获取历史order列表
	ListHistoryOrder(ctx context.Context, userId string, offset, limit int64) ([]domain.Order, int64, error)

	// ListOrdersByUser 查看自己提交的工单
	ListOrdersByUser(ctx context.Context, userId string, offset, limit int64) ([]domain.Order, int64, error)
}

type service struct {
	repo     repository.OrderRepository
	producer event.CreateProcessEventProducer
	l        *elog.Component
}

func (s *service) ListOrdersByUser(ctx context.Context, userId string, offset, limit int64) ([]domain.Order, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Order
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.ListOrder(ctx, userId, []int{domain.PROCESS.ToInt()}, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.CountOrder(ctx, userId, []int{domain.PROCESS.ToInt()})
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}

func (s *service) UpdateStatusByInstanceId(ctx context.Context, instanceId int, status uint8) error {
	return s.repo.UpdateStatusByInstanceId(ctx, instanceId, status)
}

func (s *service) RegisterProcessInstanceId(ctx context.Context, id int64, instanceId int) error {
	return s.repo.RegisterProcessInstanceId(ctx, id, instanceId, domain.PROCESS.ToUint8())
}

func NewService(repo repository.OrderRepository, producer event.CreateProcessEventProducer) Service {
	return &service{
		repo:     repo,
		producer: producer,
		l:        elog.DefaultLogger,
	}
}

func (s *service) CreateOrder(ctx context.Context, req domain.Order) error {
	orderId, err := s.repo.CreateOrder(ctx, req)
	if err != nil {
		return err
	}

	return s.sendGenerateFlowEvent(ctx, req, orderId)
}

func (s *service) ListOrderByProcessInstanceIds(ctx context.Context, instanceIds []int) ([]domain.Order, error) {
	return s.repo.ListOrderByProcessInstanceIds(ctx, instanceIds)
}

func (s *service) DetailByProcessInstId(ctx context.Context, instanceId int) (domain.Order, error) {
	return s.repo.DetailByProcessInstId(ctx, instanceId)
}

func (s *service) ListHistoryOrder(ctx context.Context, userId string, offset, limit int64) (
	[]domain.Order, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Order
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.ListOrder(ctx, userId, []int{domain.END.ToInt(), domain.WITHDRAW.ToInt()}, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.CountOrder(ctx, userId, []int{domain.END.ToInt(), domain.WITHDRAW.ToInt()})
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}

func (s *service) sendGenerateFlowEvent(ctx context.Context, req domain.Order, orderId int64) error {
	if req.Data == nil {
		req.Data = make(map[string]interface{})
	}

	variables, err := s.variables(req)
	if err != nil {
		return err
	}

	// 把工单ID传递过去，通过Event事件绑定流程ID
	variables = append(variables, event.Variables{
		Key:   "order_id",
		Value: strconv.FormatInt(orderId, 10),
	})

	data, err := json.Marshal(variables)
	if err != nil {
		return err
	}

	evt := event.OrderEvent{
		Id:         orderId,
		Provide:    event.Provide(req.Provide),
		WorkflowId: req.WorkflowId,
		Data:       req.Data,
		Variables:  string(data),
	}

	err = s.producer.Produce(ctx, evt)
	if err != nil {
		// 要做好监控和告警
		s.l.Error("发送创建流程事件失败",
			elog.FieldErr(err),
			elog.Any("evt", evt))
	}

	return nil
}

func (s *service) variables(req domain.Order) ([]event.Variables, error) {
	var data []event.Variables
	data = append(data, event.Variables{
		Key:   "starter",
		Value: req.CreateBy,
	})

	switch req.Provide {
	case domain.WECHAT:
		oaData, err := wechatOaData(req.Data)
		if err != nil {
			return nil, err
		}

		data = convert(data, oaData)
	case domain.SYSTEM:
		for key, value := range req.Data {
			// 判断如果浮点数类型，转换成string
			strValue := value
			valueType := reflect.TypeOf(value)
			if valueType.Kind() == reflect.Float64 {
				strValue = fmt.Sprintf("%f", value)
			}

			data = append(data, event.Variables{
				Key:   key,
				Value: strValue,
			})
		}
	}

	return data, nil
}

func wechatOaData(data map[string]interface{}) (workwx.OAApprovalDetail, error) {
	wechatOaJson, err := json.Marshal(data)
	if err != nil {
		return workwx.OAApprovalDetail{}, nil
	}

	var wechatOaDetail workwx.OAApprovalDetail
	err = json.Unmarshal(wechatOaJson, &wechatOaDetail)
	if err != nil {
		return workwx.OAApprovalDetail{}, err
	}

	return wechatOaDetail, nil
}

func convert(data []event.Variables, oaData workwx.OAApprovalDetail) []event.Variables {
	for _, contents := range oaData.ApplyData.Contents {
		id := strings.Split(contents.ID, "-")
		key := id[1]

		switch contents.Control {
		case "Selector":
			switch contents.Value.Selector.Type {
			case "single":
				data = append(data, event.Variables{
					Key:   key,
					Value: contents.Value.Selector.Options[0].Value[0].Text,
				})
			case "multi":
				value := slice.Map(contents.Value.Selector.Options, func(idx int, src workwx.OAContentSelectorOption) string {
					return src.Value[0].Text
				})

				data = append(data, event.Variables{
					Key:   key,
					Value: value,
				})
			}

		case "Textarea":
			data = append(data, event.Variables{
				Key:   key,
				Value: contents.Value.Text,
			})
		case "default":
			fmt.Println("不符合筛选规则")
		}
	}

	return data
}
