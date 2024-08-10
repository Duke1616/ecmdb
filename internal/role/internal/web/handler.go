package web

import (
	"fmt"
	"github.com/Duke1616/ecmdb/internal/menu"
	"github.com/Duke1616/ecmdb/internal/policy"
	"github.com/Duke1616/ecmdb/internal/role/internal/domain"
	"github.com/Duke1616/ecmdb/internal/role/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc       service.Service
	menuSvc   menu.Service
	policySvc policy.Service
}

func NewHandler(svc service.Service, menuSvc menu.Service, policySvc policy.Service) *Handler {
	return &Handler{
		svc:       svc,
		menuSvc:   menuSvc,
		policySvc: policySvc,
	}
}

func (h *Handler) PublicRoutes(server *gin.Engine) {
	g := server.Group("/api/role")
	g.POST("/update", ginx.WrapBody[UpdateRoleReq](h.UpdateRole))
	g.POST("/create", ginx.WrapBody[CreateRoleReq](h.CreateRole))
	g.POST("/list", ginx.WrapBody[Page](h.ListRole))
	g.POST("/permission", ginx.WrapBody[RolePermissionReq](h.ListRolePermission))
	g.POST("/permission/add", ginx.WrapBody[AddPermissionForRoleReq](h.AddPermissionForRole))
}

func (h *Handler) CreateRole(ctx *gin.Context, req CreateRoleReq) (ginx.Result, error) {
	rId, err := h.svc.CreateRole(ctx, h.toDomain(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: rId,
	}, nil
}

func (h *Handler) AddPermissionForRole(ctx *gin.Context, req AddPermissionForRoleReq) (ginx.Result, error) {
	// 查询需要添加权限的菜单信息
	menus, err := h.menuSvc.FindByIds(ctx, req.MenuIds)
	if err != nil {
		return systemErrorResult, err
	}

	// 根据菜单信息，查询API接口权限
	var policies []policy.Policy
	for _, m := range menus {
		p := slice.Map(m.Endpoints, func(idx int, src menu.Endpoint) policy.Policy {
			return policy.Policy{
				Path:   src.Path,
				Method: src.Method,
				Effect: "allow",
			}
		})

		policies = append(policies, p...)
	}

	// 添加权限
	ok, err := h.policySvc.CreateOrUpdateFilteredPolicies(ctx, policy.Policies{
		RoleCode: req.RoleCode,
		Policies: policies,
	})
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: ok,
	}, nil
}

func (h *Handler) ListRolePermission(ctx *gin.Context, req RolePermissionReq) (ginx.Result, error) {
	// 获取角色拥有的权限
	ps, err := h.policySvc.GetPermissionsForRole(ctx, req.RoleCode)
	if err != nil {
		return systemErrorResult, err
	}

	// 生成唯一标识 map 结构
	pMap := slice.ToMap(ps, func(element policy.Policy) string {
		return fmt.Sprintf("%s:%s", element.Path, element.Method)
	})

	// 获取所有的菜单
	menus, err := h.menuSvc.ListMenu(ctx)
	if err != nil {
		return systemErrorResult, err
	}

	// 获取数据结构
	menuTree, authzIds, err := getPermission(menus, pMap)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: RetrieveRolePermission{
			AuthzIds: authzIds,
			Menu:     menuTree,
		},
	}, nil
}

func (h *Handler) UpdateRole(ctx *gin.Context, req UpdateRoleReq) (ginx.Result, error) {
	e := h.toDomainUpdate(req)
	t, err := h.svc.UpdateRole(ctx, e)

	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: t,
	}, nil
}

func (h *Handler) ListRole(ctx *gin.Context, req Page) (ginx.Result, error) {
	rts, total, err := h.svc.ListRole(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "查询角色列表成功",
		Data: RetrieveRoles{
			Total: total,
			Roles: slice.Map(rts, func(idx int, src domain.Role) Role {
				return h.toVoRole(src)
			}),
		},
	}, nil
}

func (h *Handler) toDomain(req CreateRoleReq) domain.Role {
	return domain.Role{
		Name:   req.Name,
		Code:   req.Code,
		Desc:   req.Desc,
		Status: req.Status,
	}
}

func (h *Handler) toVoRole(req domain.Role) Role {
	return Role{
		Id:     req.Id,
		Name:   req.Name,
		Code:   req.Code,
		Desc:   req.Desc,
		Status: req.Status,
	}
}

func (h *Handler) toDomainUpdate(req UpdateRoleReq) domain.Role {
	return domain.Role{
		Id:     req.Id,
		Name:   req.Name,
		Code:   req.Code,
		Desc:   req.Desc,
		Status: req.Status,
	}
}
