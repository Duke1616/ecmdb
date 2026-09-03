package web

import (
	"errors"

	"github.com/Duke1616/ecmdb/internal/domain"
	modelservice "github.com/Duke1616/ecmdb/internal/service/model"
	service "github.com/Duke1616/ecmdb/internal/service/model"
	relationservice "github.com/Duke1616/ecmdb/internal/service/relation"
	resourceservice "github.com/Duke1616/ecmdb/internal/service/resource"
	"github.com/Duke1616/ecmdb/pkg/contract/permission"
	"github.com/Duke1616/eiam/pkg/web/capability"
	"github.com/ecodeclub/ginx"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)


var _ ginx.Handler = &Handler{}

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

func (h *Handler) PublicRoutes(_ *gin.Engine) {}

func (h *Handler) IdentifyRoutes(_ *gin.Engine) {}

// PrivateRoutes 注册模型管理模块需要中心化登录及权限判定（由 EIAM SDK 统一拦截承载）的私有路由
func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/model")

	// ==========================================
	// 1. 模型分组管理接口 (派生模型分组层级)
	// ==========================================
	group := h.Sub("", "模型分组")

	// 创建模型分组
	g.POST("/group/create", group.Define("创建分组", "group_add").
		Bind(ginx.B[CreateModelGroupReq](h.CreateModelGroup)),
	)

	// 删除模型分组
	g.POST("/group/delete", group.Define("删除分组", "group_delete").
		Bind(ginx.B[DeleteModelGroup](h.DeleteModelGroup)),
	)

	// 重命名模型分组
	g.POST("/group/rename", group.Define("重命名分组", "group_rename").
		Bind(ginx.B[RenameModelGroupReq](h.RenameModelGroup)),
	)

	// ==========================================
	// 2. 模型核心基础操作接口
	// ==========================================

	// 创建模型
	g.POST("/create", h.Define("创建模型", "add").
		Bind(ginx.B[CreateModelReq](h.CreateModel)),
	)

	// 查询模型详情
	g.GET("/detail/:id", h.Define("模型详情", "get").
		Needs(permission.Attribute.View, permission.Relation.View, permission.Model.RelationView).
		Bind(ginx.W(h.DetailModel)),
	)

	// 删除模型
	g.POST("/delete", h.Define("删除模型", "delete").
		Bind(ginx.B[DeleteModelByUidReq](h.DeleteModelByUid)),
	)

	// 按分组查询模型列表
	g.POST("/list", h.Define("模型列表", "view").
		Bind(ginx.B[Page](h.ListModelsByGroup)),
	)

	// 按 UID 批量查询模型列表
	g.POST("by_uids", h.Define("按UID批量查询模型", "view_by_uids").
		NoSync().
		Bind(ginx.B[GetByUidsReq](h.GetByUids)),
	)

	// ==========================================
	// 3. 模型关联与拓扑图接口 (派生关联关系层级)
	// ==========================================
	relation := h.Sub("", "关联关系")

	// 查询模型关联拓扑图
	g.POST("/relation/graph", relation.Define("模型拓扑图", "relation_graph").
		Bind(ginx.B[Page](h.FindModelsGraph)),
	)

	// 创建模型关联关系
	g.POST("/relation/create", relation.Define("创建模型关联关系", "relation_add").
		Bind(ginx.B[CreateModelRelationReq](h.CreateModelRelation)),
	)

	// 查询模型拥有的所有关联信息
	g.POST("/relation/list", relation.Define("模型关联列表", "relation_view").
		Bind(ginx.B[ListModelRelationReq](h.ListModelUIDRelation)),
	)

	// 删除模型关联关系
	g.POST("/relation/delete", relation.Define("删除模型关联关系", "relation_delete").
		Bind(ginx.B[DeleteModelRelationReq](h.DeleteModelRelation)),
	)

	// 更新模型关联关系
	g.POST("/relation/update", relation.Define("更新模型关联关系", "relation_edit").
		Bind(ginx.B[UpdateModelRelationReq](h.UpdateModelRelation)),
	)

}


func (h *Handler) GetByUids(ctx *ginx.Context, req GetByUidsReq) (ginx.Result, error) {
	ms, err := h.svc.GetByUids(ctx.Context, req.Uids)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: RetrieveModelsListResp{
			Models: lo.Map(ms, func(m domain.Model, _ int) Model {
				return toModelVo(m)
			}),
		},
	}, nil
}

