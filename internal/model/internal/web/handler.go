package web

import (
	"errors"

	"github.com/Duke1616/ecmdb/internal/model/internal/domain"
	"github.com/Duke1616/ecmdb/internal/model/internal/service"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/ginx/gctx"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc         service.Service
	mgSvc       service.MGService
	resourceSvc resource.EncryptedSvc
	RMSvc       relation.RMSvc
}

func NewHandler(svc service.Service, mgSvc service.MGService, rmSvc relation.RMSvc,
	resourceSvc resource.EncryptedSvc) *Handler {
	return &Handler{
		svc:         svc,
		mgSvc:       mgSvc,
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
	// NOTE: 业务编排已下沉到 Service 层，Handler 仅负责协议适配
	id, err := h.svc.CreateModelWithDefaults(ctx.Request.Context(), domain.Model{
		Name:    req.Name,
		GroupId: req.GroupId,
		UID:     req.UID,
		Icon:    req.Icon,
	})
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

var (
	errModelRelationNotEmpty = errors.New("模型关联不为空")
	errModelResourceNotEmpty = errors.New("模型关联资产数据不为空")
)

func (h *Handler) DeleteModelByUid(ctx *gin.Context, req DeleteModelByUidReq) (ginx.Result, error) {
	// NOTE: 依赖检查已下沉到 Service 层的 IDeleteModelDependencyChecker 机制
	count, err := h.svc.DeleteByModelUid(ctx.Request.Context(), req.ModelUid)
	if err != nil {
		if errors.Is(err, errModelRelationNotEmpty) {
			return ginx.Result{Code: 501002, Msg: err.Error()}, nil
		}
		if errors.Is(err, errModelResourceNotEmpty) {
			return ginx.Result{Code: 501003, Msg: err.Error()}, nil
		}
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: count,
		Msg:  "删除模型成功",
	}, nil
}

func (h *Handler) ListModelsByGroup(ctx *gin.Context, req Page) (ginx.Result, error) {
	// 1. 先分页获取模型分组列表
	mgs, _, err := h.mgSvc.List(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	// 2. 根据分组 ID 获取对应的模型列表
	mgids := slice.Map(mgs, func(idx int, src domain.ModelGroup) int64 {
		return src.ID
	})
	models, err := h.svc.ListModelByGroupIds(ctx, mgids)
	if err != nil {
		return systemErrorResult, err
	}

	// 3. 提取所有模型 UID，实现精准按需统计
	modelUids := slice.Map(models, func(idx int, src domain.Model) string {
		return src.UID
	})

	// 4. 仅查询当前页面所涉模型的资产数量，彻底消除 MongoDB 全表无匹配大范围 Match & Group 统计灾难
	var resourceCount map[string]int
	if len(modelUids) > 0 {
		resourceCount, err = h.resourceSvc.CountByModelUids(ctx, modelUids)
		if err != nil {
			return systemErrorResult, err
		}
	} else {
		resourceCount = make(map[string]int)
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
	mn := slice.Map(models, func(idx int, src domain.Model) ModelNode {
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

	ml := slice.Map(ds, func(idx int, src relation.ModelDiagram) ModelLine {
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
