package web

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/menu"
	"github.com/Duke1616/ecmdb/internal/permission/internal/service"
	"github.com/Duke1616/ecmdb/internal/policy"
	"github.com/Duke1616/ecmdb/internal/role"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

type Handler struct {
	svc       service.Service
	roleSvc   role.Service
	menuSvc   menu.Service
	policySvc policy.Service
}

func NewHandler(roleSvc role.Service, menuSvc menu.Service, policySvc policy.Service, svc service.Service) *Handler {
	return &Handler{
		svc:       svc,
		roleSvc:   roleSvc,
		menuSvc:   menuSvc,
		policySvc: policySvc,
	}
}

func (h *Handler) PublicRoutes(server *gin.Engine) {
	g := server.Group("/api/permission")
	g.POST("/get_user_menu", ginx.WrapBody[FindUserPermission](h.FindUserPermissionMenus))
}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/permission")
	g.POST("/list", ginx.WrapBody[RolePermissionReq](h.ListRolePermission))
	g.POST("/change", ginx.WrapBody[ChangePermissionForRoleReq](h.ChangePermissionForRoleReq))
}

func (h *Handler) ListRolePermission(ctx *gin.Context, req RolePermissionReq) (ginx.Result, error) {
	var (
		eg errgroup.Group
		r  role.Role
		ms []menu.Menu
	)
	eg.Go(func() error {
		var err error
		r, err = h.roleSvc.FindByRoleCode(ctx, req.RoleCode)
		return err
	})

	eg.Go(func() error {
		var err error
		ms, err = h.menuSvc.GetAllMenu(ctx)
		return err
	})
	if err := eg.Wait(); err != nil {
		return systemErrorResult, err
	}

	// 返回所有的菜单ids，并去除重复
	menuIds := h.getMenuIds([]role.Role{r})

	return ginx.Result{
		Data: RetrieveRolePermission{
			AuthzIds: menuIds,
			Menu:     GetMenusTree(ms),
		},
		Msg: "获取角色权限成功",
	}, nil
}

func (h *Handler) ChangePermissionForRoleReq(ctx *gin.Context, req ChangePermissionForRoleReq) (ginx.Result, error) {
	// 角色拥有菜单权限
	_, err := h.roleSvc.CreateOrUpdateRoleMenuIds(ctx, req.RoleCode, req.MenuIds)
	if err != nil {
		return systemErrorResult, err
	}

	// casbin 刷新后端接口权限
	err = h.svc.AddPermissionForRole(ctx, req.RoleCode, req.MenuIds)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "添加角色权限成功",
		Data: "ok",
	}, nil
}

func (h *Handler) FindUserPermissionMenus(ctx *gin.Context, req FindUserPermission) (ginx.Result, error) {
	// 获取用户所有的角色编码
	roleCodes, err := h.policySvc.GetRolesForUser(ctx, req.UserId)
	if err != nil {
		return systemErrorResult, err
	}

	// 获取用户拥有的角色信息
	roles, err := h.roleSvc.FindByIncludeCodes(ctx, roleCodes)
	if err != nil {
		return systemErrorResult, err
	}

	// 获取对应的相信菜单信息
	menuIds := h.getMenuIds(roles)

	// 生成树形结构
	tree, err := h.getUserMenuTree(ctx, menuIds)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: RetrieveUserPermission{
			Menus:     tree,
			RoleCodes: roleCodes,
		},
		Msg: "获取用户权限成功",
	}, nil
}

func (h *Handler) getUserMenuTree(ctx context.Context, menuIds []int64) ([]*Menu, error) {
	// 如果没有任何权限，直接返回
	if len(menuIds) == 0 {
		return nil, nil
	}

	// 获取对应的相信菜单信息
	menus, err := h.menuSvc.FindByIds(ctx, menuIds)
	if err != nil {
		return nil, err
	}

	// 需要处理 button 按钮
	tree := GetMenusTreeByButton(menus)

	// 生成树形结构
	return tree, nil
}

func (h *Handler) getMenuIds(roles []role.Role) []int64 {
	// 获取拥有的菜单ID， 进行去重
	menuIds := make([]int64, 0)
	uniquePermissions := make(map[int64]bool)
	for _, r := range roles {
		ms := slice.FilterMap(r.MenuIds, func(idx int, src int64) (int64, bool) {
			menuId := src
			if !uniquePermissions[menuId] {
				uniquePermissions[menuId] = true
				return menuId, true
			}

			return 0, false
		})

		menuIds = append(menuIds, ms...)
	}

	return menuIds
}

// FindUserPermissionMenus 版本一： 通过后端权限认证菜单
//func (h *Handler) ListRolePermission(ctx *gin.Context, req RolePermissionReq) (ginx.Result, error) {
//	// 获取角色拥有的权限
//	ps, err := h.policySvc.GetPermissionsForRole(ctx, req.RoleCode)
//	if err != nil {
//		return systemErrorResult, err
//	}
//
//	// 生成唯一标识 map 结构
//	pMap := slice.ToMap(ps, func(element policy.Policy) string {
//		return fmt.Sprintf("%s:%s", element.Path, element.Method)
//	})
//
//	// 获取所有的菜单
//	menus, err := h.menuSvc.ListMenu(ctx)
//	if err != nil {
//		return systemErrorResult, err
//	}
//
//	// 获取数据结构
//	menuTree, authzIds, err := getPermission(menus, pMap)
//	if err != nil {
//		return systemErrorResult, err
//	}
//
//	return ginx.Result{
//		Data: RetrieveRolePermission{
//			AuthzIds: authzIds,
//			Menu:     menuTree,
//		},
//		Msg: "获取角色权限成功",
//	}, nil
//}

// FindUserPermissionMenus 版本一： 通过后端权限认证菜单
//func (h *Handler) FindUserPermissionMenus(ctx *gin.Context, req FindUserPermissionMenus) (ginx.Result, error) {
//	var (
//		eg errgroup.Group
//		ms []menu.Menu
//		ps []policy.Policy
//	)
//	eg.Go(func() error {
//		var err error
//		ps, err = h.policySvc.GetImplicitPermissionsForUser(ctx, req.UserId)
//		return err
//	})
//
//	eg.Go(func() error {
//		var err error
//		ms, err = h.menuSvc.GetAllMenu(ctx)
//		return err
//	})
//	if err := eg.Wait(); err != nil {
//		return systemErrorResult, err
//	}
//
//	tree, err := GetPermissionMenusTree(ms, ps)
//	if err != nil {
//		return systemErrorResult, err
//	}
//
//	return ginx.Result{
//		Data: RetrieveUserPermission{
//			Menu: tree,
//		},
//		Msg: "获取用户权限成功",
//	}, nil
//}
