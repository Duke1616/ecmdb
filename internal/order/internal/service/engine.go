package service

import (
	"context"
	"fmt"

	"github.com/Bunny3th/easy-workflow/workflow/engine"
	engineSvc "github.com/Duke1616/ecmdb/internal/engine"
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
	svc       Service
	engineSvc engineSvc.Service
}

func NewProcessEngine(svc Service) ProcessEngine {
	return &processEngine{svc: svc}
}

func (e *processEngine) Pass(ctx context.Context, taskId int, comment string, extraData map[string]interface{}) error {
	// 如果携带了表单数据，先更新工单数据
	if len(extraData) > 0 {
		// 1. 获取任务详情，拿到流程实例ID
		taskInfo, err := e.engineSvc.TaskInfo(ctx, taskId)
		if err != nil {
			return fmt.Errorf("获取任务详情失败: %w", err)
		}

		// 2. 根据流程实例ID查找工单
		order, err := e.svc.DetailByProcessInstId(ctx, taskInfo.ProcInstID)
		if err != nil {
			return fmt.Errorf("查询关联工单失败: %w", err)
		}

		// 3. 更新工单数据
		if err = e.svc.MergeOrderData(ctx, order.Id, extraData); err != nil {
			return fmt.Errorf("更新工单数据失败: %w", err)
		}
	}

	return engine.TaskPass(taskId, comment, "", false)
}

func (e *processEngine) Reject(ctx context.Context, taskId int, comment string) error {
	return engine.TaskReject(taskId, comment, "")
}

func (e *processEngine) Revoke(ctx context.Context, instanceId int, userId string, force bool) error {
	return engine.InstanceRevoke(instanceId, force, userId)
}
