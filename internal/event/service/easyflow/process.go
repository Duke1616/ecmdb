package easyflow

import (
	"context"
	"strconv"
	"time"

	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	engineSvc "github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/event/domain"
	"github.com/Duke1616/ecmdb/internal/event/producer"
	"github.com/Duke1616/ecmdb/internal/event/service/strategy"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/task"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/gotomicro/ego/core/elog"

	"log"
)

const (
	SystemPass          = 3
	SystemReject        = 4
	UserRevoke          = 5
	SystemPassComment   = "其余节点审批通过，系统判定无法继续审批"
	SystemRejectComment = "其余节点进行驳回，系统判定无法继续审批"
	SysAutoUser         = "sys_auto"
	SysProxyNodeName    = "系统代理流转"
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

// EventStart 节点结束事件
func (e *ProcessEvent) EventStart(ProcessInstanceID int, CurrentNode *model.Node, PrevNode model.Node) error {
	//可以做一些处理，比如通知流程开始人，节点到了哪个步骤
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// 查看工单关联
	orderInfo, wfInfo, err := e.fetchOrderAndWorkflow(ctx, ProcessInstanceID)

	var ok bool
	ok, err = e.strategy.Send(ctx, domain.StrategyInfo{
		NodeName:    domain.Start,
		OrderInfo:   orderInfo,
		WfInfo:      wfInfo,
		InstanceId:  ProcessInstanceID,
		CurrentNode: CurrentNode,
	})

	if err != nil || !ok {
		e.logger.Error("【EventStart】 消息发送失败：", elog.FieldErr(err), elog.Any("流程ID", ProcessInstanceID))
	}

	// 这个必须成功，不然会导致后续任务无法进行
	return e.orderSvc.RegisterProcessInstanceId(ctx, orderInfo.Id, ProcessInstanceID)
}

// EventAutomation 自动化任务处理（创建任务）
func (e *ProcessEvent) EventAutomation(ProcessInstanceID int, CurrentNode *model.Node, PrevNode model.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// 使用goroutine执行任务创建，并等待其完成
	var err error
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, err = e.taskSvc.CreateTask(ctx, ProcessInstanceID, CurrentNode.NodeID)
		if err != nil {
			e.logger.Error("创建自动化任务失败",
				elog.Any("流程ID", ProcessInstanceID),
				elog.Any("节点ID", CurrentNode.NodeID),
				elog.Any("错误信息", err),
			)
		}

	}()

	// 等待goroutine完成或超时
	select {
	case <-done:
		// goroutine正常完成
		if err != nil {
			return err
		}
	case <-ctx.Done():
		// 超时或取消
		e.logger.Error("创建自动化任务超时或被取消")
		return ctx.Err()
	}

	return err
}

// EventEnd 节点结束事件
func (e *ProcessEvent) EventEnd(ProcessInstanceID int, CurrentNode *model.Node, PrevNode model.Node) error {
	processName, err := engine.GetProcessNameByInstanceID(ProcessInstanceID)
	if err != nil {
		return err
	}

	e.logger.Info("节点结束了", elog.Any("processName", processName))
	log.Printf("--------流程[%s]节点[%s]结束-------", processName, CurrentNode.NodeName)
	return nil
}

