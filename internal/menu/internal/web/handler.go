package web

import (
	"errors"

	"github.com/Duke1616/ecmdb/internal/menu/internal/domain"
	"github.com/Duke1616/ecmdb/internal/menu/internal/errs"
	"github.com/Duke1616/ecmdb/internal/menu/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc service.Service
}

func NewHandler(svc service.Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/menu")
	// 创建接口
	g.POST("/create", ginx.WrapBody[CreateMenuReq](h.CreateMenu))

	// 修改接口
	g.POST("/update", ginx.WrapBody[UpdateMenuReq](h.UpdateMenu))

	// 删除接口
	g.POST("/delete", ginx.WrapBody[DeleteMenuReq](h.DeleteMenu))

	// 获取菜单列表，树形结构
	g.POST("/list/tree", ginx.Wrap(h.ListMenuTree))

	// 根据平台获取
	g.POST("/list/tree/by_platform", ginx.WrapBody[ListByPlatformReq](h.ListByPlatform))

	// 改变 API 接口
	g.POST("/change_endpoints", ginx.WrapBody[ChangeEndpointsReq](h.ChangeEndpointsReq))

	// 菜单拖拽排序
	g.POST("/sort", ginx.WrapBody[SortMenuReq](h.SortMenu))
}

func (h *Handler) ChangeEndpointsReq(ctx *gin.Context, req ChangeEndpointsReq) (ginx.Result, error) {
	count, err := h.svc.ChangeMenuEndpoints(ctx, req.ID, domain.Action(req.Action), slice.Map(req.Endpoints,
		func(idx int, src Endpoint) domain.Endpoint {
			return h.toEndpoint(src)
		}))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "success",
		Data: count,
	}, nil
}

func (h *Handler) CreateMenu(ctx *gin.Context, req CreateMenuReq) (ginx.Result, error) {
	eId, err := h.svc.CreateMenu(ctx, h.toDomain(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: eId,
	}, nil
}

func (h *Handler) ListByPlatform(ctx *gin.Context, req ListByPlatformReq) (ginx.Result, error) {
	ms, err := h.svc.ListByPlatform(ctx, req.Platform)
	if err != nil {
		return ginx.Result{}, err
	}

	return ginx.Result{
		Msg:  "OK",
		Data: GetMenusTree(ms),
	}, nil
}

func (h *Handler) DeleteMenu(ctx *gin.Context, req DeleteMenuReq) (ginx.Result, error) {
	count, err := h.svc.DeleteMenu(ctx, req.Id)
	if err != nil {
		if errors.Is(err, errs.MenuHasChildren) {
			return menuHasChildrenResult, nil
		}
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: count,
	}, nil
}

func (h *Handler) ListMenuTree(ctx *gin.Context) (ginx.Result, error) {
	ms, err := h.svc.ListMenu(ctx)
	if err != nil {
		return ginx.Result{}, err
	}

	return ginx.Result{
		Msg:  "OK",
		Data: GetMenusTree(ms),
	}, nil
}

// UpdateMenu 修改菜单信息
func (h *Handler) UpdateMenu(ctx *gin.Context, req UpdateMenuReq) (ginx.Result, error) {
	e := h.toDomainUpdate(req)
	t, err := h.svc.UpdateMenu(ctx, e)

	// 当修改发生变换的时候，向Kafka推送一条信息，添加对应的权限
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: t,
	}, nil
}

func (h *Handler) SortMenu(ctx *gin.Context, req SortMenuReq) (ginx.Result, error) {
	err := h.svc.Sort(ctx, req.ID, req.TargetPid, req.TargetPosition)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (h *Handler) toDomain(req CreateMenuReq) domain.Menu {
	return domain.Menu{
		Pid:       req.Pid,
		Path:      req.Path,
		Sort:      req.Sort,
		Type:      domain.Type(req.Type),
		Component: req.Component,
		Redirect:  req.Redirect,
		Name:      req.Name,
		Status:    domain.Status(req.Status),
		Meta: domain.Meta{
			Title:       req.Meta.Title,
			IsHidden:    req.Meta.IsHidden,
			IsAffix:     req.Meta.IsAffix,
			IsKeepAlive: req.Meta.IsKeepAlive,
			Platforms:   req.Meta.Platforms,
			Icon:        req.Meta.Icon,
		},
		Endpoints: slice.Map(req.Endpoints, func(idx int, src Endpoint) domain.Endpoint {
			return domain.Endpoint{
				Path:   src.Path,
				Method: src.Method,
				Desc:   src.Desc,
			}
		}),
	}
}

func (h *Handler) toEndpoint(req Endpoint) domain.Endpoint {
	return domain.Endpoint{
		Path:     req.Path,
		Method:   req.Method,
		Resource: req.Resource,
		Desc:     req.Desc,
	}
}

func (h *Handler) toDomainUpdate(req UpdateMenuReq) domain.Menu {
	return domain.Menu{
		Id:        req.Id,
		Pid:       req.Pid,
		Path:      req.Path,
		Sort:      req.Sort,
		Type:      domain.Type(req.Type),
		Component: req.Component,
		Redirect:  req.Redirect,
		Name:      req.Name,
		Status:    domain.Status(req.Status),
		Meta: domain.Meta{
			Title:       req.Meta.Title,
			IsHidden:    req.Meta.IsHidden,
			Platforms:   req.Meta.Platforms,
			IsAffix:     req.Meta.IsAffix,
			IsKeepAlive: req.Meta.IsKeepAlive,
			Icon:        req.Meta.Icon,
		},
		Endpoints: slice.Map(req.Endpoints, func(idx int, src Endpoint) domain.Endpoint {
			return domain.Endpoint{
				Path:   src.Path,
				Method: src.Method,
				Desc:   src.Desc,
			}
		}),
	}
}
