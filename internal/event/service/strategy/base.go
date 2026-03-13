package strategy

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"sort"
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

// Service 提供所有通知策略共用的基础服务和辅助方法
//
//go:generate mockgen -source=./base.go -package=strategymocks -destination=../../mocks/strategy.mock.go -typed Service
type Service interface {
	// FetchRequiredData 并行获取基础通知数据
	FetchRequiredData(ctx context.Context, info Info, nodes []easyflow.Node) (*NotificationData, error)
	// FetchTasksWithRetry 在引擎异步落库过程中重试获取任务信息
	FetchTasksWithRetry(ctx context.Context, info Info) ([]model.Task, error)
	// GetNodeProperty 获取指定节点的属性配置
	GetNodeProperty(info Info, nodeID string) ([]easyflow.Node, any, error)
	// IsGlobalNotify 校验全局通知开关
	IsGlobalNotify(wf workflow.Workflow) bool
	// EnrichTargets 处理运行时动态元数据注入
	EnrichTargets(info Info, assignees []easyflow.Assignee) []resolve.Target
	// PrepareCommonFields 统一解析工单数据的公示字段
	PrepareCommonFields(info Info, data *NotificationData) []notification.Field

	// ResolveAssignees 解析审批人并同步到节点
	ResolveAssignees(ctx context.Context, info *Info, assignees []easyflow.Assignee) ([]user.User, error)

	// SafeGo 安全执行异步任务 (带 panic 恢复)
	SafeGo(ctx context.Context, timeout time.Duration, fn func(ctx context.Context))

	// PassTask 推进任务
	PassTask(ctx context.Context, taskId int, remark string) error
	// FindTaskForms 获取工单表单数据
	FindTaskForms(ctx context.Context, orderId int64) ([]order.FormValue, error)

	Logger() *elog.Component
}

type service struct {
	templateSvc     template.Service
	userSvc         user.Service
	taskSvc         task.Service
	orderSvc        order.Service
	engineSvc       engineSvc.Service
	assigneeService resolve.Engine
	logger          *elog.Component

	// 通用重试配置
	InitialInterval time.Duration
	MaxInterval     time.Duration
	MaxRetries      int32
}

func NewService(userSvc user.Service, templateSvc template.Service,
	taskSvc task.Service, orderSvc order.Service, engineSvc engineSvc.Service,
	assigneeService resolve.Engine) Service {
	return &service{
		templateSvc:     templateSvc,
		userSvc:         userSvc,
		taskSvc:         taskSvc,
		orderSvc:        orderSvc,
		engineSvc:       engineSvc,
		assigneeService: assigneeService,
		logger:          elog.DefaultLogger,
		InitialInterval: 5 * time.Second,
		MaxInterval:     15 * time.Second,
		MaxRetries:      3,
	}
}

func (s *service) PassTask(ctx context.Context, taskId int, remark string) error {
	return s.engineSvc.Pass(ctx, taskId, remark)
}

func (s *service) FindTaskForms(ctx context.Context, orderId int64) ([]order.FormValue, error) {
	return s.orderSvc.FindTaskFormsByOrderID(ctx, orderId)
}

func (s *service) ResolveAssignees(ctx context.Context, info *Info, assignees []easyflow.Assignee) ([]user.User, error) {
	targets := s.EnrichTargets(*info, assignees)
	users, err := s.assigneeService.Resolve(ctx, targets)
	if err != nil {
		nodeID := ""
		if info.CurrentNode != nil {
			nodeID = info.CurrentNode.NodeID
		}
		return nil, fmt.Errorf("解析审批人失败 [Node: %s, Workflow: %s]: %w",
			nodeID, info.Workflow.Name, err)
	}

	// 自动同步到当前节点的审批人列表 (用于流程推进)
	if info.CurrentNode != nil {
		info.CurrentNode.UserIDs = slice.Map(users, func(idx int, u user.User) string {
			return u.Username
		})
	}
	return users, nil
}

func (s *service) SafeGo(ctx context.Context, timeout time.Duration, fn func(ctx context.Context)) {
	go func() {
		sendCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		defer func() {
			if r := recover(); r != nil {
				s.Logger().Error("异步任务执行发生 panic",
					elog.Any("recover", r),
					elog.FieldStack(debug.Stack()))
			}
		}()
		fn(sendCtx)
	}()
}

