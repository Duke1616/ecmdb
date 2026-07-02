package web

import (
	"errors"

	"github.com/Duke1616/ecmdb/internal/domain"
	modelservice "github.com/Duke1616/ecmdb/internal/service/model"
	service "github.com/Duke1616/ecmdb/internal/service/model"
	relationservice "github.com/Duke1616/ecmdb/internal/service/relation"
	resourceservice "github.com/Duke1616/ecmdb/internal/service/resource"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/Duke1616/eiam/pkg/web/capability"
	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/ginx/gctx"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

type Handler struct {
	svc         modelservice.Service
	mgSvc       modelservice.MGService
	resourceSvc resourceservice.EncryptedSvc
	RMSvc       relationservice.RelationModelService
	capability.IRegistry
}

func NewHandler(svc modelservice.Service, mgSvc modelservice.MGService, rmSvc relationservice.RelationModelService,
	resourceSvc resourceservice.EncryptedSvc) *Handler {
	return &Handler{
		svc:         svc,
		mgSvc:       mgSvc,
		RMSvc:       rmSvc,
		resourceSvc: resourceSvc,
		IRegistry:   capability.NewRegistry("cmdb", "model", "模型管理"),
	}
}

// PrivateRoutes 注册模型管理模块需要中心化登录及权限判定（由 EIAM SDK 统一拦截承载）的私有路由
func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/model")

	// ==========================================
	// 1. 模型分组管理接口
	// ==========================================

	// 创建模型分组
	g.POST("/group/create", h.Capability("创建分组", "group_add").
		Handle(ginx.WrapBody[CreateModelGroupReq](h.CreateModelGroup)),
	)

	// 删除模型分组
	g.POST("/group/delete", h.Capability("删除分组", "group_delete").
		Handle(ginx.WrapBody[DeleteModelGroup](h.DeleteModelGroup)),
	)

	// 重命名模型分组
	g.POST("/group/rename", h.Capability("重命名分组", "group_rename").
		Handle(ginx.WrapBody[RenameModelGroupReq](h.RenameModelGroup)),
	)

	// ==========================================
	// 2. 模型核心基础操作接口
	// ==========================================

	// 创建模型
	g.POST("/create", h.Capability("创建模型", "add").
		Handle(ginx.WrapBody[CreateModelReq](h.CreateModel)),
	)

	// 查询模型详情
	g.GET("/detail/:id", h.Capability("模型详情", "get").
		Needs("cmdb:attribute:view", "cmdb:relation:view", "cmdb:model-relation:view").
		Handle(ginx.Wrap(h.DetailModel)),
	)

	// 删除模型
	g.POST("/delete", h.Capability("删除模型", "delete").
		Handle(ginx.WrapBody[DeleteModelByUidReq](h.DeleteModelByUid)),
	)

	// 按分组查询模型列表
	g.POST("/list", h.Capability("模型列表", "view").
		Handle(ginx.WrapBody[Page](h.ListModelsByGroup)),
	)

	// 按 UID 批量查询模型列表
	g.POST("by_uids", h.Capability("按UID批量查询模型", "view_by_uids").
		NoSync().
		Handle(ginx.WrapBody(h.GetByUids)),
	)

	// ==========================================
	// 3. 模型关联与拓扑图接口
	// ==========================================
	// 查询模型关联拓扑图
	g.POST("/relation/graph", h.Capability("模型拓扑图", "relation_graph").
		Group("模型管理/模型拓扑").
		Handle(ginx.WrapBody[Page](h.FindModelsGraph)),
	)

	// ==========================================
	// 4. 模型关联关系管理接口
	// ==========================================

	// 创建模型关联关系
	g.POST("/relation/create", h.Capability("创建模型关联关系", "relation_add").
		Group("模型管理/关联关系").
		Handle(ginx.WrapBody[CreateModelRelationReq](h.CreateModelRelation)),
	)

	// 查询模型拥有的所有关联信息
	g.POST("/relation/list", h.Capability("模型关联列表", "relation_view").
		Group("模型管理/关联关系").
		Handle(ginx.WrapBody[ListModelRelationReq](h.ListModelUIDRelation)),
	)

	// 删除模型关联关系
	g.POST("/relation/delete", h.Capability("删除模型关联关系", "relation_delete").
		Group("模型管理/关联关系").
		Handle(ginx.WrapBody[DeleteModelRelationReq](h.DeleteModelRelation)),
	)

	// 更新模型关联关系
	g.POST("/relation/update", h.Capability("更新模型关联关系", "relation_edit").
		Group("模型管理/关联关系").
		Handle(ginx.WrapBody[UpdateModelRelationReq](h.UpdateModelRelation)),
	)
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
	mgs, total, err := h.mgSvc.List(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}
	if len(mgs) == 0 {
		return ginx.Result{
			Data: RetrieveModelGroupedListResp{
				Total:  total,
				Groups: []ModelGroupItem{},
				Models: []ModelSummaryVO{},
			},
		}, nil
	}

	// 2. 根据分组 ID 获取对应的模型列表
	mgids := lo.Map(mgs, func(src domain.ModelGroup, idx int) int64 {
		return src.ID
	})
	models, err := h.svc.ListModelByGroupIds(ctx, mgids)
	if err != nil {
		return systemErrorResult, err
	}

	// 3. 提取所有模型 UID，实现精准按需统计
	modelUids := lo.Map(models, func(src domain.Model, idx int) string {
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
		Data: RetrieveModelGroupedListResp{
			Total:  total,
			Groups: retrieveModelGroups(models, mgs),
			Models: retrieveModelSummaries(models, resourceCount),
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

	ml := slice.Map(ds, func(idx int, src domain.ModelDiagram) ModelLine {
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
