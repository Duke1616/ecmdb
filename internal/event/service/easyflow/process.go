package easyflow

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	engineSvc "github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/event/producer"
	"github.com/Duke1616/ecmdb/internal/event/service/strategy"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/task"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
	"golang.org/x/sync/errgroup"

	"log"
)

const (
	SystemPass           = 3
	SystemReject         = 4
	SystemSkipped        = 5 // 条件不满足，系统自动跳过
	SystemPassComment    = "其余节点审批通过，系统判定无法继续审批"
	SystemRejectComment  = "其余节点进行驳回，系统判定无法继续审批"
	SystemSkippedComment = "条件不满足，系统自动跳过"
	SysAutoUser          = "sys_auto"
	SysProxyNodeName     = "系统代理流转"
)

// ProcessEvent easy-workflow 流程引擎事件处理
type ProcessEvent struct {
	strategy    strategy.SendStrategy
	producer    producer.OrderStatusModifyEventProducer
	taskSvc     task.Service
	orderSvc    order.Service
	engineSvc   engineSvc.Service
	workflowSvc workflow.Service
	logger      *elog.Component
}

func NewProcessEvent(producer producer.OrderStatusModifyEventProducer, engineSvc engineSvc.Service,
	taskSvc task.Service, orderSvc order.Service, workflowSvc workflow.Service,
	strategy strategy.SendStrategy) (*ProcessEvent, error) {

	return &ProcessEvent{
		logger:      elog.DefaultLogger,
		workflowSvc: workflowSvc,
		engineSvc:   engineSvc,
		taskSvc:     taskSvc,
		producer:    producer,
		strategy:    strategy,
		orderSvc:    orderSvc,
	}, nil
}

// EventStart 节点启动事件
func (e *ProcessEvent) EventStart(instID int, node *model.Node, prevNode model.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	fCtx, err := e.LoadContext(ctx, instID, node)
	if err != nil {
		return err
	}

	// 1. 发送开始通知
	if err = e.dispatchNotify(ctx, fCtx, strategy.Start); err != nil {
		e.logger.Error("【EventStart】消息通知分发失败", elog.FieldErr(err), elog.Int("instID", instID))
	}

	// 2. 绑定工单与流程实例
	return e.orderSvc.RegisterProcessInstanceId(ctx, fCtx.Order.Id, instID)
}

// EventAutomation 自动化任务处理：同步创建任务并受 Context 超时控制
func (e *ProcessEvent) EventAutomation(instID int, node *model.Node, prevNode model.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	orderId, err := e.engineSvc.GetOrderIdByVariable(ctx, instID)
	if err != nil {
		return err
	}

	orderID, _ := strconv.ParseInt(orderId, 10, 64)
	_, err = e.taskSvc.CreateTask(ctx, orderID, instID, node.NodeID)
	if err != nil {
		e.logger.Error("创建自动化任务失败", elog.Int("instID", instID), elog.FieldErr(err))
	}

	return err
}

// EventChatGroup 群通知节点事件：发送消息后自动推进
func (e *ProcessEvent) EventChatGroup(instID int, node *model.Node, prevNode model.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	fCtx, err := e.LoadContext(ctx, instID, node)
	if err != nil {
		return err
	}

	// 1. 发送群通知
	if err = e.dispatchNotify(ctx, fCtx, strategy.ChatGroup); err != nil {
		e.logger.Error("【EventChatGroup】群消息发送失败", elog.FieldErr(err), elog.Int("instID", instID))
	}

	// 2. 自动 Pass 节点推进流程
	go e.autoPassNode(instID, node.NodeID, "ChatGroup Auto Pass")
	return nil
}

// EventEnd 节点结束事件
func (e *ProcessEvent) EventEnd(instID int, node *model.Node, prevNode model.Node) error {
	processName, err := engine.GetProcessNameByInstanceID(instID)
	if err != nil {
		return err
	}

	e.logger.Info("节点结束了", elog.Any("processName", processName))
	log.Printf("--------流程[%s]节点[%s]结束-------", processName, node.NodeName)
	return nil
}

