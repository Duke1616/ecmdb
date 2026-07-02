package web

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Duke1616/ecmdb/internal/domain"
	attributeservice "github.com/Duke1616/ecmdb/internal/service/attribute"
	modelservice "github.com/Duke1616/ecmdb/internal/service/model"
	relationservice "github.com/Duke1616/ecmdb/internal/service/relation"
	service "github.com/Duke1616/ecmdb/internal/service/resource"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/Duke1616/eiam/pkg/web/capability"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

type Handler struct {
	svc      service.EncryptedSvc
	attrSvc  attributeservice.Service
	modelSvc modelservice.Service
	RRSvc    relationservice.RelationResourceService
	capability.IRegistry
}

func NewHandler(svc service.EncryptedSvc, attributeSvc attributeservice.Service, modelSvc modelservice.Service, RRSvc relationservice.RelationResourceService) *Handler {
	return &Handler{
		svc:       svc,
		attrSvc:   attributeSvc,
		modelSvc:  modelSvc,
		RRSvc:     RRSvc,
		IRegistry: capability.NewRegistry("cmdb", "resource", "资产仓库"),
	}
}

// PrivateRoutes 注册资源管理模块需要中心化登录及权限判定（由 EIAM SDK 统一拦截承载）的私有路由
func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/resource")

	// ==========================================
	// 1. 资产基础操作接口
	// ==========================================

	// 创建资产
	g.POST("/create", h.Capability("创建资产", "add").
		Handle(ginx.WrapBody[CreateResourceReq](h.CreateResource)),
	)

	// 查询资产详情
	g.POST("/detail", h.Capability("资产详情", "get").
		Handle(ginx.WrapBody[DetailResourceReq](h.DetailResource)),
	)

	// 根据模型 UID 查询资产列表
	g.POST("/list", h.Capability("资产列表", "view").
		Needs("cmdb:model:view", "cmdb:attribute:view", "cmdb:plugin:actions").
		Handle(ginx.WrapBody[ListResourceReq](h.ListResource)),
	)

	// 删除资产
	g.POST("/delete", h.Capability("删除资产", "delete").
		Handle(ginx.WrapBody[DeleteResourceReq](h.DeleteResource)),
	)

	// 修改资产信息
	g.POST("/update", h.Capability("修改资产", "edit").
		Handle(ginx.WrapBody[UpdateResourceReq](h.UpdateResource)),
	)

	// 设置自定义属性
	g.POST("/set_custom_field", h.Capability("设置自定义属性", "edit_custom_field").
		Handle(ginx.WrapBody[SetCustomFieldReq](h.SetCustomField)),
	)

	// ==========================================
	// 2. 资产关联拓扑接口
	// ==========================================

	// 查询可关联的资产列表
	g.POST("/relation/can_be_related", h.Capability("查询可关联的资产列表", "view_can_be_related").
		NoSync().
		Handle(ginx.WrapBody[ListCanBeRelatedReqByModel](h.ListCanBeFilterRelated)),
	)

	// 查询资产拓扑关系图谱
	//g.POST("/relation/diagram", h.Capability("查询资产关系图谱", "relation_view_diagram").
	//	Handle(ginx.WrapBody[ListDiagramReq](h.FindDiagram)),
	//)

	// 查询资产关联拓扑图
	g.POST("/relation/graph", h.Capability("资产关联拓扑图", "view_relation_graph").
		Group("资产仓库/关联关系").
		Needs("cmdb:resource:add_relation_left", "cmdb:resource:add_relation_right").
		Handle(ginx.WrapBody[ListDiagramReq](h.FindAllGraph)),
	)

	// 拓扑图向左拓展
	g.POST("/relation/graph/add/left", h.Capability("拓扑图向左拓展", "add_relation_left").
		Group("资产仓库/关联关系").
		NoSync().
		Handle(ginx.WrapBody[ListDiagramReq](h.FindLeftGraph)),
	)

	// 拓扑图向右拓展
	g.POST("/relation/graph/add/right", h.Capability("拓扑图向右拓展", "add_relation_right").
		Group("资产仓库/关联关系").
		NoSync().
		Handle(ginx.WrapBody[ListDiagramReq](h.FindRightGraph)),
	)

	// ==========================================
	// 3. 资产检索与安全字段接口
	// ==========================================

	// 批量查询资产
	g.POST("/list/ids", h.Capability("批量查询资产", "view_by_ids").
		NoSync().
		Handle(ginx.WrapBody[ListResourceByIdsReq](h.ListResourceByIds)),
	)

	// 全文检索资产
	g.POST("/search", h.Capability("全文检索资产", "search").
		Group("资产仓库/全局搜索").
		Handle(ginx.WrapBody[SearchReq](h.Search)),
	)

	// 查询加密字段数据
	g.POST("/secure", h.Capability("查询加密字段", "get_secure").
		Handle(ginx.WrapBody[FindSecureReq](h.FindSecureData)),
	)

	// ==========================================
	// 4. 资源关联关系管理接口
	// ==========================================

	// 创建资源关联关系
	g.POST("/relation/create", h.Capability("创建资产关系", "relation_add").
		Group("资产仓库/关联关系").
		Needs("cmdb:resource:view_can_be_related").
		Handle(ginx.WrapBody[CreateResourceRelationReq](h.CreateResourceRelation)),
	)

	// 所有资产关系聚合查询
	g.POST("/relation/pipeline/all", h.Capability("所有资产关系聚合查询", "relation_pipeline_all").
		Group("资产仓库/关联关系").
		Needs("cmdb:relation:view", "cmdb:model-relation:view", "cmdb:model:view_by_uids",
			"cmdb:attribute:view_fields", "cmdb:resource:view_by_ids").
		Handle(ginx.WrapBody[ListResourceDiagramReq](h.ListAllAggregated)),
	)

	// 删除资产关系
	g.POST("/relation/delete", h.Capability("删除资产关系", "relation_delete").
		Group("资产仓库/关联关系").
		Handle(ginx.WrapBody[DeleteResourceRelationReq](h.DeleteResourceRelation)),
	)
}

