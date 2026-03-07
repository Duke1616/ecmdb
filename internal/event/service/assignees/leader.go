package assignees

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/department"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/Duke1616/ecmdb/pkg/resolve"
)

// LeaderResolver 部门领导解析器
// target.Values[0] 为发起人 username，通过其所在部门找到所有直属部门领导
type LeaderResolver struct {
	userSvc       user.Service
	departmentSvc department.Service
}

func NewLeaderResolver(userSvc user.Service, departmentSvc department.Service) *LeaderResolver {
	return &LeaderResolver{userSvc: userSvc, departmentSvc: departmentSvc}
}

// Name 返回该解析器覆盖的规则唯一标识
func (r *LeaderResolver) Name() string {
	return string(easyflow.LEADER)
}

func (r *LeaderResolver) Resolve(ctx context.Context, target resolve.Target) ([]user.User, error) {
	if len(target.Values) == 0 {
		return nil, fmt.Errorf("缺少发起人信息")
	}

	startUser, err := r.userSvc.FindByUsername(ctx, target.Values[0])
	if err != nil {
		return nil, fmt.Errorf("查询发起人失败: %w", err)
	}

	if startUser.DepartmentId == 0 {
		return nil, fmt.Errorf("发起人 [%s] 未分配部门", target.Values[0])
	}

	depart, err := r.departmentSvc.FindById(ctx, startUser.DepartmentId)
	if err != nil {
		return nil, err
	}

	return r.userSvc.FindByUsernames(ctx, depart.Leaders)
}
