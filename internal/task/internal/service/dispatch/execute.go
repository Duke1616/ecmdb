package dispatch

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	taskv1 "github.com/Duke1616/ecmdb/api/proto/gen/etask/task/v1"
	"github.com/Duke1616/ecmdb/internal/task/internal/domain"
	"github.com/Duke1616/ecmdb/internal/task/internal/repository"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/ecodeclub/ekit/slice"
	"github.com/google/uuid"
	"github.com/gotomicro/ego/core/elog"
)

type executeService struct {
	grpcClient taskv1.TaskServiceClient
	repo       repository.TaskRepository
	crypto     cryptox.Crypto
	logger     *elog.Component
}

func NewExecuteService(grpcClient taskv1.TaskServiceClient, repo repository.TaskRepository, crypto cryptox.Crypto) TaskDispatcher {
	return &executeService{
		grpcClient: grpcClient,
		repo:       repo,
		crypto:     crypto,
		logger:     elog.DefaultLogger.With(elog.FieldComponentName("executeService")),
	}
}

func (e *executeService) Dispatch(ctx context.Context, task domain.Task) error {
	// 准备参数
	args, _ := json.Marshal(task.Args)

	// 核心逻辑：始终生成一个明确的 Cron 点对点表达式
	// 如果是即时任务，默认设定为 2 秒后执行（给平台预留接收处理时间）实现“立即运行”的效果
	executeTime := time.Now().Add(time.Second * 2)
	if task.IsTiming && task.Timing.Stime > time.Now().UnixMilli() {
		executeTime = time.UnixMilli(task.Timing.Stime)
	}

	// 启动任务
	createTask, err := e.grpcClient.CreateTask(ctx, &taskv1.CreateTaskRequest{
		Name:     fmt.Sprintf("%s_%s", task.CodebookName, uuid.New().String()),
		Type:     taskv1.TaskType_ONE_TIME,
		CronExpr: executeTime.Format("05 04 15 02 01 ?"),
		GrpcConfig: &taskv1.GrpcConfig{
			ServiceName: task.Execute.ServiceName,
			HandlerName: task.Execute.Handler,
			Params: map[string]string{
				"task_id":   strconv.FormatInt(task.Id, 10),
				"code":      task.Code,
				"args":      string(args),
				"variables": e.decryptVariables(task.Variables),
			},
		},
	})
	if err != nil {
		e.logger.Error("调用分布式任务平台失败:", elog.FieldErr(err), elog.Int64("任务ID", task.Id))
		return fmt.Errorf("调用分布式任务平台失败: %w", err)
	}

	// 创建任务完成后，保存 ExternalId 到数据中
	return e.repo.UpdateExternalId(ctx, task.Id, strconv.FormatInt(createTask.Id, 10))
}

// decryptVariables 处理变量，进行解密
func (e *executeService) decryptVariables(req []domain.Variables) string {
	variables := slice.Map(req, func(idx int, src domain.Variables) domain.Variables {
		if src.Secret {
			val, er := e.crypto.Decrypt(src.Value)
			if er != nil {
				return domain.Variables{}
			}

			return domain.Variables{
				Key:    src.Key,
				Value:  val,
				Secret: src.Secret,
			}
		}

		return domain.Variables{
			Key:    src.Key,
			Value:  src.Value,
			Secret: src.Secret,
		}
	})

	jsonVar, _ := json.Marshal(variables)
	return string(jsonVar)
}
