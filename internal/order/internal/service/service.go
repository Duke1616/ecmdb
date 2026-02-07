package service

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/Duke1616/ecmdb/internal/order/internal/domain"
	"github.com/Duke1616/ecmdb/internal/order/internal/event"
	"github.com/Duke1616/ecmdb/internal/order/internal/repository"
	"github.com/Duke1616/ecmdb/internal/template"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
	"github.com/xen0n/go-workwx"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	// CreateBizOrder 创建业务工单
	CreateBizOrder(ctx context.Context, order domain.Order) (domain.Order, error)

	// CreateOrder 创建工单
	CreateOrder(ctx context.Context, req domain.Order) error

	// DetailByProcessInstId 根据流程实例 ID 获取工单信息
	DetailByProcessInstId(ctx context.Context, instanceId int) (domain.Order, error)

	// Detail 根据ID获取工单详情信息
	Detail(ctx context.Context, id int64) (domain.Order, error)

	// UpdateStatusByInstanceId 更新状态
	UpdateStatusByInstanceId(ctx context.Context, instanceId int, status uint8) error

	// RegisterProcessInstanceId 注册流程引擎ID
	RegisterProcessInstanceId(ctx context.Context, id int64, instanceId int) error

	// ListOrderByProcessInstanceIds 获取代办流程
	ListOrderByProcessInstanceIds(ctx context.Context, instanceIds []int) ([]domain.Order, error)

	// ListHistoryOrder 获取历史order列表
	ListHistoryOrder(ctx context.Context, userId string, offset, limit int64) ([]domain.Order, int64, error)

	// ListOrdersByUser 查看自己提交的工单
	ListOrdersByUser(ctx context.Context, userId string, offset, limit int64) ([]domain.Order, int64, error)

	// MergeOrderData 合并工单数据（原子更新）
	MergeOrderData(ctx context.Context, orderId int64, data map[string]interface{}) error

	// CreateTaskForm 记录任务快照
	CreateTaskForm(ctx context.Context, taskId int, orderId int64, fields []domain.FormValue) error

	// FindTaskFormsBatch 批量查询任务快照
	FindTaskFormsBatch(ctx context.Context, taskIds []int) (map[int][]domain.FormValue, error)
}

func (s *service) FindTaskFormsBatch(ctx context.Context, taskIds []int) (map[int][]domain.FormValue, error) {
	return s.repo.FindTaskFormsBatch(ctx, taskIds)
}

func (s *service) CreateTaskForm(ctx context.Context, taskId int, orderId int64, fields []domain.FormValue) error {
	return s.repo.CreateTaskForm(ctx, taskId, orderId, fields)
}

func (s *service) MergeOrderData(ctx context.Context, orderId int64, data map[string]interface{}) error {
	return s.repo.MergeOrderData(ctx, orderId, data)
}

type service struct {
	repo        repository.OrderRepository
	templateSvc template.Service
	producer    event.CreateProcessEventProducer
	l           *elog.Component
}

func (s *service) CreateBizOrder(ctx context.Context, order domain.Order) (domain.Order, error) {
	if err := order.Validate(); err != nil {
		return domain.Order{}, err
	}

	// 如果是告警转工单，且 Key 和 BizID 不为空，检查是否已有相同 Key 和 BizID 的进行中工单
	if order.Provide.IsAlert() && order.Key != "" && order.BizID > 0 {
		existingOrder, err := s.repo.FindByBizIdAndKey(ctx, order.BizID, order.Key, []domain.Status{domain.START, domain.PROCESS})
		if err != nil {
			s.l.Warn("查询已有工单失败",
				elog.FieldErr(err),
				elog.Int64("bizId", order.BizID),
				elog.String("key", order.Key))
		} else if existingOrder.Id > 0 {
			// 找到已有工单，返回已有工单，并发送追加告警通知
			s.l.Info("找到已有工单，追加告警",
				elog.Int64("existingOrderId", existingOrder.Id),
				elog.Int64("bizId", order.BizID),
				elog.String("key", order.Key))

			// 异步发送追加告警通知（不阻塞主流程）
			go func() {
				defer func() {
					if r := recover(); r != nil {
						s.l.Error("发送追加告警通知发生panic", elog.Any("recover", r))
					}
				}()
				if err := s.sendAppendAlertNotification(ctx, existingOrder, order); err != nil {
					s.l.Error("发送追加告警通知失败",
						elog.FieldErr(err),
						elog.Int64("orderId", existingOrder.Id))
				}
			}()

			return existingOrder, nil
		}
	}

	// 创建新工单
	bizOrder, err := s.repo.CreateBizOrder(ctx, order)
	if err != nil {
		return domain.Order{}, err
	}

	return bizOrder, s.sendGenerateFlowEvent(ctx, order, bizOrder.Id, "TODO")
}