// EventClose 流程结束，修改 Order 状态为已完成
// Deprecated 废弃 不再通过 Kafka 修改状态，使用 EventNotify 直接调用接口进行修改
func (e *ProcessEvent) EventClose(instID int, node *model.Node, prevNode model.Node) error {
	evt := producer.OrderStatusModifyEvent{
		ProcessInstanceId: instID,
		Status:            producer.END,
	}

	err := e.producer.Produce(context.Background(), evt)
	if err != nil {
		// 要做好监控和告警
		e.logger.Error("发送修改 Order 事件失败",
			elog.FieldErr(err),
			elog.Any("evt", evt))
	}
	return nil
}

// EventNotify 流程节点（用户/自动化）通知事件
func (e *ProcessEvent) EventNotify(instID int, node *model.Node, prevNode model.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	fCtx, err := e.LoadContext(ctx, instID, node)
	if err != nil {
		return err
	}

	// 1. 处理结束节点：关闭工单
	if node.NodeType == model.EndNode {
		if err = e.orderSvc.UpdateStatusByInstanceId(ctx, instID, order.EndProcess.ToUint8()); err != nil {
			e.logger.Error("EventNotify 关闭工单失败：", elog.FieldErr(err), elog.Int("instID", instID))
		}
		return nil
	}

	// 2. 处理系统代理节点
	if len(node.UserIDs) > 0 && node.UserIDs[0] == SysAutoUser {
		go e.autoPassNode(instID, node.NodeID, "Sys Auto Pass")
		return nil
	}

	// 3. 处理通知派发
	nodeName := strategy.User
	if len(node.UserIDs) == 1 && node.UserIDs[0] == "automation" {
		nodeName = strategy.Automation
	}

	if err = e.dispatchNotify(ctx, fCtx, nodeName); err != nil {
		e.logger.Error("【EventNotify】消息发送失败", elog.FieldErr(err), elog.Int("instID", instID))
	}

	return nil
}

// EventCarbonCopy 抄送节点事件
func (e *ProcessEvent) EventCarbonCopy(instID int, node *model.Node, prevNode model.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	fCtx, err := e.LoadContext(ctx, instID, node)
	if err != nil {
		return err
	}

	if err = e.dispatchNotify(ctx, fCtx, strategy.CarbonCopy); err != nil {
		e.logger.Error("【EventCarbonCopy】消息发送失败", elog.FieldErr(err), elog.Int("instID", instID))
	}

	return nil
}

// dispatchNotify 统一的消息派发模版
func (e *ProcessEvent) dispatchNotify(ctx context.Context, fCtx *strategy.FlowContext, name strategy.NodeName) error {
	_, err := e.strategy.Send(ctx, strategy.Info{
		NodeName:    name,
		FlowContext: *fCtx,
	})
	return err
}

