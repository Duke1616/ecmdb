package web

import (
	"github.com/Duke1616/ecmdb/internal/menu/internal/domain"
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

func (h *Handler) PublicRoutes(server *gin.Engine) {
	g := server.Group("/api/menu")
	g.POST("/create", ginx.WrapBody[CreateMenuReq](h.CreateMenu))
	g.POST("/update", ginx.WrapBody[UpdateMenuReq](h.UpdateMenu))
	g.POST("/list/tree", ginx.Wrap(h.ListMenuTree))
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

func (h *Handler) ListMenuTree(ctx *gin.Context) (ginx.Result, error) {
	ms, err := h.svc.ListMenu(ctx)
	if err != nil {
		return ginx.Result{}, err
	}
	menus := slice.Map(ms, func(idx int, src domain.Menu) *Menu {
		return h.toVoMenu(src)
	})

	tree, err := GetMenusTree(menus)
	if err != nil {
		return ginx.Result{}, err
	}
	return ginx.Result{
		Code: 0,
		Msg:  "",
		Data: tree,
	}, nil
}

func (h *Handler) UpdateMenu(ctx *gin.Context, req UpdateMenuReq) (ginx.Result, error) {
	e := h.toDomainUpdate(req)

	t, err := h.svc.UpdateMenu(ctx, e)

	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: t,
	}, nil
}

func GetMenusTree(menus []*Menu) (list []*Menu, err error) {
	//生成map， 方便查找获取对象
	allMap := map[int64]*Menu{}
	list = []*Menu{}

	for k, cat := range menus {
		menus[k].Children = []*Menu{}
		allMap[cat.Id] = menus[k]
		//记录顶级分类数据
		if cat.Pid == 0 {
			list = append(list, menus[k])
		}
	}

	//形成tree
	for k, cat := range menus {
		_, ok := allMap[cat.Pid]
		if ok {
			//如果父级别数据存在，添加到Children
			allMap[cat.Pid].Children = append(allMap[cat.Pid].Children, menus[k])
			//利用指针逻辑，map中数据和列表中原始对象为统一指针。指向同一内存地址，如此对map中数据操作，也相当于对原始数据操作。
		}
	}

	return
}

func (h *Handler) toDomain(req CreateMenuReq) domain.Menu {
	return domain.Menu{
		Pid:           req.Pid,
		Path:          req.Path,
		Sort:          req.Sort,
		Type:          domain.Type(req.Type),
		Component:     req.Component,
		Redirect:      req.Redirect,
		Name:          req.Name,
		ComponentPath: req.ComponentPath,
		Status:        domain.Status(req.Status),
		Meta: domain.Meta{
			Title:       req.Meta.Title,
			IsHidden:    req.Meta.IsHidden,
			IsAffix:     req.Meta.IsAffix,
			IsKeepAlive: req.Meta.IsKeepAlive,
			Icon:        req.Meta.Icon,
		},
		Endpoints: slice.Map(req.Endpoints, func(idx int, src Endpoint) domain.Endpoint {
			return domain.Endpoint{
				Id:     src.Id,
				Path:   src.Path,
				Method: src.Method,
				Desc:   src.Desc,
			}
		}),
	}
}

func (h *Handler) toDomainUpdate(req UpdateMenuReq) domain.Menu {
	return domain.Menu{
		Id:            req.Id,
		Pid:           req.Pid,
		Path:          req.Path,
		Sort:          req.Sort,
		Type:          domain.Type(req.Type),
		Component:     req.Component,
		Redirect:      req.Redirect,
		Name:          req.Name,
		ComponentPath: req.ComponentPath,
		Status:        domain.Status(req.Status),
		Meta: domain.Meta{
			Title:       req.Meta.Title,
			IsHidden:    req.Meta.IsHidden,
			IsAffix:     req.Meta.IsAffix,
			IsKeepAlive: req.Meta.IsKeepAlive,
			Icon:        req.Meta.Icon,
		},
		Endpoints: slice.Map(req.Endpoints, func(idx int, src Endpoint) domain.Endpoint {
			return domain.Endpoint{
				Id:     src.Id,
				Path:   src.Path,
				Method: src.Method,
				Desc:   src.Desc,
			}
		}),
	}
}

func (h *Handler) toVoMenu(req domain.Menu) *Menu {
	return &Menu{
		Id:            req.Id,
		Pid:           req.Pid,
		Path:          req.Path,
		Sort:          req.Sort,
		Name:          req.Name,
		Redirect:      req.Redirect,
		Type:          req.Type.ToUint8(),
		Component:     req.Component,
		ComponentPath: req.ComponentPath,
		Status:        req.Status.ToUint8(),
		Meta: Meta{
			Title:       req.Meta.Title,
			IsHidden:    req.Meta.IsHidden,
			IsAffix:     req.Meta.IsAffix,
			IsKeepAlive: req.Meta.IsKeepAlive,
			Icon:        req.Meta.Icon,
		},
		Endpoints: slice.Map(req.Endpoints, func(idx int, src domain.Endpoint) Endpoint {
			return Endpoint{
				Id:     src.Id,
				Path:   src.Path,
				Method: src.Method,
				Desc:   src.Desc,
			}
		}),
		Children: []*Menu{},
	}
}
