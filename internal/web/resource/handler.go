package web

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Duke1616/ecmdb/internal/domain"
	attributeservice "github.com/Duke1616/ecmdb/internal/service/attribute"
	modelservice "github.com/Duke1616/ecmdb/internal/service/model"
	relationservice "github.com/Duke1616/ecmdb/internal/service/relation"
	service "github.com/Duke1616/ecmdb/internal/service/resource"
	"github.com/Duke1616/ecmdb/pkg/contract/permission"
	"github.com/Duke1616/eiam/pkg/web/capability"
	"github.com/ecodeclub/ginx"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

var _ ginx.Handler = &Handler{}

type Handler struct {
	svc      service.EncryptedSvc
	attrSvc  attributeservice.Service
	modelSvc modelservice.Service
	rrSvc    relationservice.RelationResourceService
	capability.IRegistry
}

func NewHandler(svc service.EncryptedSvc, attributeSvc attributeservice.Service, modelSvc modelservice.Service, rrSvc relationservice.RelationResourceService) *Handler {
	return &Handler{
		svc:       svc,
		attrSvc:   attributeSvc,
		modelSvc:  modelSvc,
		rrSvc:     rrSvc,
		IRegistry: capability.NewRegistry("cmdb", "resource", "资产仓库"),
	}
}

func (h *Handler) PublicRoutes(_ *gin.Engine) {}

func (h *Handler) IdentifyRoutes(_ *gin.Engine) {}

// PrivateRoutes 注册资源管理模块需要中心化登录及权限判定（由 EIAM SDK 统一拦截承载）的私有路由
func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/resource")

	// ==========================================
	// 1. 资产基础操作接口
	// ==========================================

	// 创建资产
	g.POST("/create", h.Define("创建资产", "add").
		Needs(permission.Tools.PutPresignedUrl).
		Bind(ginx.B[CreateResourceReq](h.CreateResource)),
	)

	// 查询资产详情
	g.POST("/detail", h.Define("资产详情", "get").
		Bind(ginx.B[DetailResourceReq](h.DetailResource)),
	)

	// 根据模型 UID 查询资产列表
	g.POST("/list", h.Define("资产列表", "view").
		Needs(permission.Model.View, permission.Attribute.View, permission.Tools.GetPresignedUrl, permission.Plugin.Actions).
		Bind(ginx.B[ListResourceReq](h.ListResource)),
	)

	// 删除资产
	g.POST("/delete", h.Define("删除资产", "delete").
		Bind(ginx.B[DeleteResourceReq](h.DeleteResource)),
	)

	// 修改资产信息
	g.POST("/update", h.Define("修改资产", "edit").
		Needs(permission.Tools.PutPresignedUrl).
		Bind(ginx.B[UpdateResourceReq](h.UpdateResource)),
	)

	// 设置自定义属性
	g.POST("/set_custom_field", h.Define("设置自定义属性", "edit_custom_field").
		Bind(ginx.B[SetCustomFieldReq](h.SetCustomField)),
	)

	// 批量查询资产
	g.POST("/list/ids", h.Define("批量查询资产", "view_by_ids").
		NoSync().
		Bind(ginx.B[ListResourceByIdsReq](h.ListResourceByIds)),
	)

	// 查询加密字段数据
	g.POST("/secure", h.Define("查询加密字段", "get_secure").
		Bind(ginx.B[FindSecureReq](h.FindSecureData)),
	)

	// ==========================================
	// 2. 资产关联拓扑与关系管理 (派生关联关系层级)
	// ==========================================
	relation := h.Sub("", "关联关系")

	// 查询可关联的资产列表
	g.POST("/relation/can_be_related", relation.Define("查询可关联的资产列表", "view_can_be_related").
		NoSync().
		Bind(ginx.B[ListCanBeRelatedReqByModel](h.ListCanBeFilterRelated)),
	)

	// 查询资产关联拓扑图
	g.POST("/relation/graph", relation.Define("资产关联拓扑图", "view_relation_graph").
		Needs(permission.Resource.AddRelationLeft, permission.Resource.AddRelationRight).
		Bind(ginx.B[ListDiagramReq](h.FindAllGraph)),
	)

	// 拓扑图向左拓展
	g.POST("/relation/graph/add/left", relation.Define("拓扑图向左拓展", "add_relation_left").
		NoSync().
		Bind(ginx.B[ListDiagramReq](h.FindLeftGraph)),
	)

	// 拓扑图向右拓展
	g.POST("/relation/graph/add/right", relation.Define("拓扑图向右拓展", "add_relation_right").
		NoSync().
		Bind(ginx.B[ListDiagramReq](h.FindRightGraph)),
	)

	// 创建资源关联关系
	g.POST("/relation/create", relation.Define("创建资产关系", "relation_add").
		Needs(permission.Resource.ViewCanBeRelated).
		Bind(ginx.B[CreateResourceRelationReq](h.CreateResourceRelation)),
	)

	// 所有资产关系聚合查询
	g.POST("/relation/pipeline/all", relation.Define("所有资产关系聚合查询", "view_relation_all").
		Needs(permission.Relation.View, permission.Model.RelationView, permission.Model.ViewByUids,
			permission.Attribute.ViewFields, permission.Resource.ViewByIds).
		Bind(ginx.B[ListResourceDiagramReq](h.ListAllAggregated)),
	)

	// 删除资产关系
	g.POST("/relation/delete", relation.Define("删除资产关系", "relation_delete").
		Bind(ginx.B[DeleteResourceRelationReq](h.DeleteResourceRelation)),
	)

	// ==========================================
	// 3. 全局搜索 (派生全局搜索层级)
	// ==========================================
	search := h.Sub("", "全局搜索")

	// 全文检索资产
	g.POST("/search", search.Define("全文检索资产", "search").
		Bind(ginx.B[SearchReq](h.Search)),
	)
}

