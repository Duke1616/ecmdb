package web

import (
	"errors"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/model/internal/domain"
	"github.com/Duke1616/ecmdb/internal/model/internal/service"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/ginx/gctx"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

type Handler struct {
	svc         service.Service
	mgSvc       service.MGService
	resourceSvc resource.EncryptedSvc
	RMSvc       relation.RMSvc
	AttrSvc     attribute.Service
}

func NewHandler(svc service.Service, mgSvc service.MGService, rmSvc relation.RMSvc, attrSvc attribute.Service,
	resourceSvc resource.EncryptedSvc) *Handler {
	return &Handler{
		svc:         svc,
		mgSvc:       mgSvc,
		AttrSvc:     attrSvc,
		RMSvc:       rmSvc,
		resourceSvc: resourceSvc,
	}
}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/model")
	// 模型 - 分组管理
	g.POST("/group/create", ginx.WrapBody[CreateModelGroupReq](h.CreateModelGroup))
	g.POST("/group/list", ginx.WrapBody[Page](h.ListModelGroups))
	g.POST("/group/delete", ginx.WrapBody[DeleteModelGroup](h.DeleteModelGroup))
	g.POST("/group/rename", ginx.WrapBody[RenameModelGroupReq](h.RenameModelGroup))
	// 模型 - 基础操作
	g.POST("/create", ginx.WrapBody[CreateModelReq](h.CreateModel))
	g.GET("/detail/:id", ginx.Wrap(h.DetailModel))
	g.POST("/list", ginx.WrapBody[Page](h.ListModels))
	g.POST("/delete", ginx.WrapBody[DeleteModelByUidReq](h.DeleteModelByUid))
	g.POST("/by_group", ginx.WrapBody[Page](h.ListModelsByGroup))

	// 根据 uids 查询模型
	g.POST("by_uids", ginx.WrapBody(h.GetByUids))
	// 模型 - 关联拓扑图
	g.POST("/relation/graph", ginx.WrapBody[Page](h.FindModelsGraph))
}

func (h *Handler) GetByUids(ctx *gin.Context, req GetByUidsReq) (ginx.Result, error) {
	ms, err := h.svc.GetByUids(ctx, req.Uids)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: RetrieveModelsListResp{
			Models: slice.Map(ms, func(idx int, m domain.Model) Model {
				return toModelVo(m)
			}),
		},
	}, nil
}

func (h *Handler) CreateModelGroup(ctx *gin.Context, req CreateModelGroupReq) (ginx.Result, error) {
	id, err := h.mgSvc.Create(ctx.Request.Context(), domain.ModelGroup{
		Name: req.Name,
	})

	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: id,
		Msg:  "添加模型分组成功",
	}, nil
}

func (h *Handler) CreateModel(ctx *gin.Context, req CreateModelReq) (ginx.Result, error) {
	id, err := h.svc.Create(ctx, domain.Model{
		Name:    req.Name,
		GroupId: req.GroupId,
		UID:     req.UID,
		Icon:    req.Icon,
	})

	if err != nil {
		return systemErrorResult, err
	}

	_, err = h.AttrSvc.CreateDefaultAttribute(ctx, req.UID)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: id,
		Msg:  "添加模型成功",
	}, nil
}

func (h *Handler) DetailModel(ctx *gin.Context) (ginx.Result, error) {
	g := gctx.Context{Context: ctx}
	id, err := g.Param("id").AsInt64()
	if err != nil {
		return systemErrorResult, err
	}

	modelResp, err := h.svc.FindModelById(ctx, id)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: h.toVo(modelResp),
		Msg:  "模型查找成功",
	}, nil
}

func (h *Handler) ListModelGroups(ctx *gin.Context, req Page) (ginx.Result, error) {
	mgs, total, err := h.mgSvc.List(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: RetrieveModelGroupsListResp{
			Total: total,
			Mgs: slice.Map(mgs, func(idx int, m domain.ModelGroup) ModelGroup {
				return h.toModelGroupVo(m)
			}),
		},
	}, nil
}

func (h *Handler) DeleteModelByUid(ctx *gin.Context, req DeleteModelByUidReq) (ginx.Result, error) {
	var (
		eg errgroup.Group
	)

	eg.Go(func() error {
		var err error
		mTotal, err := h.RMSvc.CountByModelUid(ctx, req.ModelUid)
		if err != nil {
			return err
		}

		if mTotal != 0 {
			return fmt.Errorf("模型关联不为空")
		}

		return err
	})

	eg.Go(func() error {
		var err error
		rTotal, err := h.resourceSvc.CountByModelUid(ctx, req.ModelUid)
		if err != nil {
			return err
		}

		if rTotal != 0 {
			return fmt.Errorf("模型关联资产数据不为空")
		}

		return err
	})

	if err := eg.Wait(); err != nil {
		return systemErrorResult, err
	}

	// TODO 删除模型，同步删除模型属性 +  模型属性分组
	count, err := h.svc.DeleteByModelUid(ctx, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: count,
		Msg:  "删除模型成功",
	}, nil
}