func (s *service) Logger() *elog.Component {
	return s.logger
}

// NotificationData 封装策略间共享的元数据
type NotificationData struct {
	WantResult map[string]interface{}
	Rules      []rule.Rule
	StartUser  user.User
	TName      string
}

// FetchRequiredData 并行获取基础通知数据
func (s *service) FetchRequiredData(ctx context.Context, info Info, nodes []easyflow.Node) (*NotificationData, error) {
	var data NotificationData
	errGroup, ctx := errgroup.WithContext(ctx)

	errGroup.Go(func() error {
		var err error
		data.WantResult, err = s.wantAllResult(ctx, info.InstID, nodes)
		return err
	})

	errGroup.Go(func() error {
		var err error
		data.Rules, data.TName, err = s.getRules(ctx, info.Order)
		return err
	})

	errGroup.Go(func() error {
		var err error
		data.StartUser, err = s.userSvc.FindByUsername(ctx, info.Order.CreateBy)
		return err
	})

	if err := errGroup.Wait(); err != nil {
		return nil, err
	}
	return &data, nil
}

// FetchTasksWithRetry 在引擎异步落库过程中重试获取任务信息
func (s *service) FetchTasksWithRetry(ctx context.Context, info Info) ([]model.Task, error) {
	strategy, err := retry.NewExponentialBackoffRetryStrategy(s.InitialInterval, s.MaxInterval, s.MaxRetries)
	if err != nil {
		return nil, err
	}

	for {
		d, ok := strategy.Next()
		if !ok {
			return nil, fmt.Errorf("获取执行任务超过最大重试次数")
		}

		tasks, taskErr := s.engineSvc.GetTasksByCurrentNodeId(ctx, info.InstID, info.CurrentNode.NodeID)
		if taskErr == nil && len(tasks) > 0 {
			s.logger.Debug("成功获取到节点任务信息",
				elog.String("nodeId", info.CurrentNode.NodeID),
				elog.Int("taskCount", len(tasks)))
			return tasks, nil
		}

		s.logger.Debug("尚未查询到节点任务，准备重试",
			elog.Int("instId", info.InstID),
			elog.String("nodeId", info.CurrentNode.NodeID),
			elog.Duration("nextRetry", d))

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(d):
			continue
		}
	}
}

