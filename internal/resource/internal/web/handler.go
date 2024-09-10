package web

import (
	"fmt"
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/internal/resource/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

type Handler struct {
	svc     service.Service
	attrSvc attribute.Service
	RRSvc   relation.RRSvc
}

func NewHandler(service service.Service, attributeSvc attribute.Service, RRSvc relation.RRSvc) *Handler {
	return &Handler{
		svc:     service,
		attrSvc: attributeSvc,
		RRSvc:   RRSvc,
	}
}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/resource")
	// 资源操作
	g.POST("/create", ginx.WrapBody[CreateResourceReq](h.CreateResource))
	// 根据 ID 查询资源列表
	g.POST("/detail", ginx.WrapBody[DetailResourceReq](h.DetailResource))
	// 根据模型 UID 查询资源列表
	g.POST("/list", ginx.WrapBody[ListResourceReq](h.ListResource))
	g.POST("/delete", ginx.WrapBody[DeleteResourceReq](h.DeleteResource))
	// 资源关联关系
	g.POST("/relation/can_be_related", ginx.WrapBody[ListCanBeRelatedReqByModel](h.ListCanBeFilterRelated))
	g.POST("/relation/diagram", ginx.WrapBody[ListDiagramReq](h.FindDiagram))
	g.POST("/relation/graph", ginx.WrapBody[ListDiagramReq](h.FindAllGraph))
	g.POST("/relation/graph/add/left", ginx.WrapBody[ListDiagramReq](h.FindLeftGraph))
	g.POST("/relation/graph/add/right", ginx.WrapBody[ListDiagramReq](h.FindRightGraph))

	// 根据模型 UID 查询资源列表
	g.POST("/list/ids", ginx.WrapBody[ListResourceByIdsReq](h.ListResourceByIds))

	// 全文检索
	g.POST("/search", ginx.WrapBody[SearchReq](h.Search))

	// 查询加密字段信息
	g.POST("/secure", ginx.WrapBody[FindSecureReq](h.FindSecureData))

	g.POST("/update", ginx.WrapBody[UpdateResourceReq](h.UpdateResource))
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

func (h *Handler) FindAllGraph(ctx *gin.Context, req ListDiagramReq) (ginx.Result, error) {
	// 查询资产关联上下级拓扑
	graph, _, err := h.RRSvc.ListDiagram(ctx, req.ModelUid, req.ResourceId)
	if err != nil {
		return systemErrorResult, err
	}
	var (
		srcId []int64
		dstId []int64
	)

	rrs := append(graph.SRC, graph.DST...)
	lines := slice.Map(rrs, func(idx int, src relation.ResourceRelation) Line {
		return Line{
			From: strconv.FormatInt(src.SourceResourceID, 10),
			To:   strconv.FormatInt(src.TargetResourceID, 10),
		}
	})

	// 查询关联的所有节点 ids
	srcId = slice.Map(graph.SRC, func(idx int, src relation.ResourceRelation) int64 {
		return src.TargetResourceID
	})
	dstId = slice.Map(graph.DST, func(idx int, src relation.ResourceRelation) int64 {
		return src.SourceResourceID
	})

	ids := append(srcId, dstId...)

	// 查询节点信息
	rs, err := h.svc.ListResourceByIds(ctx, nil, ids)
	if err != nil {
		return systemErrorResult, err
	}

	nodes := slice.Map(rs, func(idx int, src domain.Resource) Node {
		data := make(map[string]any, 1)
		data["model_uid"] = src.ModelUID
		data["isNeedLoadDataFromRemoteServer"] = true
		data["childrenLoaded"] = false
		for _, id := range srcId {
			if src.ID == id {
				return Node{
					ID:                   strconv.FormatInt(src.ID, 10),
					Text:                 src.Name,
					Data:                 data,
					ExpandHolderPosition: "right",
					Expanded:             false,
				}
			}
		}
		return Node{
			ID:                   strconv.FormatInt(src.ID, 10),
			Text:                 src.Name,
			ExpandHolderPosition: "left",
			Expanded:             false,
			Data:                 data,
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

	lines := slice.Map(graphLeft, func(idx int, src relation.ResourceRelation) Line {
		return Line{
			From: strconv.FormatInt(src.SourceResourceID, 10),
			To:   strconv.FormatInt(src.TargetResourceID, 10),
		}
	})

	// 查询关联的所有节点 ids
	srcIds = slice.Map(graphLeft, func(idx int, src relation.ResourceRelation) int64 {
		return src.SourceResourceID
	})

	// 查询节点信息
	rs, err := h.svc.ListResourceByIds(ctx, nil, srcIds)
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

	lines := slice.Map(graphRight, func(idx int, src relation.ResourceRelation) Line {
		return Line{
			From: strconv.FormatInt(src.SourceResourceID, 10),
			To:   strconv.FormatInt(src.TargetResourceID, 10),
		}
	})

	// 查询关联的所有节点 ids
	srcIds = slice.Map(graphRight, func(idx int, src relation.ResourceRelation) int64 {
		return src.TargetResourceID
	})

	// 查询节点信息
	rs, err := h.svc.ListResourceByIds(ctx, nil, srcIds)
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
		},
	}, nil
}

func (h *Handler) FindDiagram(ctx *gin.Context, req ListDiagramReq) (ginx.Result, error) {
	// 查询资产关联上下级拓扑
	diagram, _, err := h.RRSvc.ListDiagram(ctx, req.ModelUid, req.ResourceId)
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
	src = slice.Map(diagram.SRC, func(idx int, src relation.ResourceRelation) ResourceRelation {
		return h.toResourceRelationVo(src)
	})
	dst = slice.Map(diagram.DST, func(idx int, src relation.ResourceRelation) ResourceRelation {
		return h.toResourceRelationVo(src)
	})

	// 查询关联的所有节点 ids
	srcId = slice.Map(diagram.SRC, func(idx int, src relation.ResourceRelation) int64 {
		return src.TargetResourceID
	})
	dstId = slice.Map(diagram.DST, func(idx int, src relation.ResourceRelation) int64 {
		return src.SourceResourceID
	})
	ids := append(srcId, dstId...)

	// 查询节点信息
	rs, err := h.svc.ListResourceByIds(ctx, nil, ids)
	if err != nil {
		return systemErrorResult, err
	}

	// 组合前端返回数据
	assets := make(map[string][]ResourceAssets, len(diagram.DST)+len(diagram.SRC))
	assets = slice.ToMapV(rs, func(element domain.Resource) (string, []ResourceAssets) {
		return element.ModelUID, slice.FilterMap(rs, func(idx int, src domain.Resource) (ResourceAssets, bool) {
			if src.ModelUID == element.ModelUID {
				return ResourceAssets{
					ResourceID:   src.ID,
					ResourceName: src.Name,
				}, true
			}
			return ResourceAssets{}, false
		})
	})

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

func (h *Handler) toResourceRelationVo(src relation.ResourceRelation) ResourceRelation {
	return ResourceRelation{
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
		ID:   src.Id,
		Name: src.Name,
		Data: src.Data,
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