func (h *Handler) ListModelsByGroup(ctx *gin.Context, req Page) (ginx.Result, error) {
	var (
		eg            errgroup.Group
		mgs           []domain.ModelGroup
		models        []domain.Model
		total         int64
		resourceCount map[string]int
	)

	// 获取执行分组下的所有模型数据
	eg.Go(func() error {
		var err error
		mgs, total, err = h.mgSvc.List(ctx, req.Offset, req.Limit)
		if err != nil {
			return err
		}
		mgids := make([]int64, total)
		mgids = slice.Map(mgs, func(idx int, src domain.ModelGroup) int64 {
			return src.ID
		})

		models, err = h.svc.ListModelByGroupIds(ctx, mgids)
		if err != nil {
			return err
		}

		return err
	})

	// 查看所有模型拥有资产的数量
	eg.Go(func() error {
		var err error
		resourceCount, err = h.resourceSvc.CountByModelUids(ctx, []string{})
		if err != nil {
			return err
		}

		return err
	})

	if err := eg.Wait(); err != nil {
		return systemErrorResult, err
	}

	// 前端展示
	return ginx.Result{
		Data: RetrieveModelListByGroupId{
			Mgs: retrieveModelListByGroupId(models, mgs, resourceCount),
		},
	}, nil
}

func (h *Handler) ListModels(ctx *gin.Context, req Page) (ginx.Result, error) {
	models, total, err := h.svc.List(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: RetrieveModelsListResp{
			Total: total,
			Models: slice.Map(models, func(idx int, m domain.Model) Model {
				return toModelVo(m)
			}),
		},
	}, nil
}

func (h *Handler) DeleteModelGroup(ctx *gin.Context, req DeleteModelGroup) (ginx.Result, error) {
	count, err := h.mgSvc.Delete(ctx, req.ID)
	if err != nil {
		if errors.Is(err, service.ErrDependency) {
			return ginx.Result{
				Code: 501002,
				Msg:  err.Error(),
			}, nil
		}
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: count,
	}, nil
}

func (h *Handler) RenameModelGroup(ctx *gin.Context, req RenameModelGroupReq) (ginx.Result, error) {
	_, err := h.mgSvc.Rename(ctx, req.ID, req.Name)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "重命名模型分组成功",
	}, nil
}

// FindModelsGraph 查询模型拓扑图
func (h *Handler) FindModelsGraph(ctx *gin.Context, req Page) (ginx.Result, error) {
	// TODO 为了后续加入 label 概念进行过滤先查询所有的模型
	// 查询所有模型
	models, err := h.svc.ListAll(ctx)
	if err != nil {
		return systemErrorResult, err
	}
	mn := make([]ModelNode, len(models))
	mn = slice.Map(models, func(idx int, src domain.Model) ModelNode {
		data := make(map[string]string, 1)
		data["icon"] = src.Icon
		return ModelNode{
			ID:   src.UID,
			Text: src.Name,
			Data: data,
		}
	})

	// 取出所有的 uids
	modelUids := slice.Map(models, func(idx int, src domain.Model) string {
		return src.UID
	})

	// 查询包含的数据
	ds, err := h.RMSvc.FindModelDiagramBySrcUids(ctx, modelUids)
	if err != nil {
		return systemErrorResult, err
	}

	ml := make([]ModelLine, len(ds))
	ml = slice.Map(ds, func(idx int, src relation.ModelDiagram) ModelLine {
		return ModelLine{
			From: src.SourceModelUid,
			To:   src.TargetModelUid,
			Text: src.RelationTypeUid,
		}
	})

	return ginx.Result{
		Data: RetrieveRelationModelGraph{
			Nodes:  mn,
			Lines:  ml,
			RootId: "virtual",
		},
	}, nil
}

func (h *Handler) toVo(src domain.Model) Model {
	return Model{
		Id:      src.ID,
		Name:    src.Name,
		Icon:    src.Icon,
		UID:     src.UID,
		Builtin: src.Builtin,
	}
}

func (h *Handler) toModelGroupVo(m domain.ModelGroup) ModelGroup {
	return ModelGroup{
		Name: m.Name,
		Id:   m.ID,
	}
}