// PrepareCommonFields 统一解析工单数据的公示字段
func (s *service) PrepareCommonFields(info Info, data *NotificationData) []notification.Field {
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

// getRules 获取工单对应的通知规则和模板名称
func (s *service) getRules(ctx context.Context, oInfo order.Order) ([]rule.Rule, string, error) {
	t, err := s.templateSvc.DetailTemplate(ctx, oInfo.TemplateId)
	if err != nil {
		return nil, "", err
	}

	rules, err := rule.ParseRules(t.Rules)
	if err != nil {
		return nil, "", err
	}

	return rules, t.Name, nil
}

// wantAllResult 汇总流程中所有自动化节点的执行结果
func (s *service) wantAllResult(ctx context.Context, instanceId int, nodes []easyflow.Node) (map[string]interface{}, error) {
	mergedResult := make(map[string]interface{})
	for _, node := range nodes {
		if node.Type != "automation" {
			continue
		}
		if result, err := s.fetchResult(ctx, instanceId, node.ID); err == nil {
			for k, v := range result {
				mergedResult[k] = v
			}
		}
	}
	return mergedResult, nil
}

// fetchResult 获取并解析单个节点的自动化执行结果
func (s *service) fetchResult(ctx context.Context, instanceID int, nodeID string) (map[string]interface{}, error) {
	result, err := s.taskSvc.FindTaskResult(ctx, instanceID, nodeID)
	if err != nil {
		return nil, err
	}

	if result.WantResult == "" {
		return nil, fmt.Errorf("返回值为空, 不做任何数据处理")
	}

	var data map[string]interface{}
	if err = json.Unmarshal([]byte(result.WantResult), &data); err != nil {
		return nil, err
	}
	return data, nil
}

// GetNodeProperty 获取指定节点的属性配置
func (s *service) GetNodeProperty(info Info, nodeID string) ([]easyflow.Node, any, error) {
	nodes := info.Nodes
	var err error
	if len(nodes) == 0 {
		nodes, err = UnmarshalNodes(info.Workflow)
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

// GetChatChannel 映射字符串渠道为 notification.Channel 类型
func GetChatChannel(channel string) notification.Channel {
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
func (s *service) IsGlobalNotify(wf workflow.Workflow) bool {
	if !wf.IsNotify {
		s.logger.Warn("流程全局消息通知已关闭", elog.Any("wfId", wf.Id))
		return false
	}
	return true
}

// EnrichTargets 处理运行时动态元数据注入
func (s *service) EnrichTargets(info Info, assignees []easyflow.Assignee) []resolve.Target {
	return slice.Map(assignees, func(idx int, src easyflow.Assignee) resolve.Target {
		if src.Rule == "" {
			s.logger.Warn("发现未定义的审批规则类型",
				elog.String("nodeId", info.CurrentNode.NodeID),
				elog.Any("assignee", src))
		}
		values := src.Values
		switch src.Rule {
		case easyflow.LEADER, easyflow.MAIN_LEADER, easyflow.FOUNDER:
			if len(values) == 0 {
				values = []string{info.Order.CreateBy}
			}
		case easyflow.TEMPLATE:
			var usernames []string
			for _, field := range values {
				if val, ok := info.Order.Data[field]; ok {
					usernames = append(usernames, s.extractUsernamesFromField(field, val)...)
				}
			}
			values = usernames
		}

		return resolve.Target{
			Type:   string(src.Rule),
			Values: values,
		}
	})
}

// extractUsernamesFromField 从模版字段中提取用户名列表，支持多种类型
func (s *service) extractUsernamesFromField(fieldName string, value interface{}) []string {
	switch v := value.(type) {
	case []string:
		return v
	case string:
		return []string{v}
	case []interface{}:
		// 处理 JSON 反序列化后的数组
		return slice.FilterMap(v, func(idx int, item interface{}) (string, bool) {
			if str, ok := item.(string); ok {
				return str, true
			}
			s.logger.Warn("模版字段数组元素类型不支持",
				elog.String("field", fieldName),
				elog.Any("element_type", fmt.Sprintf("%T", item)))
			return "", false
		})
	default:
		s.logger.Warn("模版字段类型不支持",
			elog.String("field", fieldName),
			elog.Any("type", fmt.Sprintf("%T", value)),
			elog.Any("value", value))
		return nil
	}
}

func UnmarshalNodes(wf workflow.Workflow) ([]easyflow.Node, error) {
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
func ConvertRuleFields(fields []rule.Field) []notification.Field {
	return slice.Map(fields, func(idx int, src rule.Field) notification.Field {
		return notification.Field{
			IsShort: src.IsShort,
			Tag:     src.Tag,
			Content: src.Content,
		}
	})
}

// BuildWantResultFields 构建自动化任务结果字段
func BuildWantResultFields(wantResult map[string]interface{}) []notification.Field {
	keys := make([]string, 0, len(wantResult))
	for k := range wantResult {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var fields []notification.Field
	for _, k := range keys {
		fields = append(fields, notification.Field{
			IsShort: true,
			Tag:     "lark_md",
			Content: fmt.Sprintf("**%s:**\n%v", k, wantResult[k]),
		})
	}
	return notification.AddRowSpacers(fields)
}

// RecipientMap 封装用户及其通知接收 ID 的映射
type RecipientMap struct {
	users   map[string]user.User
	channel notification.Channel
}

func NewRecipientMap(users []user.User, channel notification.Channel) RecipientMap {
	return RecipientMap{
		users:   slice.ToMap(users, func(u user.User) string { return u.Username }),
		channel: channel,
	}
}

// GetID 根据用户名和渠道获取对应的 ID (Feishu UserId 或 Wechat UserId)
// GetID 根据用户名获取对应的通知 ID
func (rm RecipientMap) GetID(username string) string {
	u, ok := rm.users[username]
	if !ok {
		return ""
	}

	switch rm.channel {
	case notification.ChannelWechat:
		return u.WechatInfo.UserId
	default:
		return u.FeishuInfo.UserId
	}
}

// GetIDs 获取所有已解析用户的通知 ID 列表
func (rm RecipientMap) GetIDs() []string {
	ids := make([]string, 0, len(rm.users))
	for _, u := range rm.users {
		var id string
		switch rm.channel {
		case notification.ChannelWechat:
			id = u.WechatInfo.UserId
		default:
			id = u.FeishuInfo.UserId
		}
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}
