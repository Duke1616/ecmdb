package service

import (
	"context"
	"fmt"

	"github.com/Bunny3th/easy-workflow/workflow/engine"
	engineSvc "github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/order/internal/errs"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
)

type ProcessEngine interface {
	// Pass 通过
	Pass(ctx context.Context, taskId int, comment string, extraData map[string]interface{}) error
	// Reject 驳回
	Reject(ctx context.Context, taskId int, comment string) error
	// Revoke 撤销工单
	Revoke(ctx context.Context, instanceId int, userId string, force bool) error
}

type processEngine struct {
	svc         Service
	logger      *elog.Component
	engineSvc   engineSvc.Service
	workflowSvc workflow.Service
}

func NewProcessEngine(svc Service, engineSvc engineSvc.Service, workflowSvc workflow.Service) ProcessEngine {
	return &processEngine{
		svc:         svc,
		logger:      elog.DefaultLogger.With(elog.FieldComponentName("processEngine")),
		engineSvc:   engineSvc,
		workflowSvc: workflowSvc,
	}
}

func (e *processEngine) Pass(ctx context.Context, taskId int, comment string, extraData map[string]interface{}) error {
	// 1. 获取任务详情，拿到流程实例ID
	taskInfo, err := e.engineSvc.TaskInfo(ctx, taskId)
	if err != nil {
		return fmt.Errorf("获取任务详情失败: %w", err)
	}

	return engine.TaskPass(taskId, comment, "", false)

	// 2. 获取流程版本
	instance, err := e.engineSvc.GetInstanceByID(ctx, taskInfo.ProcInstID)
	if err != nil {
		return err
	}

	// 3. 获取版本流程图定义
	snapshot, err := e.workflowSvc.GetWorkflowSnapshot(ctx, taskInfo.ProcID, instance.ProcVersion)
	if err != nil {
		e.logger.Error("获取版本流程失败", elog.FieldErr(err),
			elog.Int("流程ID", taskInfo.ProcID),
			elog.Int("流程版本", instance.ProcVersion))
		return err
	}

	// 4. 解析 node 节点数据
	nodes, _ := easyflow.ParseNodes(snapshot.FlowData.Nodes)
	node, ok := slice.Find(nodes, func(node easyflow.Node) bool {
		return node.ID == taskInfo.NodeID
	})
	if !ok {
		return fmt.Errorf("node 节点不存在, %s", taskInfo.NodeID)
	}
	property, err := easyflow.ToNodeProperty[easyflow.UserProperty](node)
	if err != nil {
		return err
	}

	// 5. 如果没有定义字段，直接 Pass
	if len(property.Fields) == 0 {
		return engine.TaskPass(taskId, comment, "", false)
	}

	// 6. 校验必填数据
	if field, okk := slice.Find(property.Fields, func(field easyflow.Field) bool {
		// 如果不是必填项，直接跳过
		if !field.Required {
			return false
		}

		// 如果是必填项，检查是否有值
		val, exists := extraData[field.Key]
		return !exists || val == nil || val == ""
	}); okk {
		return fmt.Errorf("%w: 字段 [%s] 为必填项，请填写", errs.ValidationError, field.Name)
	}

	// 7. 根据流程实例ID查找工单
	order, err := e.svc.DetailByProcessInstId(ctx, taskInfo.ProcInstID)
	if err != nil {
		return fmt.Errorf("查询关联工单失败: %w", err)
	}

	// 8. 更新工单数据
	if err = e.svc.MergeOrderData(ctx, order.Id, extraData); err != nil {
		return fmt.Errorf("更新工单数据失败: %w", err)
	}

	// 9. 记录任务数据快照
	if err = e.svc.RecordSnapshotsData(ctx, taskId, extraData); err != nil {
		return fmt.Errorf("记录任务快照失败: %w", err)
	}

	return engine.TaskPass(taskId, comment, "", false)
}

func (e *processEngine) Reject(ctx context.Context, taskId int, comment string) error {
	return engine.TaskReject(taskId, comment, "")
}

func (e *processEngine) Revoke(ctx context.Context, instanceId int, userId string, force bool) error {
	return engine.InstanceRevoke(instanceId, force, userId)
}
