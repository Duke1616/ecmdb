package assignees

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/Duke1616/ecmdb/pkg/resolve"
)

// AppointResolver 指定人员解析器
// Values 中存放的是目标用户的 Username 列表（新版数据由前端直接传入）
type AppointResolver struct {
	userSvc user.Service
}

func NewAppointResolver(userSvc user.Service) *AppointResolver {
	return &AppointResolver{userSvc: userSvc}
}

// Name 返回该解析器覆盖的规则唯一标识
func (r *AppointResolver) Name() string {
	return string(easyflow.APPOINT)
}

func (r *AppointResolver) Resolve(ctx context.Context, target resolve.Target) ([]user.User, error) {
	if len(target.Values) == 0 {
		return nil, nil
	}

	users, err := r.userSvc.FindByUsernames(ctx, target.Values)
	if err != nil {
		return nil, err
	}

	return users, nil
}