func (s *service) Detail(ctx context.Context, id int64) (domain.Order, error) {
	return s.repo.Detail(ctx, id)
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

func NewService(repo repository.OrderRepository, templateSvc template.Service, producer event.CreateProcessEventProducer) Service {
	return &service{
		repo:        repo,
		producer:    producer,
		templateSvc: templateSvc,
		l:           elog.DefaultLogger,
	}
}

func (s *service) CreateOrder(ctx context.Context, req domain.Order) error {
	if err := req.Validate(); err != nil {
		return err
	}

	var (
		eg      errgroup.Group
		orderId int64
		dTm     template.Template
	)
	eg.Go(func() error {
		var err error
		orderId, err = s.repo.CreateOrder(ctx, req)
		return err
	})

	// TODO 这个地方我是利用了模版名称去做比对，现在想想设计的并不好
	eg.Go(func() error {
		var err error
		dTm, err = s.templateSvc.DetailTemplate(ctx, req.TemplateId)
		return err
	})
	if err := eg.Wait(); err != nil {
		return err
	}

	return s.sendGenerateFlowEvent(ctx, req, orderId, dTm.Name)
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

func (s *service) sendGenerateFlowEvent(ctx context.Context, req domain.Order,
	orderId int64, tName string) error {
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

	variables = append(variables, event.Variables{
		Key:   "template_name",
		Value: tName,
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

			// 如果是数组类型，转换成json string
			if valueType.Kind() == reflect.Slice || valueType.Kind() == reflect.Array {
				if v, err := json.Marshal(value); err == nil {
					strValue = string(v)
				}
			}

			data = append(data, event.Variables{
				Key:   key,
				Value: strValue,
			})
		}
	}

	return data, nil
}

// sendAppendAlertNotification 发送追加告警通知
// 当新告警追加到已有工单时，通知相关处理人
func (s *service) sendAppendAlertNotification(ctx context.Context, existingOrder domain.Order, newAlert domain.Order) error {
	// 如果工单还没有流程实例（流程还未启动），则无法获取任务信息，暂不发送通知
	if existingOrder.Process.InstanceId == 0 {
		s.l.Info("工单流程未启动，暂不发送追加告警通知",
			elog.Int64("orderId", existingOrder.Id))
		return nil
	}

	// TODO: 如果工单有流程实例，需要获取当前节点的任务信息，然后发送通知
	// 由于需要获取任务信息需要依赖 engine service，这里先记录日志
	// 后续可以通过事件或其他方式实现完整的通知功能
	s.l.Info("追加告警到已有工单",
		elog.Int64("orderId", existingOrder.Id),
		elog.Int("processInstanceId", existingOrder.Process.InstanceId),
		elog.Int64("newAlertBizId", newAlert.BizID),
		elog.String("newAlertKey", newAlert.Key))

	// 如果工单配置了通知信息，可以在这里发送通知
	// 目前先记录日志，后续可以根据实际需求完善
	if existingOrder.NotificationConf.TemplateID > 0 {
		s.l.Info("工单配置了通知信息，可以发送追加告警通知",
			elog.Int64("templateId", existingOrder.NotificationConf.TemplateID),
			elog.String("channel", existingOrder.NotificationConf.Channel.String()))
		// TODO: 实现具体的通知发送逻辑
		// 需要：
		// 1. 获取流程实例的当前节点任务（需要 engine service）
		// 2. 获取任务的用户信息（需要 user service）
		// 3. 调用 notification service 发送通知
	}

	return nil
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
		// 使用 ID 当作 key 进行存储
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
