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
	if len(CurrentNode.UserIDs) > 0 && CurrentNode.UserIDs[0] == "sys_auto" {
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

	taskNum, passNum, rejectNum, err := engine.TaskNodeStatus(TaskID)
	e.logger.Info("包含网关-处理节点状态系统自动变更", elog.Any("任务ID", TaskID),
		elog.Any("Node节点", PrevNode.NodeID),
		elog.Any("任务数量", taskNum),
		elog.Any("通过数量", passNum),
		elog.Any("驳回数量", rejectNum))

	if err != nil {
		return err
	}

	// 但凡是有驳回，一率进行处理
	if rejectNum > 0 {
		return e.engineSvc.UpdateIsFinishedByPreNodeId(ctx, PrevNode.NodeID, SystemReject, SystemRejectComment)
	}

	// 查看任务详情信息
	t, err := engine.GetTaskInfo(TaskID)
	if err != nil {
		return err
	}

	// 如果不是会签节点，直接修改所有
	if t.IsCosigned != 1 {
		return e.engineSvc.UpdateIsFinishedByPreNodeId(ctx, PrevNode.NodeID, SystemPass, SystemPassComment)
	}

	// 会签节点 pass + task 数量相同才修改
	if passNum == taskNum {
		return e.engineSvc.UpdateIsFinishedByPreNodeId(ctx, PrevNode.NodeID, SystemPass, SystemPassComment)
	}

	return nil
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