func (h *Handler) CreateResource(ctx *ginx.Context, req CreateResourceReq) (ginx.Result, error) {
	id, err := h.svc.CreateResource(ctx.Context, h.toCreateDomain(req))

	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: id,
	}, nil
}

func (h *Handler) DetailResource(ctx *ginx.Context, req DetailResourceReq) (ginx.Result, error) {
	resp, err := h.svc.FindResourceById(ctx.Context, nil, req.ID)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: resp,
		Msg:  "查看资源详情成功",
	}, nil
}

func (h *Handler) SetCustomField(ctx *ginx.Context, req SetCustomFieldReq) (ginx.Result, error) {
	count, err := h.svc.SetCustomField(ctx.Context, req.Id, req.Field, req.Data)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: count,
	}, nil
}

func (h *Handler) ListResource(ctx *ginx.Context, req ListResourceReq) (ginx.Result, error) {
	resp, total, err := h.svc.ListResource(ctx.Context, nil, req.ModelUid, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	rs := lo.Map(resp, func(src domain.Resource, _ int) Resource {
		return Resource{
			ID:       src.ID,
			Name:     src.Name,
			ModelUID: src.ModelUID,
			Data:     src.Data,
		}
	})

	return ginx.Result{
		Data: RetrieveResources{
			Resources: rs,
			Total:     total,
		},
		Msg: "查看资源列表成功",
	}, nil
}

func (h *Handler) UpdateResource(ctx *ginx.Context, req UpdateResourceReq) (ginx.Result, error) {
	resource := h.toUpdateDomain(req)
	t, err := h.svc.UpdateResource(ctx.Context, resource)

	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: t,
	}, nil
}

func (h *Handler) ListCanBeFilterRelated(ctx *ginx.Context, req ListCanBeRelatedReqByModel) (ginx.Result, error) {
	if req.RelationName == "" {
		return systemErrorResult, fmt.Errorf("关联名称为空")
	}

	// NOTE: 通过关系服务获取对端模型及已关联资源 ID，彻底规避下划线 Split 导致的模型解析错误
	mUid, excludeIds, err := h.rrSvc.GetCanBeRelatedContext(ctx.Context, req.ModelUid, req.RelationName, req.ResourceId)
	if err != nil {
		return systemErrorResult, err
	}

	fields, err := h.attrSvc.SearchAttributeFieldsByModelUid(ctx.Context, mUid)

	if err != nil {
		return systemErrorResult, err
	}

	// 排除已关联数据, 并且进行过滤，返回未关联数据
	rrs, total, err := h.svc.ListExcludeAndFilterResourceByIds(ctx.Context, fields, mUid, req.Offset, req.Limit, excludeIds,
		domain.Condition{
			Name:      req.FilterName,
			Condition: req.FilterCondition,
			Input:     req.FilterInput,
		})
	if err != nil {
		return systemErrorResult, err
	}

	rs := lo.Map(rrs, func(src domain.Resource, _ int) Resource {
		return Resource{
			ID:       src.ID,
			Name:     src.Name,
			ModelUID: src.ModelUID,
			Data:     src.Data,
		}
	})

	return ginx.Result{
		Data: RetrieveResources{
			Resources: rs,
			Total:     total,
		},
	}, nil
}

