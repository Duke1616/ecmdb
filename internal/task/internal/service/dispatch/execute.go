package dispatch

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	taskv1 "github.com/Duke1616/ecmdb/api/proto/gen/etask/task/v1"
	"github.com/Duke1616/ecmdb/internal/task/domain"
	"github.com/Duke1616/ecmdb/internal/task/internal/repository"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/Duke1616/etask/pkg/grpc/interceptors/jwt"
	"github.com/ecodeclub/ekit/slice"
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
	// 1. 核心属性预处理 (一次性完成序列化与解密)
	taskId := strconv.FormatInt(task.Id, 10)
	args := e.marshalArgs(task.Args)
	vars := e.decryptVariables(task.Variables)

	// 2. 极致性能哈希：基于确定顺序的字节流写入，生成确定性任务名称
	taskHash := e.sumHash(taskId, task.Code, args, vars)

	// 3. 启动分布式任务派发
	taskResult, err := e.grpcClient.CreateTask(jwt.SetTicketBizID(ctx), &taskv1.CreateTaskRequest{
		Name:     fmt.Sprintf("%s_%s", task.CodebookName, taskHash),
		Type:     taskv1.TaskType_ONE_TIME,
		CronExpr: e.calculateCronExpr(task),
		GrpcConfig: &taskv1.GrpcConfig{
			ServiceName: task.Target,
			HandlerName: task.Handler,
			Params: map[string]string{
				"task_id":   taskId,
				"code":      task.Code,
				"args":      args,
				"variables": vars,
			},
		},
	})

	if err != nil {
		e.logger.Error("调用分布式任务平台失败:", elog.FieldErr(err), elog.Int64("任务ID", task.Id))
		return fmt.Errorf("调用分布式任务平台失败: %w", err)
	}

	externalId := taskResult.Id
	switch taskResult.Code {
	case taskv1.TaskErrorCode_SUCCESS:
		// 正常成功，无需额外处理
	case taskv1.TaskErrorCode_DUPLICATE_NAME:
		// 处理业务冲突：如果名称已存在（索引冲突），尝试通过 Retry 接口获取 ID 并触发任务
		retryResp, retryErr := e.grpcClient.RetryTaskByName(ctx, &taskv1.RetryTaskByNameRequest{
			Name: fmt.Sprintf("%s_%s", task.CodebookName, taskHash),
		})
		if retryErr != nil {
			e.logger.Error("任务已存在且尝试重试获取 ID 失败:", elog.FieldErr(retryErr), elog.Int64("任务ID", task.Id))
			return fmt.Errorf("任务已存在且尝试重试获取 ID 失败: %w", retryErr)
		}
		externalId = retryResp.Id
	default:
		e.logger.Error("任务平台返回业务错误:", elog.Int32("code", int32(taskResult.Code)),
			elog.String("msg", taskResult.Message))
		return fmt.Errorf("任务平台业务错误: %s (code: %d)", taskResult.Message, taskResult.Code)
	}

	// 4. 绑定外部平台生成的任务实例 ID
	return e.repo.UpdateExternalId(ctx, task.Id, strconv.FormatInt(externalId, 10))
}

// sumHash 采用最快的字节直接写入方式计算 MD5 指纹，确保标识的可重复性
func (e *executeService) sumHash(ss ...string) string {
	h := md5.New()
	for _, s := range ss {
		h.Write([]byte(s))
		h.Write([]byte("|"))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// calculateCronExpr 构建准即时或定时的 Cron 调度表达式
func (e *executeService) calculateCronExpr(task domain.Task) string {
	executeTime := time.Now().Add(time.Second * 2)
	if task.IsTiming && task.ScheduledTime > time.Now().UnixMilli() {
		executeTime = time.UnixMilli(task.ScheduledTime)
	}
	return executeTime.Format("05 04 15 02 01 ?")
}

// marshalArgs 将动态执行参数序列化为 JSON
func (e *executeService) marshalArgs(args map[string]interface{}) string {
	res, _ := json.Marshal(args)
	return string(res)
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
