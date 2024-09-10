package easyflow

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	engineSvc "github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/order"
	templateSvc "github.com/Duke1616/ecmdb/internal/template"
	"github.com/Duke1616/enotify/notify"
	"github.com/Duke1616/enotify/notify/feishu"
	"github.com/Duke1616/enotify/notify/feishu/card"
	"github.com/Duke1616/enotify/template"
	"github.com/ecodeclub/ekit/retry"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"sync"
	"time"
)

type NotificationService interface {
	Send(ctx context.Context, instanceId int, userIDs []string) (bool, error)
}

type Notify struct {
	engineSvc   engineSvc.Service
	templateSvc templateSvc.Service
	orderSvc    order.Service
	tmpl        *template.Template
	nc          notify.Notifier[*larkim.CreateMessageReq]
	logger      *elog.Component

	initialInterval time.Duration
	maxInterval     time.Duration
	maxRetries      int32
}

func NewNotify(engineSvc engineSvc.Service, templateSvc templateSvc.Service, orderSvc order.Service,
	lark *lark.Client) (*Notify, error) {
	tmpl, err := template.FromGlobs([]string{})
	if err != nil {
		return nil, err
	}

	nc, err := feishu.NewFeishuNotifyByClient(lark)
	if err != nil {
		return nil, err
	}

	return &Notify{
		engineSvc:       engineSvc,
		templateSvc:     templateSvc,
		orderSvc:        orderSvc,
		logger:          elog.DefaultLogger,
		tmpl:            tmpl,
		nc:              nc,
		initialInterval: 10 * time.Second,
		maxRetries:      int32(3),
		maxInterval:     30 * time.Second,
	}, nil
}

func (n *Notify) builder(title string, fields []card.Field, cardVal []card.Value) notify.NotifierWrap {
	return notify.WrapNotifierDynamic(n.nc, func() (notify.BasicNotificationMessage[*larkim.CreateMessageReq], error) {
		return feishu.NewFeishuMessage(
			"user_id", "bcegag66",
			feishu.NewFeishuCustomCard(n.tmpl, "feishu-card-callback",
				card.NewApprovalCardBuilder().
					SetToTitle(title).
					SetToFields(fields).
					SetToCallbackValue(cardVal).Build(),
			),
		), nil
	})
}

func (n *Notify) Send(ctx context.Context, instanceId int, userIDs []string) (bool, error) {
	// 返回用户提交信息
	fields, title, err := n.getFields(ctx, instanceId)
	if err != nil {
		return false, err
	}

	// 只有当 Event 结束才能正确获取到 TaskId 信息，放到 Go Routine 异步运行, 通过重试机制获取到数据
	go func() {
		strategy, err := retry.NewExponentialBackoffRetryStrategy(n.initialInterval, n.maxInterval, n.maxRetries)
		if err != nil {
			return
		}

		var tasks []model.Task
		for {
			d, ok := strategy.Next()
			if !ok {
				n.logger.Error("处理执行任务超过最大重试次数",
					elog.Any("error", err),
					elog.Any("instId", instanceId),
				)
				break
			}

			tasks, err = n.engineSvc.GetTasksByInstUsers(ctx, instanceId, userIDs)
			if err != nil || len(tasks) == 0 {
				time.Sleep(d)
				continue
			}

			break
		}

		// 继续组合消息
		var messages []notify.NotifierWrap
		for _, ts := range tasks {
			cardVal := []card.Value{
				{
					Key:   "task_id",
					Value: ts.TaskID,
				},
			}

			messages = append(messages, n.builder(title, fields, cardVal))
		}

		// 异步发送消息
		if ok, err := n.send(ctx, messages); err != nil || !ok {
			n.logger.Warn("发送消息失败",
				elog.Any("error", err),
				elog.Any("user_ids", userIDs),
			)
		}
	}()

	return true, nil
}

func (n *Notify) send(ctx context.Context, notifyWrap []notify.NotifierWrap) (bool, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstError error
	success := true

	// 使用 goroutines 发送消息
	for _, msg := range notifyWrap {
		wg.Add(1)
		nw := msg
		go func(m notify.NotifierWrap) {
			defer wg.Done()

			ok, err := nw.Send(ctx)
			if err != nil {
				mu.Lock() // 锁定访问共享资源
				if firstError == nil {
					firstError = err // 记录第一个出现的错误
				}
				success = false
				mu.Unlock()
			}

			if !ok {
				mu.Lock()
				if firstError == nil {
					firstError = fmt.Errorf("消息发送失败")
				}
				success = false
				mu.Unlock()
			}
		}(msg)
	}

	// 等待所有 goroutine 完成
	wg.Wait()

	if firstError != nil {
		return false, firstError
	}
	return success, nil
}

func (n *Notify) getFields(ctx context.Context, InstanceId int) ([]card.Field, string, error) {
	// 获取工单详情信息
	o, err := n.orderSvc.DetailByProcessInstId(ctx, InstanceId)
	if err != nil {
		return nil, "", err
	}

	// 获取模版详情信息
	t, err := n.templateSvc.DetailTemplate(ctx, o.TemplateId)
	if err != nil {
		return nil, "", err
	}

	rules, err := parseRules(t.Rules)
	if err != nil {
		return nil, "", err
	}
	ruleMap := slice.ToMap(rules, func(element Rule) string {
		return element.Field
	})

	// 拼接消息体
	var fields []card.Field
	num := 1
	for field, value := range o.Data {
		title := field
		val, ok := ruleMap[field]
		if ok {
			title = val.Title
		}

		fields = append(fields, card.Field{
			IsShort: true,
			Tag:     "lark_md",
			Content: fmt.Sprintf(`**%s:**\n%v`, title, value),
		})

		if num%2 == 0 {
			fields = append(fields, card.Field{
				IsShort: false,
				Tag:     "lark_md",
				Content: "",
			})
		}

		num++
	}

	return fields, o.TemplateName, nil
}

func parseRules(ruleData interface{}) ([]Rule, error) {
	var rules []Rule
	rulesJson, err := json.Marshal(ruleData)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(rulesJson, &rules)
	return rules, err
}
