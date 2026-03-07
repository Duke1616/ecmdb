package strategy

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Bunny3th/easy-workflow/workflow/model"
	engineSvc "github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/pkg/notification"
	"github.com/Duke1616/ecmdb/internal/pkg/rule"
	"github.com/Duke1616/ecmdb/internal/task"
	"github.com/Duke1616/ecmdb/internal/template"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/Duke1616/ecmdb/pkg/resolve"
	"github.com/ecodeclub/ekit/retry"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/sync/errgroup"
)

// BaseStrategy 提供所有通知策略共用的基础服务和辅助方法
type BaseStrategy struct {
	TemplateSvc template.Service
	UserSvc     user.Service
	TaskSvc     task.Service
	OrderSvc    order.Service
	EngineSvc   engineSvc.Service
	Logger      *elog.Component

	// 通用重试配置
	InitialInterval time.Duration
	MaxInterval     time.Duration
	MaxRetries      int32
}

func NewBaseStrategy(userSvc user.Service, templateSvc template.Service,
	taskSvc task.Service, orderSvc order.Service, engineSvc engineSvc.Service) BaseStrategy {
	return BaseStrategy{
		TemplateSvc:     templateSvc,
		UserSvc:         userSvc,
		TaskSvc:         taskSvc,
		OrderSvc:        orderSvc,
		EngineSvc:       engineSvc,
		Logger:          elog.DefaultLogger,
		InitialInterval: 5 * time.Second,
		MaxInterval:     15 * time.Second,
		MaxRetries:      3,
	}
}

// NotificationData 封装策略间共享的元数据
type NotificationData struct {
	WantResult map[string]interface{}
	Rules      []rule.Rule
	StartUser  user.User
	TName      string
}

// FetchRequiredData 并行获取基础通知数据
func (b *BaseStrategy) FetchRequiredData(ctx context.Context, info Info, nodes []easyflow.Node) (*NotificationData, error) {
	var data NotificationData
	errGroup, ctx := errgroup.WithContext(ctx)

	errGroup.Go(func() error {
		var err error
		data.WantResult, err = b.WantAllResult(ctx, info.InstID, nodes)
		return err
	})

	errGroup.Go(func() error {
		var err error
		data.Rules, data.TName, err = b.GetRules(ctx, info.Order)
		return err
	})

	errGroup.Go(func() error {
		var err error
		data.StartUser, err = b.UserSvc.FindByUsername(ctx, info.Order.CreateBy)
		return err
	})

	if err := errGroup.Wait(); err != nil {
		return nil, err
	}
	return &data, nil
}

// FetchTasksWithRetry 在引擎异步落库过程中重试获取任务信息
func (b *BaseStrategy) FetchTasksWithRetry(ctx context.Context, info Info) ([]model.Task, error) {
	strategy, err := retry.NewExponentialBackoffRetryStrategy(b.InitialInterval, b.MaxInterval, b.MaxRetries)
	if err != nil {
		return nil, err
	}

	for {
		d, ok := strategy.Next()
		if !ok {
			return nil, fmt.Errorf("获取执行任务超过最大重试次数")
		}

		tasks, err := b.EngineSvc.GetTasksByCurrentNodeId(ctx, info.InstID, info.CurrentNode.NodeID)
		if err == nil && len(tasks) > 0 {
			return tasks, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(d):
			continue
		}
	}
}

// PrepareCommonFields 统一解析工单数据的公示字段
func (b *BaseStrategy) PrepareCommonFields(info Info, data *NotificationData) []notification.Field {
	ruleFields := rule.GetFields(data.Rules, info.Order.Provide.ToUint8(), info.Order.Data)
	fields := slice.Map(ruleFields, func(idx int, src rule.Field) notification.Field {
		return notification.Field{
			IsShort: src.IsShort,
			Tag:     src.Tag,
			Content: src.Content,
		}
	})

	for field, value := range data.WantResult {
		fields = append(fields, notification.Field{
			IsShort: true,
			Tag:     "lark_md",
			Content: fmt.Sprintf(`**%s:**\n%v`, field, value),
		})
	}
	return fields
}

// GetRules 获取工单对应的通知规则和模板名称
func (b *BaseStrategy) GetRules(ctx context.Context, oInfo order.Order) ([]rule.Rule, string, error) {
	t, err := b.TemplateSvc.DetailTemplate(ctx, oInfo.TemplateId)
	if err != nil {
		return nil, "", err
	}

	rules, err := rule.ParseRules(t.Rules)
	if err != nil {
		return nil, "", err
	}

	return rules, t.Name, nil
}

// WantAllResult 汇总流程中所有自动化节点的执行结果
func (b *BaseStrategy) WantAllResult(ctx context.Context, instanceId int, nodes []easyflow.Node) (map[string]interface{}, error) {
	mergedResult := make(map[string]interface{})
	for _, node := range nodes {
		if node.Type != "automation" {
			continue
		}
		if result, err := b.FetchResult(ctx, instanceId, node.ID); err == nil {
			for k, v := range result {
				mergedResult[k] = v
			}
		}
	}
	return mergedResult, nil
}

