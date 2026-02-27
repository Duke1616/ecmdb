package job

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	executorv1 "github.com/Duke1616/ecmdb/api/proto/gen/etask/executor/v1"
	"github.com/Duke1616/ecmdb/internal/task/internal/domain"
	"github.com/Duke1616/ecmdb/internal/task/internal/service"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/task/ecron"
)

var _ ecron.NamedJob = (*TaskExecutionSyncJob)(nil)

type TaskExecutionSyncJob struct {
	svc         service.Service
	executorSvc executorv1.TaskExecutionServiceClient
	logger      *elog.Component
	limit       int64
}

func NewTaskExecutionSyncJob(svc service.Service, executorSvc executorv1.TaskExecutionServiceClient, limit int64) *TaskExecutionSyncJob {
	return &TaskExecutionSyncJob{
		svc:         svc,
		executorSvc: executorSvc,
		logger:      elog.DefaultLogger,
		limit:       limit,
	}
}

func (c *TaskExecutionSyncJob) Name() string {
	return "TaskExecutionSync"
}

func (c *TaskExecutionSyncJob) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	offset := int64(0)
	for {
		// 分页只拉取 RUNNING 状态且属于 Execute 分布式模式的任务
		tasks, total, err := c.svc.ListTaskByStatusAndMode(ctx, offset, c.limit,
			domain.RUNNING.ToUint8(),
			domain.RunModeExecute.ToString())
		if err != nil {
			return fmt.Errorf("sync 获取运行中任务列表失败: %w", err)
		}

		if offset == 0 && len(tasks) > 0 {
			c.logger.Info("sync task execution job start", elog.Int64("total", total))
		}

		for _, task := range tasks {
			// 只同步分布式平台执行的任务，且已经生成 external_id 的
			if task.RunMode != domain.RunModeExecute || task.ExternalId == "" {
				continue
			}

			wg.Add(1)
			go func(t domain.Task) {
				defer wg.Done()
				c.syncTaskExecution(context.Background(), t)
			}(task)
		}

		if int64(len(tasks)) < c.limit {
			break
		}
		offset += c.limit
		if offset >= total {
			break
		}
	}

	wg.Wait()
	return nil
}

func (c *TaskExecutionSyncJob) syncTaskExecution(ctx context.Context, task domain.Task) {
	taskId, err := strconv.ParseInt(task.ExternalId, 10, 64)
	if err != nil {
		c.logger.Error("sync: 解析 external_id 失败", elog.FieldErr(err), elog.String("external_id", task.ExternalId))
		return
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.executorSvc.ListTaskExecutions(reqCtx, &executorv1.ListTaskExecutionsRequest{
		TaskId: taskId,
	})
	if err != nil {
		c.logger.Error("sync: 查询远程任务执行记录失败", elog.FieldErr(err), elog.Int64("task_id", taskId))
		return
	}

	if len(resp.Executions) == 0 {
		return
	}

	// 取最新执行记录（通常可以依据 ID 或 StartTime 决定）
	var latest *executorv1.TaskExecution
	for _, exec := range resp.Executions {
		if latest == nil || exec.Id > latest.Id {
			latest = exec
		}
	}

	if latest == nil {
		return
	}

	// 根据执行节点状态来更新本地的 task
	switch latest.Status {
	case executorv1.ExecutionStatus_SUCCESS:
		_, err = c.svc.UpdateTaskStatus(ctx, domain.TaskResult{
			Id:              task.Id,
			Status:          domain.SUCCESS,
			Result:          "任务执行成功",
			TriggerPosition: "远程调度节点返回",
		})
	case executorv1.ExecutionStatus_FAILED, executorv1.ExecutionStatus_FAILED_RETRYABLE, executorv1.ExecutionStatus_FAILED_RESCHEDULABLE:
		_, err = c.svc.UpdateTaskStatus(ctx, domain.TaskResult{
			Id:              task.Id,
			Status:          domain.FAILED,
			Result:          "任务执行失败",
			TriggerPosition: "远程调度节点返回",
		})
	default:
		// 如 RUNNING, UNKNOWN 暂不更新本地状态
		return
	}

	if err != nil {
		c.logger.Error("sync: 更新本地任务状态失败", elog.FieldErr(err), elog.Int64("task_id", task.Id))
	} else {
		c.logger.Info("sync: 同步任务状态成功", elog.Int64("task_id", task.Id), elog.Any("status", latest.Status))
	}
}
