package easyflow

import (
	"context"
	"fmt"
	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	engineSvc "github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/event/producer"
	"github.com/Duke1616/ecmdb/internal/task"
	"github.com/gotomicro/ego/core/elog"
	"time"

	"log"
)

// RoleUser 这里创建了一个角色-用户的人员库，用来模拟数据库中存储的角色-用户对应关系
var RoleUser = make(map[string][]string)

func init() {
	//初始化人事数据
	RoleUser["主管"] = []string{"张经理"}
	RoleUser["人事经理"] = []string{"人事老刘"}
	RoleUser["老板"] = []string{"李老板", "老板娘"}
	RoleUser["副总"] = []string{"赵总", "钱总", "孙总"}
}

type ProcessEvent struct {
	producer  producer.OrderStatusModifyEventProducer
	taskSvc   task.Service
	engineSvc engineSvc.Service
	logger    *elog.Component
}

func NewProcessEvent(producer producer.OrderStatusModifyEventProducer, engineSvc engineSvc.Service,
	taskSvc task.Service) *ProcessEvent {
	return &ProcessEvent{
		logger:    elog.DefaultLogger,
		engineSvc: engineSvc,
		taskSvc:   taskSvc,
		producer:  producer,
	}
}

// EventStart 节点结束事件
func (e *ProcessEvent) EventStart(ProcessInstanceID int, CurrentNode *model.Node, PrevNode model.Node) error {
	//可以做一些处理，比如通知流程开始人，节点到了哪个步骤
	processName, err := engine.GetProcessNameByInstanceID(ProcessInstanceID)
	if err != nil {
		return err
	}
	log.Printf("--------流程[%s]节点[%s]结束-------", processName, CurrentNode.NodeName)
	return nil
}

// EventAutomation 自动化任务处理（创建任务）
func (e *ProcessEvent) EventAutomation(ProcessInstanceID int, CurrentNode *model.Node, PrevNode model.Node) error {
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	//defer cancel()

	err := e.taskSvc.CreateTask(context.Background(), ProcessInstanceID, CurrentNode.NodeID)
	fmt.Println(ProcessInstanceID, CurrentNode.NodeID)
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

// EventNotify 通知
func (e *ProcessEvent) EventNotify(ProcessInstanceID int, CurrentNode *model.Node, PrevNode model.Node) error {
	processName, err := engine.GetProcessNameByInstanceID(ProcessInstanceID)
	if err != nil {
		return err
	}
	log.Printf("--------流程[%s]节点[%s]，通知节点中对应人员--------", processName, CurrentNode.NodeName)
	if CurrentNode.NodeType == model.EndNode {
		log.Printf("============================== 流程[%s]结束 ==============================", processName)
		variables, err := engine.ResolveVariables(ProcessInstanceID, []string{"$starter"})
		if err != nil {
			return err
		}
		log.Printf("通知流程创建人%s,流程[%s]已完成", variables["$starter"], processName)

	} else {
		for _, user := range CurrentNode.UserIDs {
			log.Printf("通知用户[%s],抓紧去处理", user)
		}
	}
	return nil
}

// EventTaskInclusionNodePass 用户任务并行包容处理事件
// 当处于并行 或 包容网关的时候，其中一个节点驳回，其余并行节点并不会修改状态
func (e *ProcessEvent) EventTaskInclusionNodePass(TaskID int, CurrentNode *model.Node, PrevNode model.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	taskNum, passNum, rejectNum, err := engine.TaskNodeStatus(TaskID)
	e.logger.Info("处理节点状态系统自动变更", elog.Any("任务ID", TaskID),
		elog.Any("Node节点", PrevNode.NodeID),
		elog.Any("任务数量", taskNum),
		elog.Any("通过数量", passNum),
		elog.Any("驳回数量", rejectNum))

	if err != nil {
		return err
	}

	if rejectNum > 0 {
		return e.engineSvc.UpdateIsFinishedByPreNodeId(ctx, PrevNode.NodeID)
	}

	task, err := engine.GetTaskInfo(TaskID)
	if err != nil {
		return err
	}

	// 如果不是会签节点，直接修改所有
	if task.IsCosigned != 1 {
		return e.engineSvc.UpdateIsFinishedByPreNodeId(ctx, PrevNode.NodeID)
	}

	// 会签节点 pass + task 数量相同才修改
	if passNum == taskNum {
		return e.engineSvc.UpdateIsFinishedByPreNodeId(ctx, PrevNode.NodeID)
	}

	return nil
}

// EventTaskParallelNodePass 用户任务并行处理事件
// 当处于并行 或 包容网关的时候，其中一个节点驳回，其余并行节点并不会修改状态
func (e *ProcessEvent) EventTaskParallelNodePass(TaskID int, CurrentNode *model.Node, PrevNode model.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
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
		return e.engineSvc.UpdateIsFinishedByPreNodeId(ctx, PrevNode.NodeID)
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
