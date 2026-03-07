package assignees

import (
	"context"
	"fmt"
	"strconv"

	teamv1 "github.com/Duke1616/ecmdb/api/proto/gen/ealert/team"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/Duke1616/ecmdb/pkg/resolve"
	"github.com/ecodeclub/ekit/slice"
)

// TeamResolver 团队解析器
// target.Values 为团队 ID 列表 (string 化的 int64)
// 根据团队 ID 查询各团队的成员，合并去重后返回完整用户信息
type TeamResolver struct {
	teamSvc teamv1.TeamServiceClient
	userSvc user.Service
}

func NewTeamResolver(teamSvc teamv1.TeamServiceClient, userSvc user.Service) *TeamResolver {
	return &TeamResolver{teamSvc: teamSvc, userSvc: userSvc}
}

// Name 返回该解析器覆盖的规则唯一标识
func (r *TeamResolver) Name() string {
	return string(easyflow.TEAM)
}

func (r *TeamResolver) Resolve(ctx context.Context, target resolve.Target) ([]user.User, error) {
	if len(target.Values) == 0 {
		return nil, nil
	}

	// 将 string ID 列表转为 int64
	ids := make([]int64, 0, len(target.Values))
	for _, v := range target.Values {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("无效的团队 ID [%s]: %w", v, err)
		}
		ids = append(ids, id)
	}

	// 批量查询团队信息
	resp, err := r.teamSvc.GetTeamByIds(ctx, &teamv1.GetTeamByIdsRequest{
		Ids: ids,
	})
	if err != nil {
		return nil, fmt.Errorf("查询团队信息失败: %w", err)
	}

	// 合并所有团队的成员名单并去重
	var usernames []string
	for _, team := range resp.Teams {
		usernames = slice.UnionSet(usernames, team.Members)
	}

	if len(usernames) == 0 {
		return nil, nil
	}

	// 根据用户名批量查询人员详情
	users, err := r.userSvc.FindByUsernames(ctx, usernames)
	if err != nil {
		return nil, fmt.Errorf("查询团队成员详情失败: %w", err)
	}

	return users, nil
}