func (h *Handler) graphModels(ctx context.Context, resources []domain.Resource, rootModelUID string) ([]GraphModel, error) {
	modelUIDs := lo.Uniq(append(lo.Map(resources, func(src domain.Resource, _ int) string {
		return src.ModelUID
	}), rootModelUID))
	modelUIDs = lo.Filter(modelUIDs, func(uid string, _ int) bool {
		return uid != ""
	})
	if len(modelUIDs) == 0 {
		return []GraphModel{}, nil
	}

	models, err := h.modelSvc.GetByUids(ctx, modelUIDs)
	if err != nil {
		return nil, err
	}

	return lo.Map(models, func(src domain.Model, _ int) GraphModel {
		return GraphModel{
			ModelUID:  src.UID,
			ModelName: src.Name,
			Icon:      src.Icon,
		}
	}), nil
}

// assembleGraph 抽取统一的图拓扑结构组装器，消减拓扑图组装中重复的资源查询、图例匹配与节点包装逻辑
func (h *Handler) assembleGraph(
	ctx context.Context,
	req ListDiagramReq,
	rrs []domain.ResourceRelation,
	nodeIDs []int64,
	posResolver func(resourceID int64) string,
	includeRootNode bool,
) (RetrieveGraph, error) {
	lines := lo.Map(rrs, func(src domain.ResourceRelation, _ int) Line {
		return Line{
			From: strconv.FormatInt(src.SourceResourceID, 10),
			To:   strconv.FormatInt(src.TargetResourceID, 10),
		}
	})

	rs, err := h.svc.ListResourceByIds(ctx, []string{"name"}, nodeIDs)
	if err != nil {
		return RetrieveGraph{}, err
	}

	models, err := h.graphModels(ctx, rs, req.ModelUid)
	if err != nil {
		return RetrieveGraph{}, err
	}

	nodes := lo.Map(rs, func(src domain.Resource, _ int) Node {
		data := make(map[string]any, 3)
		data["model_uid"] = src.ModelUID
		data["isNeedLoadDataFromRemoteServer"] = true
		data["childrenLoaded"] = false

		return Node{
			ID:                   strconv.FormatInt(src.ID, 10),
			Text:                 src.Name,
			Data:                 data,
			ExpandHolderPosition: posResolver(src.ID),
			Expanded:             false,
		}
	})

	if includeRootNode {
		nodes = append(nodes, Node{
			ID:       strconv.FormatInt(req.ResourceId, 10),
			Text:     req.ResourceName,
			Expanded: true,
			Data: map[string]any{
				"model_uid": req.ModelUid,
			},
		})
	}

	return RetrieveGraph{
		Lines:  lines,
		Nodes:  nodes,
		RootId: strconv.FormatInt(req.ResourceId, 10),
		Models: models,
	}, nil
}

func (h *Handler) FindAllGraph(ctx *ginx.Context, req ListDiagramReq) (ginx.Result, error) {
	maxDepth := req.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 3
	}
	graph, err := h.rrSvc.ListRecursiveDiagram(ctx.Context, req.ModelUid, req.ResourceId, maxDepth)
	if err != nil {
		return systemErrorResult, err
	}

	rrs := append(graph.SRC, graph.DST...)
	srcId := lo.Map(graph.SRC, func(src domain.ResourceRelation, _ int) int64 {
		return src.TargetResourceID
	})
	dstId := lo.Map(graph.DST, func(src domain.ResourceRelation, _ int) int64 {
		return src.SourceResourceID
	})
	ids := lo.Uniq(append(srcId, dstId...))

	res, err := h.assembleGraph(ctx.Context, req, rrs, ids, func(id int64) string {
		if lo.Contains(srcId, id) {
			return "right"
		}
		return "left"
	}, true)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{Data: res}, nil
}

