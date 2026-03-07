package assignees

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Duke1616/ecmdb/internal/rota"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/Duke1616/ecmdb/pkg/resolve"
	"github.com/ecodeclub/ekit/slice"
)

// OnCallResolver 值班解析器
// target.Values 为排班组 ID 列表（string 化的 int64）
// 根据当前时间查询各排班组的值班人员，合并去重后返回完整用户信息
type OnCallResolver struct {
	rotaSvc rota.Service
	userSvc user.Service
}

func NewOnCallResolver(rotaSvc rota.Service, userSvc user.Service) *OnCallResolver {
	return &OnCallResolver{rotaSvc: rotaSvc, userSvc: userSvc}
}

// Name 返回该解析器覆盖的规则唯一标识
func (r *OnCallResolver) Name() string {
	return string(easyflow.ON_CALL)
}

func (r *OnCallResolver) Resolve(ctx context.Context, target resolve.Target) ([]user.User, error) {
	if len(target.Values) == 0 {
		return nil, nil
	}

	// 将 string ID 列表转为 int64
	ids, err := parseIDs(target.Values)
	if err != nil {
		return nil, fmt.Errorf("解析排班组 ID 失败: %w", err)
	}

	// 批量查询当前时段的值班排班
	schedules, err := r.rotaSvc.GetCurrentSchedulesByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("查询值班排班失败: %w", err)
	}

	// 合并所有排班组的值班人员并去重
	var members []string
	for _, sc := range schedules {
		members = slice.UnionSet(members, sc.RotaGroup.Members)
	}

	if len(members) == 0 {
		return nil, nil
	}

	users, err := r.userSvc.FindByUsernames(ctx, members)
	if err != nil {
		return nil, fmt.Errorf("查询值班人员信息失败: %w", err)
	}

	return users, nil
}

// parseIDs 将 []string 转换为 []int64
func parseIDs(values []string) ([]int64, error) {
	ids := make([]int64, 0, len(values))
	for _, v := range values {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("无效的 ID [%s]: %w", v, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}
