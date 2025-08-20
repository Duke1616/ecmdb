package service

import (
	"context"
	"encoding/json"

	"github.com/Duke1616/ecmdb/internal/task/internal/domain"
	"github.com/Duke1616/ecmdb/internal/worker"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/ecodeclub/ekit/slice"
	"github.com/spf13/viper"
)

type ExecService interface {
	Execute(ctx context.Context, task domain.Task) error
}
type execService struct {
	workerSvc worker.Service
}

func NewExecService(workerSvc worker.Service) ExecService {
	return &execService{workerSvc: workerSvc}
}

func (e *execService) Execute(ctx context.Context, task domain.Task) error {
	return e.workerSvc.Execute(ctx, worker.Execute{
		TaskId:    task.Id,
		Topic:     task.Topic,
		Code:      task.Code,
		Language:  task.Language,
		Args:      task.Args,
		Variables: decryptVariables(task.Variables),
	})
}

// decryptVariables 处理变量，进行解密
func decryptVariables(req []domain.Variables) string {
	variables := slice.Map(req, func(idx int, src domain.Variables) domain.Variables {
		if src.Secret {
			val, er := cryptox.DecryptAES[any](viper.Get("crypto_aes_key").(string), src.Value.(string))
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