func (h *Handler) CreateModelGroup(ctx *ginx.Context, req CreateModelGroupReq) (ginx.Result, error) {
	id, err := h.mgSvc.Create(ctx.Context, domain.ModelGroup{
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

func (h *Handler) CreateModel(ctx *ginx.Context, req CreateModelReq) (ginx.Result, error) {
	// NOTE: 业务编排已下沉到 Service 层，Handler 仅负责协议适配
	id, err := h.svc.CreateModelWithDefaults(ctx.Context, domain.Model{
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

func (h *Handler) DetailModel(ctx *ginx.Context) (ginx.Result, error) {
	id, err := ctx.Param("id").AsInt64()
	if err != nil {
		return systemErrorResult, err
	}

	modelResp, err := h.svc.FindModelById(ctx.Context, id)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: h.toVo(modelResp),
		Msg:  "模型查找成功",
	}, nil
}

func (h *Handler) ListModelGroups(ctx *ginx.Context, req Page) (ginx.Result, error) {
	mgs, total, err := h.mgSvc.List(ctx.Context, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: RetrieveModelGroupsListResp{
			Total: total,
			Mgs: lo.Map(mgs, func(m domain.ModelGroup, _ int) ModelGroup {
				return h.toModelGroupVo(m)
			}),
		},
	}, nil
}

var (
	errModelRelationNotEmpty = errors.New("模型关联不为空")
	errModelResourceNotEmpty = errors.New("模型关联资产数据不为空")
)

func (h *Handler) DeleteModelByUid(ctx *ginx.Context, req DeleteModelByUidReq) (ginx.Result, error) {
	// NOTE: 依赖检查已下沉到 Service 层的 IDeleteModelDependencyChecker 机制
	count, err := h.svc.DeleteByModelUid(ctx.Context, req.ModelUid)
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

func (h *Handler) ListModelsByGroup(ctx *ginx.Context, req Page) (ginx.Result, error) {
	// 1. 先分页获取模型分组列表
	mgs, total, err := h.mgSvc.List(ctx.Context, req.Offset, req.Limit)
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
	mgids := lo.Map(mgs, func(src domain.ModelGroup, _ int) int64 {
		return src.ID
	})
	models, err := h.svc.ListModelByGroupIds(ctx.Context, mgids)
	if err != nil {
		return systemErrorResult, err
	}

	// 3. 提取所有模型 UID，实现精准按需统计
	modelUids := lo.Map(models, func(src domain.Model, _ int) string {
		return src.UID
	})

	// 4. 仅查询当前页面所涉模型的资产数量，彻底消除 MongoDB 全表无匹配大范围 Match & Group 统计灾难
	var resourceCount map[string]int
	if len(modelUids) > 0 {
		resourceCount, err = h.resourceSvc.CountByModelUids(ctx.Context, modelUids)
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

func (h *Handler) ListModels(ctx *ginx.Context, req Page) (ginx.Result, error) {
	models, total, err := h.svc.List(ctx.Context, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: RetrieveModelsListResp{
			Total: total,
			Models: lo.Map(models, func(m domain.Model, _ int) Model {
				return toModelVo(m)
			}),
		},
	}, nil
}

func (h *Handler) DeleteModelGroup(ctx *ginx.Context, req DeleteModelGroup) (ginx.Result, error) {
	count, err := h.mgSvc.Delete(ctx.Context, req.ID)
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

func (h *Handler) RenameModelGroup(ctx *ginx.Context, req RenameModelGroupReq) (ginx.Result, error) {
	_, err := h.mgSvc.Rename(ctx.Context, req.ID, req.Name)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "重命名模型分组成功",
	}, nil
}

// FindModelsGraph 查询模型拓扑图
func (h *Handler) FindModelsGraph(ctx *ginx.Context, req Page) (ginx.Result, error) {
	// TODO 为了后续加入 label 概念进行过滤先查询所有的模型
	// 查询所有模型
	models, err := h.svc.ListAll(ctx.Context)
	if err != nil {
		return systemErrorResult, err
	}
	mn := lo.Map(models, func(src domain.Model, _ int) ModelNode {
		data := make(map[string]string, 1)
		data["icon"] = src.Icon
		return ModelNode{
			ID:   src.UID,
			Text: src.Name,
			Data: data,
		}
	})

	// 取出所有的 uids
	modelUids := lo.Map(models, func(src domain.Model, _ int) string {
		return src.UID
	})

	// 查询包含的数据
	ds, err := h.RMSvc.FindModelDiagramBySrcUids(ctx.Context, modelUids)
	if err != nil {
		return systemErrorResult, err
	}

	ml := lo.Map(ds, func(src domain.ModelDiagram, _ int) ModelLine {
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