func (h *Handler) CreateResource(ctx *gin.Context, req CreateResourceReq) (ginx.Result, error) {
	id, err := h.svc.CreateResource(ctx, h.toDomain(req))

	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: id,
		Msg:  "创建资源成功",
	}, nil
}

func (h *Handler) DetailResource(ctx *gin.Context, req DetailResourceReq) (ginx.Result, error) {
	fields, err := h.attrSvc.SearchAttributeFieldsByModelUid(ctx, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}

	resp, err := h.svc.FindResourceById(ctx, fields, req.ID)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: resp,
		Msg:  "查看资源详情成功",
	}, nil
}

func (h *Handler) SetCustomField(ctx *gin.Context, req SetCustomFieldReq) (ginx.Result, error) {
	count, err := h.svc.SetCustomField(ctx, req.Id, req.Field, req.Data)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: count,
	}, nil
}

func (h *Handler) ListResource(ctx *gin.Context, req ListResourceReq) (ginx.Result, error) {
	fields, err := h.attrSvc.SearchAttributeFieldsByModelUid(ctx, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}

	resp, total, err := h.svc.ListResource(ctx, fields, req.ModelUid, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	rs := slice.Map(resp, func(idx int, src domain.Resource) Resource {
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

func (h *Handler) UpdateResource(ctx *gin.Context, req UpdateResourceReq) (ginx.Result, error) {
	resource := h.toDomainUpdate(req)
	t, err := h.svc.UpdateResource(ctx, resource)

	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: t,
	}, nil
}

func (h *Handler) ListCanBeFilterRelated(ctx *gin.Context, req ListCanBeRelatedReqByModel) (ginx.Result, error) {
	var (
		mUid       string
		err        error
		excludeIds []int64
	)
	/*
		查询已经关联的数据
		model_uid = physical
		relation_name = "physical_run_mongo"
	*/
	if req.RelationName == "" {
		return systemErrorResult, fmt.Errorf("关联名称为空")
	}

	// 传递的是当前模型UID （特别注意）
	rn := strings.Split(req.RelationName, "_")
	if rn[0] == req.ModelUid {
		mUid = rn[2]
		excludeIds, err = h.RRSvc.ListSrcRelated(ctx, req.ModelUid, req.RelationName, req.ResourceId)
	} else {
		mUid = rn[0]
		excludeIds, err = h.RRSvc.ListDstRelated(ctx, rn[2], req.RelationName, req.ResourceId)
	}
	if err != nil {
		return systemErrorResult, err
	}

	fields, err := h.attrSvc.SearchAttributeFieldsByModelUid(ctx, mUid)

	if err != nil {
		return systemErrorResult, err
	}

	// 排除已关联数据, 并且进行过滤，返回未关联数据
	rrs, total, err := h.svc.ListExcludeAndFilterResourceByIds(ctx, fields, mUid, req.Offset, req.Limit, excludeIds,
		domain.Condition{
			Name:      req.FilterName,
			Condition: req.FilterCondition,
			Input:     req.FilterInput,
		})
	if err != nil {
		return systemErrorResult, err
	}

	rs := slice.Map(rrs, func(idx int, src domain.Resource) Resource {
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

func (h *Handler) graphModels(ctx *gin.Context, resources []domain.Resource, rootModelUID string) ([]GraphModel, error) {
	modelUIDs := lo.Uniq(append(lo.Map(resources, func(src domain.Resource, _ int) string {
		return src.ModelUID
	}), rootModelUID))
	modelUIDs = lo.Filter(modelUIDs, func(uid string, _ int) bool {
		return uid != ""
	})
	if len(modelUIDs) == 0 {
		return []GraphModel{}, nil
	}

	models, err := h.modelSvc.GetByUids(ctx.Request.Context(), modelUIDs)
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

func (h *Handler) FindAllGraph(ctx *gin.Context, req ListDiagramReq) (ginx.Result, error) {
	// 查询资产关联上下级拓扑（支持多级递归，默认递归3层）
	maxDepth := req.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 3
	}
	graph, err := h.RRSvc.ListRecursiveDiagram(ctx, req.ModelUid, req.ResourceId, maxDepth)
	if err != nil {
		return systemErrorResult, err
	}
	var (
		srcId []int64
		dstId []int64
	)

	rrs := append(graph.SRC, graph.DST...)
	lines := slice.Map(rrs, func(idx int, src domain.ResourceRelation) Line {
		return Line{
			From: strconv.FormatInt(src.SourceResourceID, 10),
			To:   strconv.FormatInt(src.TargetResourceID, 10),
		}
	})

	// 查询关联的所有节点 ids
	srcId = slice.Map(graph.SRC, func(idx int, src domain.ResourceRelation) int64 {
		return src.TargetResourceID
	})
	dstId = slice.Map(graph.DST, func(idx int, src domain.ResourceRelation) int64 {
		return src.SourceResourceID
	})

	ids := append(srcId, dstId...)

	// 查询节点信息
	rs, err := h.svc.ListResourceByIds(ctx, []string{"name"}, ids)
	if err != nil {
		return systemErrorResult, err
	}
	models, err := h.graphModels(ctx, rs, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}

	srcIdMap := make(map[int64]struct{}, len(srcId))
	for _, id := range srcId {
		srcIdMap[id] = struct{}{}
	}

	nodes := slice.Map(rs, func(idx int, src domain.Resource) Node {
		data := make(map[string]any, 1)
		data["model_uid"] = src.ModelUID
		data["isNeedLoadDataFromRemoteServer"] = true
		data["childrenLoaded"] = false

		expandHolderPosition := "left"
		if _, ok := srcIdMap[src.ID]; ok {
			expandHolderPosition = "right"
		}

		return Node{
			ID:                   strconv.FormatInt(src.ID, 10),
			Text:                 src.Name,
			Data:                 data,
			ExpandHolderPosition: expandHolderPosition,
			Expanded:             false,
		}
	})

	nodes = append(nodes, Node{
		ID:       strconv.FormatInt(req.ResourceId, 10),
		Text:     req.ResourceName,
		Expanded: true,
		Data: map[string]any{
			"model_uid": req.ModelUid,
		},
	})

	return ginx.Result{
		Data: RetrieveGraph{
			Lines:  lines,
			Nodes:  nodes,
			RootId: strconv.FormatInt(req.ResourceId, 10),
			Models: models,
		},
	}, nil
}

func (h *Handler) FindLeftGraph(ctx *gin.Context, req ListDiagramReq) (ginx.Result, error) {
	// 查询资产关联上下级拓扑
	graphLeft, _, err := h.RRSvc.ListDstResources(ctx, req.ModelUid, req.ResourceId)
	if err != nil {
		return systemErrorResult, err
	}
	var (
		srcIds []int64
	)

	lines := slice.Map(graphLeft, func(idx int, src domain.ResourceRelation) Line {
		return Line{
			From: strconv.FormatInt(src.SourceResourceID, 10),
			To:   strconv.FormatInt(src.TargetResourceID, 10),
		}
	})

	// 查询关联的所有节点 ids
	srcIds = slice.Map(graphLeft, func(idx int, src domain.ResourceRelation) int64 {
		return src.SourceResourceID
	})

	// 查询节点信息
	rs, err := h.svc.ListResourceByIds(ctx, []string{"name"}, srcIds)
	if err != nil {
		return systemErrorResult, err
	}
	models, err := h.graphModels(ctx, rs, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}

	nodes := slice.Map(rs, func(idx int, src domain.Resource) Node {
		data := make(map[string]any, 1)
		data["model_uid"] = src.ModelUID
		data["isNeedLoadDataFromRemoteServer"] = true
		data["childrenLoaded"] = false
		return Node{
			ID:                   strconv.FormatInt(src.ID, 10),
			Text:                 src.Name,
			ExpandHolderPosition: "left",
			Expanded:             false,
			Data:                 data,
		}
	})

	return ginx.Result{
		Data: RetrieveGraph{
			Lines:  lines,
			Nodes:  nodes,
			RootId: strconv.FormatInt(req.ResourceId, 10),
			Models: models,
		},
	}, nil
}

func (h *Handler) FindRightGraph(ctx *gin.Context, req ListDiagramReq) (ginx.Result, error) {
	// 查询资产关联上下级拓扑
	graphRight, _, err := h.RRSvc.ListSrcResources(ctx, req.ModelUid, req.ResourceId)
	if err != nil {
		return systemErrorResult, err
	}
	var (
		srcIds []int64
	)

	lines := slice.Map(graphRight, func(idx int, src domain.ResourceRelation) Line {
		return Line{
			From: strconv.FormatInt(src.SourceResourceID, 10),
			To:   strconv.FormatInt(src.TargetResourceID, 10),
		}
	})

	// 查询关联的所有节点 ids
	srcIds = slice.Map(graphRight, func(idx int, src domain.ResourceRelation) int64 {
		return src.TargetResourceID
	})

	// 查询节点信息
	rs, err := h.svc.ListResourceByIds(ctx, []string{"name"}, srcIds)
	if err != nil {
		return systemErrorResult, err
	}
	models, err := h.graphModels(ctx, rs, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}

	nodes := slice.Map(rs, func(idx int, src domain.Resource) Node {
		data := make(map[string]any, 1)
		data["model_uid"] = src.ModelUID
		data["isNeedLoadDataFromRemoteServer"] = true
		data["childrenLoaded"] = false
		return Node{
			ID:                   strconv.FormatInt(src.ID, 10),
			Text:                 src.Name,
			ExpandHolderPosition: "right",
			Expanded:             false,
			Data:                 data,
		}
	})

	return ginx.Result{
		Data: RetrieveGraph{
			Lines:  lines,
			Nodes:  nodes,
			RootId: strconv.FormatInt(req.ResourceId, 10),
			Models: models,
		},
	}, nil
}

func (h *Handler) FindDiagram(ctx *gin.Context, req ListDiagramReq) (ginx.Result, error) {
	// 查询资产关联上下级拓扑（支持多级递归，默认递归3层）
	maxDepth := req.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 3
	}
	diagram, err := h.RRSvc.ListRecursiveDiagram(ctx, req.ModelUid, req.ResourceId, maxDepth)
	if err != nil {
		return systemErrorResult, err
	}
	var (
		src   []ResourceRelation
		dst   []ResourceRelation
		srcId []int64
		dstId []int64
	)

	// 组合前端展示数据
	src = slice.Map(diagram.SRC, func(idx int, src domain.ResourceRelation) ResourceRelation {
		return h.toResourceRelationVo(src)
	})
	dst = slice.Map(diagram.DST, func(idx int, src domain.ResourceRelation) ResourceRelation {
		return h.toResourceRelationVo(src)
	})

	// 查询关联的所有节点 ids
	srcId = slice.Map(diagram.SRC, func(idx int, src domain.ResourceRelation) int64 {
		return src.TargetResourceID
	})
	dstId = slice.Map(diagram.DST, func(idx int, src domain.ResourceRelation) int64 {
		return src.SourceResourceID
	})
	ids := append(srcId, dstId...)

	// 查询节点信息
	rs, err := h.svc.ListResourceByIds(ctx, []string{"name"}, ids)
	if err != nil {
		return systemErrorResult, err
	}

	// 组合前端返回数据
	assets := make(map[string][]ResourceAssets)
	for _, src := range rs {
		assets[src.ModelUID] = append(assets[src.ModelUID], ResourceAssets{
			ResourceID:   src.ID,
			ResourceName: src.Name,
		})
	}

	return ginx.Result{
		Data: RetrieveDiagram{
			SRC:    src,
			DST:    dst,
			Assets: assets,
		},
	}, nil
}

func (h *Handler) ListResourceByIds(ctx *gin.Context, req ListResourceByIdsReq) (ginx.Result, error) {
	fields, err := h.attrSvc.SearchAttributeFieldsByModelUid(ctx, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}

	resp, err := h.svc.ListResourceByIds(ctx, fields, req.ResourceIds)
	if err != nil {
		return systemErrorResult, err
	}

	rs := slice.Map(resp, func(idx int, src domain.Resource) Resource {
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

func (h *Handler) Search(ctx *gin.Context, req SearchReq) (ginx.Result, error) {
	search, err := h.svc.Search(ctx, req.Text)
	if err != nil {
		return systemErrorResult, err
	}

	modelUids := slice.Map(search, func(idx int, src domain.SearchResource) string {
		return src.ModelUid
	})

	fields, err := h.attrSvc.SearchAttributeFieldsBySecure(ctx, modelUids)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: slice.Map(search, func(idx int, src domain.SearchResource) RetrieveSearchResources {
			val, ok := fields[src.ModelUid]
			if ok {
				for _, name := range src.Data {
					for key := range name {
						if contains(val, key) {
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

func (h *Handler) DeleteResource(ctx *gin.Context, req DeleteResourceReq) (ginx.Result, error) {
	count, err := h.svc.DeleteResource(ctx, req.Id)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: count,
	}, nil
}

func (h *Handler) FindSecureData(ctx *gin.Context, req FindSecureReq) (ginx.Result, error) {
	data, err := h.svc.FindSecureData(ctx, req.ID, req.FieldUid)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: data,
	}, err
}

func (h *Handler) toDomain(req CreateResourceReq) domain.Resource {
	return domain.Resource{
		Name:     req.Name,
		ModelUID: req.ModelUid,
		Data:     req.Data,
	}
}

func (h *Handler) toResourceRelationVo(src domain.ResourceRelation) ResourceRelation {
	return ResourceRelation{
		ID:               src.ID,
		SourceModelUID:   src.SourceModelUID,
		TargetModelUID:   src.TargetModelUID,
		SourceResourceID: src.SourceResourceID,
		TargetResourceID: src.TargetResourceID,
		RelationTypeUID:  src.RelationTypeUID,
		RelationName:     src.RelationName,
	}
}

func (h *Handler) toDomainUpdate(src UpdateResourceReq) domain.Resource {
	return domain.Resource{
		ID:       src.Id,
		Name:     src.Name,
		ModelUID: src.ModelUid,
		Data:     src.Data,
	}
}

func contains(slice []string, elem string) bool {
	for _, e := range slice {
		if e == elem {
			return true
		}
	}
	return false
}
