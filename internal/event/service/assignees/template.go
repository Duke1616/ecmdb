package assignees

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/Duke1616/ecmdb/pkg/resolve"
)

// TemplateResolver 模版字段解析器
// 调用方负责从工单中读取模版字段值并解析成 username 列表，填入 target.Values
// 本 resolver 只负责根据 username 列表查询并返回完整用户信息
type TemplateResolver struct {
	userSvc user.Service
}

func NewTemplateResolver(userSvc user.Service) *TemplateResolver {
	return &TemplateResolver{userSvc: userSvc}
}

// Name 返回该解析器覆盖的规则唯一标识
func (r *TemplateResolver) Name() string {
	return string(easyflow.TEMPLATE)
}

func (r *TemplateResolver) Resolve(ctx context.Context, target resolve.Target) ([]user.User, error) {
	return r.userSvc.FindByUsernames(ctx, target.Values)
}
