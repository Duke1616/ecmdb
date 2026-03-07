package assignees

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/Duke1616/ecmdb/pkg/resolve"
)

// FounderResolver 发起人解析器
// target.Values[0] 为发起人 username，直接查询并返回完整用户信息
type FounderResolver struct {
	userSvc user.Service
}

func NewFounderResolver(userSvc user.Service) *FounderResolver {
	return &FounderResolver{userSvc: userSvc}
}

// Name 返回该解析器覆盖的规则唯一标识
func (r *FounderResolver) Name() string {
	return string(easyflow.FOUNDER)
}

func (r *FounderResolver) Resolve(ctx context.Context, target resolve.Target) ([]user.User, error) {
	if len(target.Values) == 0 {
		return nil, fmt.Errorf("缺少发起人信息")
	}

	u, err := r.userSvc.FindByUsername(ctx, target.Values[0])
	if err != nil {
		return nil, fmt.Errorf("查询发起人失败: %w", err)
	}

	if u.Id != 0 {
		return []user.User{u}, nil
	}
	return nil, nil
}