// EventClose 流程结束，修改 Order 状态为已完成
// Deprecated 废弃 不再通过 Kafka 修改状态，使用 EventNotify 直接调用接口进行修改
func (e *ProcessEvent) EventClose(ProcessInstanceID int, CurrentNode *model.Node, PrevNode model.Node) error {
	evt := producer.OrderStatusModifyEvent{
		ProcessInstanceId: ProcessInstanceID,
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

// EventNotify 通知 中间有 Error 通过日志记录，保证不影响主体程序运行
func (e *ProcessEvent) EventNotify(ProcessInstanceID int, CurrentNode *model.Node, PrevNode model.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	// 如果是结束节点，暂时不做任何处理
	if CurrentNode.NodeType == model.EndNode {
		// 关闭工单
		err := e.orderSvc.UpdateStatusByInstanceId(ctx, ProcessInstanceID, order.EndProcess.ToUint8())
		if err != nil {
			e.logger.Error("EventNotify 关闭工单失败：",
				elog.FieldErr(err),
				elog.Int("流程ID", ProcessInstanceID))
		}
	}

	// 判断是否为系统自动节点
	if len(CurrentNode.UserIDs) > 0 && CurrentNode.UserIDs[0] == SysAutoUser {
		go e.autoPassProxyNode(ProcessInstanceID, CurrentNode.NodeID)
		return nil
	}

	orderInfo, wfInfo, err := e.fetchOrderAndWorkflow(ctx, ProcessInstanceID)

	// 判断消息的来源，处理不同的消息通知模式
	nodeMethod := domain.User
	if len(CurrentNode.UserIDs) == 1 && CurrentNode.UserIDs[0] == "automation" {
		nodeMethod = domain.Automation
	}

	var ok bool
	ok, err = e.strategy.Send(ctx, domain.StrategyInfo{
		NodeName:    nodeMethod,
		OrderInfo:   orderInfo,
		WfInfo:      wfInfo,
		InstanceId:  ProcessInstanceID,
		CurrentNode: CurrentNode,
	})

	if err != nil || !ok {
		e.logger.Error("【EventNotify】 消息发送失败：", elog.FieldErr(err), elog.Any("流程ID", ProcessInstanceID))
	}

	return nil
}

// EventTaskInclusionNodePass 用户任务并行包容处理事件
// 当处于并行 或 包容网关的时候，其中一个节点驳回，其余并行节点并不会修改状态
func (e *ProcessEvent) EventTaskInclusionNodePass(TaskID int, CurrentNode *model.Node, PrevNode model.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	e.logger.Info("包含网关-获取当前节点", elog.Any("Node名称", CurrentNode.NodeName),
		elog.Any("Node节点", CurrentNode.NodeID))

	taskNum, passNum, rejectNum, err := engine.TaskNodeStatus(TaskID)
	e.logger.Info("包含网关-处理节点状态系统自动变更", elog.Any("任务ID", TaskID),
		elog.Any("Node名称", PrevNode.NodeName),
		elog.Any("Node节点", PrevNode.NodeID),
		elog.Any("任务数量", taskNum),
		elog.Any("通过数量", passNum),
		elog.Any("驳回数量", rejectNum))

	if err != nil {
		return err
	}

	// 如果是代理节点，需要查询代理节点的上级
	nodeId, err := e.getTargetNodeID(ctx, PrevNode.NodeID, CurrentNode)
	if err != nil {
		return err
	}

	e.logger.Info("包含网关-触发处理", elog.String("nodeId", nodeId))

	// 但凡是有驳回，一率进行处理
	if rejectNum > 0 {
		return e.engineSvc.UpdateIsFinishedByPreNodeId(ctx, nodeId, SystemReject, SystemRejectComment)
	}

	// 查看任务详情信息
	t, err := engine.GetTaskInfo(TaskID)
	if err != nil {
		return err
	}

	// 如果不是会签节点，直接修改所有
	if t.IsCosigned != 1 {
		return e.engineSvc.UpdateIsFinishedByPreNodeId(ctx, nodeId, SystemPass, SystemPassComment)
	}

	// 会签节点 pass + task 数量相同才修改
	if passNum == taskNum {
		return e.engineSvc.UpdateIsFinishedByPreNodeId(ctx, nodeId, SystemPass, SystemPassComment)
	}

	return nil
}

func (e *ProcessEvent) getTargetNodeID(ctx context.Context, prevNodeID string, currentNode *model.Node) (string, error) {
	if currentNode.NodeName == SysProxyNodeName {
		return e.engineSvc.GetProxyPrevNodeID(ctx, prevNodeID)
	}
	return prevNodeID, nil
}

// EventTaskParallelNodePass 用户任务并行处理事件
// 当处于并行 或 包容网关的时候，其中一个节点驳回，其余并行节点并不会修改状态
func (e *ProcessEvent) EventTaskParallelNodePass(TaskID int, CurrentNode *model.Node, PrevNode model.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// 查看错误数量
	IsReject, err := e.engineSvc.IsReject(ctx, TaskID)
	if err != nil {
		return err
	}

	e.logger.Info("处理节点状态系统自动变更", elog.Any("任务ID", TaskID),
		elog.Any("Node节点", PrevNode.NodeID),
		elog.Any("是否驳回", IsReject))

	if IsReject {
		return e.engineSvc.UpdateIsFinishedByPreNodeId(ctx, PrevNode.NodeID, SystemReject, SystemRejectComment)
	}

	return nil
}

// EventConcurrentRejectCleanup 并行节点驳回清理事件
// 当并行分支中的某一个节点驳回时，自动清理（取消）同一分支下的其他兄弟任务
func (e *ProcessEvent) EventConcurrentRejectCleanup(TaskID int, CurrentNode *model.Node, PrevNode model.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// 1. 获取任务详情，检查状态
	taskInfo, err := engine.GetTaskInfo(TaskID)
	if err != nil {
		e.logger.Error("查询任务详情失败", elog.FieldErr(err))
		return err
	}

	// 只有驳回(Status=2)才触发清理
	if taskInfo.Status != 2 {
		return nil
	}

	e.logger.Info("并行节点驳回，触发兄弟节点清理",
		elog.Any("TaskID", TaskID),
		elog.Any("NodeName", CurrentNode.NodeName),
		elog.Any("PrevNodeID", PrevNode.NodeID))

	// 2. 调用服务层清理逻辑
	// 使用 UpdateIsFinishedByPreNodeId 将同级任务置为 SystemReject (系统驳回/取消)
	// 注意：这里使用的是 PrevNode.NodeID，即分支汇聚点（或分叉点）的ID
	return e.engineSvc.UpdateIsFinishedByPreNodeId(ctx, PrevNode.NodeID, SystemReject, SystemRejectComment)
}

// EventGatewayConditionReject 如果回退前是代理节点，那么需要修改为正确的节点ID
func (e *ProcessEvent) EventGatewayConditionReject(TaskID int, CurrentNode *model.Node, PrevNode model.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// 1. 获取任务详情，检查状态
	taskInfo, err := engine.GetTaskInfo(TaskID)
	if err != nil {
		e.logger.Error("查询任务详情失败", elog.FieldErr(err))
		return nil // 获取失败不阻断流程，只是Hack失败
	}

	// 只有驳回(Status=2)才触发穿透处理
	if taskInfo.Status != 2 {
		return nil
	}

	e.logger.Info("检测到网关后置节点驳回，尝试查找 proxy 节点",
		elog.Int("TaskID", TaskID),
		elog.String("CurrentNode", CurrentNode.NodeName))

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
		elog.String("CurrentNode", CurrentNode.NodeName),
		elog.String("OriginalPrev", taskInfo.PrevNodeID),
		elog.String("NewPrev", targetNodeID))

	err = e.engineSvc.UpdateTaskPrevNodeID(ctx, TaskID, targetNodeID)
	if err != nil {
		e.logger.Error("修改任务上一级节点失败", elog.FieldErr(err))
		return err
	}

	e.logger.Info("成功修改 prev_node_id，驳回将回退到正确节点")

	// 4. 删除 proxy 节点
	// NOTE: 驳回发生时，Proxy 节点已经完成了它的历史使命，需要删除，防止干扰后续流程
	// 这与 EventUserNodeRejectProxyCleanup 的目的类似
	err = e.engineSvc.DeleteProxyNode(ctx, taskInfo.ProcInstID)
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
func (e *ProcessEvent) EventUserNodeRejectProxyCleanup(TaskID int, CurrentNode *model.Node, PrevNode model.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// 1. 获取任务详情，检查状态
	taskInfo, err := engine.GetTaskInfo(TaskID)
	if err != nil {
		e.logger.Error("查询任务详情失败", elog.FieldErr(err))
		return err
	}

	// 只有驳回(Status=2)才触发 proxy 节点清理
	if taskInfo.Status != 2 {
		return nil
	}

	e.logger.Info("用户节点驳回，触发 proxy 节点清理",
		elog.Any("TaskID", TaskID),
		elog.Any("NodeName", CurrentNode.NodeName),
		elog.Any("PrevNodeID", PrevNode.NodeID))

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
		elog.String("UserNodeID", CurrentNode.NodeID),
		elog.Int("ProcInstID", taskInfo.ProcInstID))

	// 3. 删除 proxy 节点任务记录
	// NOTE: 修改状态无法阻止工作流引擎的判断，必须直接删除任务记录
	err = e.engineSvc.DeleteProxyNode(ctx, taskInfo.ProcInstID)
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
func (e *ProcessEvent) EventRevoke(ProcessInstanceID int, RevokeUserID string) error {
	processName, err := engine.GetProcessNameByInstanceID(ProcessInstanceID)
	if err != nil {
		return err
	}

	log.Printf("--------流程[%s],由[%s]发起撤销--------", processName, RevokeUserID)

	return nil
}

func (e *ProcessEvent) fetchOrderAndWorkflow(ctx context.Context, processInstanceID int) (
	order.Order, workflow.Workflow, error) {
	// 获取流程变量中记录的工单ID
	orderId, err := e.engineSvc.GetOrderIdByVariable(ctx, processInstanceID)
	if err != nil {
		return order.Order{}, workflow.Workflow{}, err
	}

	// 转换为 int64
	id, err := strconv.ParseInt(orderId, 10, 64)
	if err != nil {
		return order.Order{}, workflow.Workflow{}, err
	}

	// 获取工单详情
	nOrder, err := e.orderSvc.Detail(ctx, id)
	if err != nil {
		e.logger.Error("查询工单详情错误",
			elog.FieldErr(err),
			elog.Any("instId", processInstanceID),
		)
		return order.Order{}, workflow.Workflow{}, err
	}

	// 获取流程信息
	wf, err := e.workflowSvc.Find(ctx, nOrder.WorkflowId)
	if err != nil {
		e.logger.Error("查询流程信息错误",
			elog.FieldErr(err),
			elog.Any("instId", processInstanceID),
		)
		return order.Order{}, workflow.Workflow{}, err
	}

	return nOrder, wf, nil
}

func (e *ProcessEvent) autoPassProxyNode(instanceID int, nodeID string) {
	// 创建 10 秒超时上下文，用于重试等待任务生成
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			e.logger.Error("User代理节点自动流转失败：等待任务创建超时",
				elog.Any("InstanceID", instanceID),
				elog.Any("NodeID", nodeID))
			return
		case <-ticker.C:
			tasks, err := e.engineSvc.GetTasksByCurrentNodeId(ctx, instanceID, nodeID)
			if err == nil && len(tasks) > 0 {
				// 找到任务，执行通过
				// 注意：这里使用的是 TaskID 字段，修正了之前的 ID 报错问题
				if passErr := e.engineSvc.Pass(ctx, tasks[0].TaskID, "Sys Auto Pass"); passErr != nil {
					e.logger.Error("User代理节点自动流转通过失败",
						elog.Any("TaskID", tasks[0].TaskID),
						elog.FieldErr(passErr))
				}
				return
			}
		}
	}
}