// LoadContext 并发加载流程运行所需的元数据上下文
func (e *ProcessEvent) LoadContext(ctx context.Context, instID int, node *model.Node) (*strategy.FlowContext, error) {
	var (
		eg        errgroup.Group
		orderInfo order.Order
		inst      engineSvc.Instance
	)

	// 并发获取基础信息
	eg.Go(func() error {
		orderIdStr, err := e.engineSvc.GetOrderIdByVariable(ctx, instID)
		if err != nil {
			return err
		}
		id, _ := strconv.ParseInt(orderIdStr, 10, 64)
		orderInfo, err = e.orderSvc.Detail(ctx, id)
		return err
	})

	eg.Go(func() error {
		var err error
		inst, err = e.engineSvc.GetInstanceByID(ctx, instID)
		return err
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// 二次加载获取流程定义快照
	wf, err := e.workflowSvc.FindInstanceFlow(ctx, orderInfo.WorkflowId, inst.ProcID, inst.ProcVersion)
	if err != nil {
		return nil, err
	}

	return &strategy.FlowContext{
		InstID:      instID,
		Order:       orderInfo,
		Workflow:    wf,
		Instance:    inst,
		CurrentNode: node,
	}, nil
}

// EventTaskInclusionNodePass 用户任务并行包容处理事件
// 当处于并行 或 包容网关的时候，其中一个节点驳回，其余并行节点并不会修改状态
func (e *ProcessEvent) EventTaskInclusionNodePass(taskID int, node *model.Node, prevNode model.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	e.logger.Info("包含网关-获取当前节点", elog.Any("Node名称", node.NodeName),
		elog.Any("Node节点", node.NodeID))

	taskNum, passNum, rejectNum, err := engine.TaskNodeStatus(taskID)
	e.logger.Info("包含网关-处理节点状态系统自动变更", elog.Any("任务ID", taskID),
		elog.Any("Node名称", prevNode.NodeName),
		elog.Any("Node节点", prevNode.NodeID),
		elog.Any("任务数量", taskNum),
		elog.Any("通过数量", passNum),
		elog.Any("驳回数量", rejectNum))

	if err != nil {
		return err
	}

	// 查看任务详情信息
	t, err := engine.GetTaskInfo(taskID)
	if err != nil {
		return err
	}

	// 如果是代理节点，需要查询代理节点的上级
	nodeId, err := e.getTargetNodeID(ctx, t.ProcInstID, prevNode.NodeID, node)
	if err != nil {
		return err
	}

	e.logger.Info("包含网关-触发处理", elog.String("nodeId", nodeId))

	// 但凡是有驳回，一率进行处理
	if rejectNum > 0 {
		return e.engineSvc.UpdateIsFinishedByPreNodeId(ctx, t.ProcInstID, nodeId, SystemReject, SystemRejectComment)
	}

	// 如果不是会签节点，直接修改所有
	if t.IsCosigned != 1 {
		return e.engineSvc.UpdateIsFinishedByPreNodeId(ctx, t.ProcInstID, nodeId, SystemPass, SystemPassComment)
	}

	// 会签节点 pass + task 数量相同才修改
	if passNum == taskNum {
		return e.engineSvc.UpdateIsFinishedByPreNodeId(ctx, t.ProcInstID, nodeId, SystemPass, SystemPassComment)
	}

	return nil
}

// EventSelectiveGatewaySplit 条件并行网关分叉处理
func (e *ProcessEvent) EventSelectiveGatewaySplit(instID int, node *model.Node, prevNode model.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// 1. 获取流程快照与定义
	inst, err := e.engineSvc.GetInstanceByID(ctx, instID)
	if err != nil {
		return err
	}
	processDefine, err := e.engineSvc.GetProcessDefineByVersion(ctx, inst.ProcID, inst.ProcVersion)
	if err != nil {
		return err
	}

	// 2. 检索 Selective 网关的直接下级（Condition 节点）
	nodeMap := slice.ToMap(processDefine.Nodes, func(n model.Node) string { return n.NodeID })
	conditionNodeIDs := node.GWConfig.InevitableNodes
	if len(conditionNodeIDs) == 0 {
		return nil
	}

	// 3. 逐个评估分支条件
	for _, condNodeID := range conditionNodeIDs {
		condNode, exists := nodeMap[condNodeID]
		if !exists || len(condNode.GWConfig.Conditions) == 0 {
			continue
		}

		// 只要分支内有一个 Expression 成立，该分支即为“激活”
		branchActive := false
		for _, cond := range condNode.GWConfig.Conditions {
			if passed, _ := e.evaluateExpression(instID, cond.Expression); passed {
				branchActive = true
				break
			}
		}

		// 4. 若全部分支条件均不满足，则静默跳过该分支的所有目标节点
		if !branchActive {
			for _, cond := range condNode.GWConfig.Conditions {
				e.skipBranch(ctx, instID, cond.NodeID, prevNode.NodeID)
			}
		}
	}

	return nil
}

// evaluateExpression 内部评估器：处理表达式中的变量注入、SQL 转义及 JSON 适配
func (e *ProcessEvent) evaluateExpression(instID int, expr string) (bool, error) {
	// 注入变量：匹配以 $ 开头的变量标识符
	reg := regexp.MustCompile(`[$]\w+`)
	variables := reg.FindAllString(expr, -1)
	if len(variables) > 0 {
		kv, err := engine.ResolveVariables(instID, variables)
		if err != nil {
			return false, err
		}
		for k, v := range kv {
			v = strings.Replace(v, "'", "\\'", -1) // 基础 SQL 防注入转义
			expr = strings.Replace(expr, k, fmt.Sprintf("'%s'", v), -1)
		}
	}

	// 特殊适配：JSON 数组的 IN 查询转换
	jsonArrayInPattern := regexp.MustCompile(`'(\[.*?])'\s+(?i)in\s+\(\s*'([^']+)'\s*\)`)
	if jsonArrayInPattern.MatchString(expr) {
		expr = jsonArrayInPattern.ReplaceAllString(expr, `JSON_CONTAINS('$1', '"$2"')`)
	}

	return engine.ExpressionEvaluator(expr)
}

// skipBranch 辅助方法：静默跳过节点并生成历史记录
func (e *ProcessEvent) skipBranch(ctx context.Context, instID int, nodeID, prevNodeID string) {
	e.logger.Info("条件不满足，自动跳过目标节点", elog.String("nodeID", nodeID), elog.Int("instID", instID))
	_ = e.engineSvc.CreateSkippedTask(ctx, instID, nodeID, prevNodeID, SystemSkippedComment, SystemSkipped)
}

func (e *ProcessEvent) getTargetNodeID(ctx context.Context, processInstId int, prevNodeID string, currentNode *model.Node) (string, error) {
	if currentNode.NodeName == SysProxyNodeName {
		return e.engineSvc.GetProxyPrevNodeID(ctx, processInstId, prevNodeID)
	}
	return prevNodeID, nil
}

// EventTaskParallelNodePass 用户任务并行处理事件
func (e *ProcessEvent) EventTaskParallelNodePass(taskID int, node *model.Node, prevNode model.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// 只要存在驳回，则强制同步驳回同一分支下的所有其他并行节点
	isReject, _ := e.engineSvc.IsReject(ctx, taskID)
	if isReject {
		taskInfo, _ := engine.GetTaskInfo(taskID)
		return e.engineSvc.UpdateIsFinishedByPreNodeId(ctx, taskInfo.ProcInstID, prevNode.NodeID, SystemReject, SystemRejectComment)
	}

	return nil
}

// EventConcurrentRejectCleanup 并行节点驳回清理事件
// 当并行分支中的某一个节点驳回时，自动清理（取消）同一分支下的其他兄弟任务
func (e *ProcessEvent) EventConcurrentRejectCleanup(taskID int, node *model.Node, prevNode model.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// 1. 获取任务详情，检查状态
	taskInfo, err := engine.GetTaskInfo(taskID)
	if err != nil {
		e.logger.Error("查询任务详情失败", elog.FieldErr(err))
		return err
	}

	// 只有驳回(Status=2)才触发清理
	if taskInfo.Status != 2 {
		return nil
	}

	e.logger.Info("并行节点驳回，触发兄弟节点清理",
		elog.Any("taskID", taskID),
		elog.Any("NodeName", node.NodeName),
		elog.Any("prevNodeID", prevNode.NodeID))

	// 2. 调用服务层清理逻辑
	// 使用 UpdateIsFinishedByPreNodeId 将同级任务置为 SystemReject (系统驳回/取消)
	// 注意：这里使用的是 prevNode.NodeID，即分支汇聚点（或分叉点）的ID
	return e.engineSvc.UpdateIsFinishedByPreNodeId(ctx, taskInfo.ProcInstID, prevNode.NodeID, SystemReject, SystemRejectComment)
}

// EventGatewayConditionReject 如果回退前是代理节点，那么需要修改为正确的节点ID
func (e *ProcessEvent) EventGatewayConditionReject(taskID int, node *model.Node, prevNode model.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// 1. 获取任务详情，检查状态
	taskInfo, err := engine.GetTaskInfo(taskID)
	if err != nil {
		e.logger.Error("查询任务详情失败", elog.FieldErr(err))
		return nil // 获取失败不阻断流程，只是Hack失败
	}

	// 只有驳回(Status=2)才触发穿透处理
	if taskInfo.Status != 2 {
		return nil
	}

	e.logger.Info("检测到网关后置节点驳回，尝试查找 proxy 节点",
		elog.Int("taskID", taskID),
		elog.String("node", node.NodeName))

	// 2. 查找 proxy 节点
	// NOTE: 通过流程实例ID查找 proxy 节点，获取其 prev_node_id 作为真正的回退目标
	proxyTask, err := e.engineSvc.GetProxyTaskByProcessInstId(ctx, taskInfo.ProcInstID)
	if err != nil {
		// 没有找到 proxy 节点，说明这个网关内没有 proxy，不需要处理
		e.logger.Info("未找到 proxy 节点，无需穿透",
			elog.Int("ProcInstID", taskInfo.ProcInstID))
		return nil
	}

	// 判定是 proxy 节点在处理逻辑
	if proxyTask.UserID != SysAutoUser {
		return nil
	}

	e.logger.Info("找到 proxy 节点，准备执行穿透",
		elog.String("ProxyNodeID", proxyTask.NodeID),
		elog.String("ProxyPrevNodeID", proxyTask.PrevNodeID))

	// 3. 篡改数据库：将当前任务的 prev_node_id 改为 proxy 的 prev_node_id
	// NOTE: proxy.prev_node_id 就是真正的回退目标节点（如：李四）
	targetNodeID := proxyTask.PrevNodeID

	e.logger.Info("执行来源篡改",
		elog.String("node", node.NodeName),
		elog.String("OriginalPrev", taskInfo.PrevNodeID),
		elog.String("NewPrev", targetNodeID))

	err = e.engineSvc.UpdateTaskPrevNodeID(ctx, taskID, targetNodeID)
	if err != nil {
		e.logger.Error("修改任务上一级节点失败", elog.FieldErr(err))
		return err
	}

	e.logger.Info("成功修改 prev_node_id，驳回将回退到正确节点")

	// 4. 删除 proxy 节点
	// NOTE: 驳回发生时，Proxy 节点已经完成了它的历史使命，需要删除，防止干扰后续流程
	// 这与 EventUserNodeRejectProxyCleanup 的目的类似
	err = e.engineSvc.DeleteProxyNodeByNodeId(ctx, taskInfo.ProcInstID, proxyTask.NodeID)
	if err != nil {
		e.logger.Error("删除 proxy 节点失败", elog.FieldErr(err))
		// 删除失败不阻断主流程，因为 prev_node_id 已经修改成功，流程回退路径已修正
	} else {
		e.logger.Info("成功清理 proxy 节点记录", elog.Int("ProcInstID", taskInfo.ProcInstID))
	}

	return nil
}

// EventUserNodeRejectProxyCleanup 用户节点驳回时清理 proxy 节点事件
// NOTE: 当检测到网关内存在 proxy 节点时，网关内的所有用户节点都应该注册此事件
// 作用：当用户节点被驳回时，自动将同级的 proxy 节点状态修改为驳回
func (e *ProcessEvent) EventUserNodeRejectProxyCleanup(taskID int, node *model.Node, prevNode model.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// 1. 获取任务详情，检查状态
	taskInfo, err := engine.GetTaskInfo(taskID)
	if err != nil {
		e.logger.Error("查询任务详情失败", elog.FieldErr(err))
		return err
	}

	// 只有驳回(Status=2)才触发 proxy 节点清理
	if taskInfo.Status != 2 {
		return nil
	}

	e.logger.Info("用户节点驳回，触发 proxy 节点清理",
		elog.Any("taskID", taskID),
		elog.Any("NodeName", node.NodeName),
		elog.Any("PrevNodeID", prevNode.NodeID))

	// 2. 查找同级的 proxy 节点
	// NOTE: 使用流程实例ID查找 proxy 节点，因为一个流程实例只可能有一个 proxy 节点
	proxyNodeID, err := e.engineSvc.GetProxyNodeByProcessInstId(ctx, taskInfo.ProcInstID)
	if err != nil {
		// NOTE: 如果找不到 proxy 节点，说明该网关内可能没有 proxy 节点，不影响主流程
		e.logger.Warn("未找到 proxy 节点",
			elog.Int("ProcInstID", taskInfo.ProcInstID),
			elog.FieldErr(err))
		return nil
	}

	e.logger.Info("检测到 proxy 节点，准备删除任务记录",
		elog.String("ProxyNodeID", proxyNodeID),
		elog.String("UserNodeID", node.NodeID),
		elog.Int("ProcInstID", taskInfo.ProcInstID))

	// 3. 删除 proxy 节点任务记录
	// NOTE: 修改状态无法阻止工作流引擎的判断，必须直接删除任务记录
	// 这里使用获取到的 proxyNodeID 进行精确删除
	err = e.engineSvc.DeleteProxyNodeByNodeId(ctx, taskInfo.ProcInstID, proxyNodeID)
	if err != nil {
		e.logger.Error("删除 proxy 节点任务记录失败", elog.FieldErr(err))
		return err
	}

	e.logger.Info("成功删除 proxy 节点任务记录",
		elog.String("ProxyNodeID", proxyNodeID),
		elog.Int("ProcInstID", taskInfo.ProcInstID))

	return nil
}

// EventRevoke 流程撤销
func (e *ProcessEvent) EventRevoke(instID int, RevokeUserID string) error {
	processName, err := engine.GetProcessNameByInstanceID(instID)
	if err != nil {
		return err
	}

	log.Printf("--------流程[%s],由[%s]发起撤销--------", processName, RevokeUserID)

	return nil
}

func (e *ProcessEvent) autoPassNode(instID int, nodeID string, comment string) {
	// 创建 10 秒超时上下文，用于重试等待任务生成
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			e.logger.Error("节点自动流转失败：等待任务创建超时",
				elog.Int("InstanceID", instID),
				elog.String("NodeID", nodeID))
			return
		case <-ticker.C:
			tasks, err := e.engineSvc.GetTasksByCurrentNodeId(ctx, instID, nodeID)
			if err == nil && len(tasks) > 0 {
				// 找到任务，执行通过
				if passErr := e.engineSvc.Pass(ctx, tasks[0].TaskID, comment); passErr != nil {
					e.logger.Error("节点自动流转通过失败",
						elog.Int("taskID", tasks[0].TaskID),
						elog.FieldErr(passErr))
				}
				return
			}
		}
	}
}