func (h *Handler) FindLeftGraph(ctx *ginx.Context, req ListDiagramReq) (ginx.Result, error) {
	graphLeft, _, err := h.rrSvc.ListDstResources(ctx.Context, req.ModelUid, req.ResourceId)
	if err != nil {
		return systemErrorResult, err
	}

	srcIds := lo.Uniq(lo.Map(graphLeft, func(src domain.ResourceRelation, _ int) int64 {
		return src.SourceResourceID
	}))

	res, err := h.assembleGraph(ctx.Context, req, graphLeft, srcIds, func(_ int64) string {
		return "left"
	}, false)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{Data: res}, nil
}

func (h *Handler) FindRightGraph(ctx *ginx.Context, req ListDiagramReq) (ginx.Result, error) {
	graphRight, _, err := h.rrSvc.ListSrcResources(ctx.Context, req.ModelUid, req.ResourceId)
	if err != nil {
		return systemErrorResult, err
	}

	dstIds := lo.Uniq(lo.Map(graphRight, func(src domain.ResourceRelation, _ int) int64 {
		return src.TargetResourceID
	}))

	res, err := h.assembleGraph(ctx.Context, req, graphRight, dstIds, func(_ int64) string {
		return "right"
	}, false)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{Data: res}, nil
}

func (h *Handler) ListResourceByIds(ctx *ginx.Context, req ListResourceByIdsReq) (ginx.Result, error) {
	resp, err := h.svc.ListResourceByIds(ctx.Context, nil, req.ResourceIds)
	if err != nil {
		return systemErrorResult, err
	}

	rs := lo.Map(resp, func(src domain.Resource, _ int) Resource {
		return Resource{
			ID:       src.ID,
			Name:     src.Name,
			ModelUID: src.ModelUID,
			Data:     src.Data,
		}
	})

	return ginx.Result{
		Data: RetrieveResources{
			Resources: rs,
		},
		Msg: "根据ID查询资源成功",
	}, nil
}

func (h *Handler) Search(ctx *ginx.Context, req SearchReq) (ginx.Result, error) {
	search, err := h.svc.Search(ctx.Context, req.Text)
	if err != nil {
		return systemErrorResult, err
	}

	// NOTE: 如果未检索到匹配数据，直接短路返回空列表，规避无意义的 MongoDB 聚合查询开销
	if len(search) == 0 {
		return ginx.Result{
			Data: []RetrieveSearchResources{},
		}, nil
	}

	modelUids := lo.Uniq(lo.Map(search, func(src domain.SearchResource, _ int) string {
		return src.ModelUid
	}))

	fields, err := h.attrSvc.SearchAttributeFieldsBySecure(ctx.Context, modelUids)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: lo.Map(search, func(src domain.SearchResource, _ int) RetrieveSearchResources {
			val, ok := fields[src.ModelUid]
			if ok {
				for _, name := range src.Data {
					for key := range name {
						if lo.Contains(val, key) {
							name[key] = ""
						}
					}
				}
			}
			return RetrieveSearchResources{
				ModelUid: src.ModelUid,
				Total:    src.Total,
				Data:     src.Data,
			}
		}),
	}, err
}

func (h *Handler) DeleteResource(ctx *ginx.Context, req DeleteResourceReq) (ginx.Result, error) {
	count, err := h.svc.DeleteResource(ctx.Context, req.Id)
	if err != nil {
		return systemErrorResult, err
	}

	// NOTE: 级联清理该资产作为源端或目标端所产生的所有关联连边，彻底杜绝拓扑图出现孤儿悬空节点
	if _, err = h.rrSvc.DeleteByResourceId(ctx.Context, req.Id); err != nil {
		return systemErrorResult, fmt.Errorf("级联删除资产关联关系失败: %w", err)
	}

	return ginx.Result{
		Data: count,
	}, nil
}

func (h *Handler) FindSecureData(ctx *ginx.Context, req FindSecureReq) (ginx.Result, error) {
	data, err := h.svc.FindSecureData(ctx.Context, req.ID, req.FieldUid)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: data,
	}, err
}

func (h *Handler) toCreateDomain(req CreateResourceReq) domain.Resource {
	return domain.Resource{
		Name:     req.Name,
		ModelUID: req.ModelUid,
		Data:     req.Data,
	}
}

func (h *Handler) toUpdateDomain(src UpdateResourceReq) domain.Resource {
	return domain.Resource{
		ID:       src.Id,
		Name:     src.Name,
		ModelUID: src.ModelUid,
		Data:     src.Data,
	}
}