// FetchResult 获取并解析单个节点的自动化执行结果
func (b *BaseStrategy) FetchResult(ctx context.Context, instanceID int, nodeID string) (map[string]interface{}, error) {
	result, err := b.TaskSvc.FindTaskResult(ctx, instanceID, nodeID)
	if err != nil {
		return nil, err
	}

	if result.WantResult == "" {
		return nil, fmt.Errorf("返回值为空, 不做任何数据处理")
	}

	var wantResult map[string]interface{}
	err = json.Unmarshal([]byte(result.WantResult), &wantResult)
	if err != nil {
		return nil, err
	}

	return wantResult, nil
}

// GetNodeProperty 获取指定节点的属性配置
func (b *BaseStrategy) GetNodeProperty(info Info, nodeID string) ([]easyflow.Node, any, error) {
	nodes := info.Nodes
	var err error
	if len(nodes) == 0 {
		nodes, err = b.UnmarshalNodes(info.Workflow)
		if err != nil {
			return nil, nil, err
		}
	}

	node, ok := slice.Find(nodes, func(src easyflow.Node) bool {
		return src.ID == nodeID
	})
	if ok {
		return nodes, node.Properties, nil
	}
	return nodes, nil, fmt.Errorf("未找到节点 %s", nodeID)
}

// GetChannel 根据流程配置获取通知渠道
func (b *BaseStrategy) GetChannel(wf workflow.Workflow) notification.Channel {
	switch wf.NotifyMethod {
	case workflow.Feishu:
		return notification.ChannelLarkCard
	case workflow.Wechat:
		return notification.ChannelWechat
	default:
		return notification.ChannelLarkCard
	}
}

// GetChatChannel 映射字符串渠道为 notification.Channel 类型
func (b *BaseStrategy) GetChatChannel(channel string) notification.Channel {
	switch channel {
	case "FEISHU", "LARK_CARD":
		return notification.ChannelLarkCard
	case "WECHAT":
		return notification.ChannelWechat
	default:
		return notification.ChannelLarkCard
	}
}

// IsGlobalNotify 校验全局通知开关
func (b *BaseStrategy) IsGlobalNotify(wf workflow.Workflow) bool {
	if !wf.IsNotify {
		b.Logger.Warn("流程全局消息通知已关闭", elog.Any("wfId", wf.Id))
		return false
	}
	return true
}

// EnrichTargets 处理运行时动态元数据注入
func (b *BaseStrategy) EnrichTargets(info Info, assignees []easyflow.Assignee) []resolve.Target {
	return slice.Map(assignees, func(idx int, src easyflow.Assignee) resolve.Target {
		values := src.Values
		switch src.Rule {
		case easyflow.LEADER, easyflow.MAIN_LEADER, easyflow.FOUNDER:
			if len(values) == 0 {
				values = []string{info.Order.CreateBy}
			}
		case easyflow.TEMPLATE:
			if len(values) > 0 {
				field := values[0]
				if val, ok := info.Order.Data[field]; ok {
					values = b.extractUsernamesFromField(field, val)
				}
			}
		}

		return resolve.Target{
			Type:   string(src.Rule),
			Values: values,
		}
	})
}

// extractUsernamesFromField 从模版字段中提取用户名列表，支持多种类型
func (b *BaseStrategy) extractUsernamesFromField(fieldName string, value interface{}) []string {
	switch v := value.(type) {
	case []string:
		return v
	case string:
		return []string{v}
	case []interface{}:
		// 处理 JSON 反序列化后的数组
		return slice.FilterMap(v, func(idx int, item interface{}) (string, bool) {
			if s, ok := item.(string); ok {
				return s, true
			}
			b.Logger.Warn("模版字段数组元素类型不支持",
				elog.String("field", fieldName),
				elog.Any("element_type", fmt.Sprintf("%T", item)))
			return "", false
		})
	default:
		b.Logger.Warn("模版字段类型不支持",
			elog.String("field", fieldName),
			elog.Any("type", fmt.Sprintf("%T", value)),
			elog.Any("value", value))
		return nil
	}
}

func (b *BaseStrategy) UnmarshalNodes(wf workflow.Workflow) ([]easyflow.Node, error) {
	var nodes []easyflow.Node
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &nodes,
		TagName: "json",
	})
	if err != nil {
		return nil, err
	}

	if err = decoder.Decode(wf.FlowData.Nodes); err != nil {
		return nil, err
	}

	return nodes, nil
}

// ConvertRuleFields 将 rule.Field 转换为 notification.Field
func (b *BaseStrategy) ConvertRuleFields(fields []rule.Field) []notification.Field {
	return slice.Map(fields, func(idx int, src rule.Field) notification.Field {
		return notification.Field{
			IsShort: src.IsShort,
			Tag:     src.Tag,
			Content: src.Content,
		}
	})
}

// BuildWantResultFields 构建自动化任务结果字段
func (b *BaseStrategy) BuildWantResultFields(wantResult map[string]interface{}) []notification.Field {
	fields := make([]notification.Field, 0, len(wantResult))
	for k, v := range wantResult {
		fields = append(fields, notification.Field{
			IsShort: true,
			Tag:     "lark_md",
			Content: fmt.Sprintf("**%s:**\n%v", k, v),
		})
	}
	return fields
}

// AnalyzeUsers 将用户列表转换为 username -> feishu_user_id 的映射
func (b *BaseStrategy) AnalyzeUsers(users []user.User) map[string]string {
	return slice.ToMapV(users, func(u user.User) (string, string) {
		return u.Username, u.FeishuInfo.UserId
	})
}

func containsAutoNotifyMethod(methods []int64, target int64) bool {
	return slice.Contains(methods, target)
}