// EventInclusionPassCleanup 包容网关通过事件
// 当包容网关的某一个分支通过时，将其他并行分支置为系统自动通过（SystemPass），实现“一票通过”或“竞争”模式
func (e *ProcessEvent) EventInclusionPassCleanup(taskID int, node *model.Node, prevNode model.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// 1. 获取任务详情，检查状态
	taskInfo, err := engine.GetTaskInfo(taskID)
	if err != nil {
		e.logger.Error("查询任务详情失败", elog.FieldErr(err))
		return err
	}

	// 只有通过(Status=1)才触发清理
	if taskInfo.Status != 1 {
		return nil
	}

	// 检查会签状态
	taskNum, passNum, _, err := engine.TaskNodeStatus(taskID)
	if err != nil {
		return err
	}

	// 如果是会签节点，且未全部通过，则不触发清理
	if taskInfo.IsCosigned == 1 && passNum < taskNum {
		e.logger.Info("包含节点会签未完成，暂不清理兄弟节点",
			elog.Any("taskID", taskID),
			elog.Int("PassNum", passNum),
			elog.Int("TotalNum", taskNum))
		return nil
	}

	e.logger.Info("包含节点通过，触发兄弟节点清理",
		elog.Any("taskID", taskID),
		elog.Any("NodeName", node.NodeName),
		elog.Any("prevNodeID", prevNode.NodeID))

	// 2. 调用服务层清理逻辑
	// 使用 UpdateIsFinishedByPreNodeId 将同级任务置为 SystemPass (系统通过)
	// 注意：这里使用的是 prevNode.NodeID，即分支汇聚点（或分叉点）的ID
	return e.engineSvc.UpdateIsFinishedByPreNodeId(ctx, taskInfo.ProcInstID, prevNode.NodeID, SystemPass, SystemPassComment)
}
