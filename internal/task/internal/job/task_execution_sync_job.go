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

		var syncTasks []domain.Task
		var taskIds []int64

		for _, task := range tasks {
			// 已经生成 external_id 的
			if task.ExternalId == "" {
				continue
			}

			taskId, err1 := strconv.ParseInt(task.ExternalId, 10, 64)
			if err1 != nil {
				c.logger.Error("sync: 解析 external_id 失败", elog.FieldErr(err), elog.String("external_id", task.ExternalId))
				continue
			}

			syncTasks = append(syncTasks, task)
			taskIds = append(taskIds, taskId)
		}

		if len(taskIds) > 0 {
			c.batchSyncTaskExecutions(ctx, taskIds, syncTasks)
		}

		if int64(len(tasks)) < c.limit {
			break
		}
		offset += c.limit
		if offset >= total {
			break
		}
	}

	return nil
}

func (c *TaskExecutionSyncJob) batchSyncTaskExecutions(ctx context.Context, taskIds []int64, syncTasks []domain.Task) {
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := c.executorSvc.BatchListTaskExecutions(reqCtx, &executorv1.BatchListTaskExecutionsRequest{
		TaskIds: taskIds,
	})
	if err != nil {
		c.logger.Error("sync: 批量查询远程任务执行记录失败", elog.FieldErr(err))
		return
	}

	var wg sync.WaitGroup
	for _, task := range syncTasks {
		taskId, _ := strconv.ParseInt(task.ExternalId, 10, 64)
		execList, ok := resp.Results[taskId]

		if !ok || len(execList.Executions) == 0 {
			continue
		}

		// 取最新执行记录
		var latest *executorv1.TaskExecution
		for _, exec := range execList.Executions {
			if latest == nil || exec.Id > latest.Id {
				latest = exec
			}
		}

		if latest == nil {
			continue
		}

		wg.Add(1)
		go func(t domain.Task, latestStatus executorv1.ExecutionStatus) {
			defer wg.Done()
			var updateErr error
			switch latestStatus {
			case executorv1.ExecutionStatus_SUCCESS:
				_, updateErr = c.svc.UpdateTaskStatus(ctx, domain.TaskResult{
					Id:              t.Id,
					Status:          domain.SUCCESS,
					Result:          "任务执行成功",
					TriggerPosition: "远程调度节点返回",
				})
			case executorv1.ExecutionStatus_FAILED, executorv1.ExecutionStatus_FAILED_RETRYABLE, executorv1.ExecutionStatus_FAILED_RESCHEDULABLE:
				_, updateErr = c.svc.UpdateTaskStatus(ctx, domain.TaskResult{
					Id:              t.Id,
					Status:          domain.FAILED,
					Result:          "任务执行失败",
					TriggerPosition: "远程调度节点返回",
				})
			default:
				// 如 RUNNING, UNKNOWN 暂不更新本地状态
				return
			}

			if updateErr != nil {
				c.logger.Error("sync: 更新本地任务状态失败", elog.FieldErr(updateErr), elog.Int64("task_id", t.Id))
			} else {
				c.logger.Info("sync: 同步任务状态成功", elog.Int64("task_id", t.Id), elog.Any("status", latestStatus))
			}
		}(task, latest.Status)
	}
	wg.Wait()
}
