package result

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/task"
)

type FetcherResult interface {
	FetchResult(ctx context.Context, instanceID int, nodeID string) (map[string]interface{}, error)
}

type TaskResult struct {
	taskSvc task.Service
}

func NewResult(taskSvc task.Service) *TaskResult {
	return &TaskResult{
		taskSvc: taskSvc,
	}
}

func (r *TaskResult) FetchResult(ctx context.Context, instanceID int, nodeID string) (map[string]interface{}, error) {
	result, err := r.taskSvc.FindTaskResult(ctx, instanceID, nodeID)
	if err != nil {
		return nil, err
	}

	if result.WantResult == "" {
		return nil, fmt.Errorf("返回值为空, 不做任何数据处理")
	}

	var wantResult map[string]interface{}
	err = json.Unmarshal([]byte(result.WantResult), &wantResult)
	if err != nil {
		return nil, err
	}

	return wantResult, nil
}
