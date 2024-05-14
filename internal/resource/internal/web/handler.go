package web

import (
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/internal/resource/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
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

func (h *Handler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/resource")
	// 资源操作
	g.POST("/create", ginx.WrapBody[CreateResourceReq](h.CreateResource))
	// 根据 ID 查询资源列表
	g.POST("/detail", ginx.WrapBody[DetailResourceReq](h.DetailResource))
	// 根据模型 UID 查询资源列表
	g.POST("/list", ginx.WrapBody[ListResourceReq](h.ListResource))

	// 资源关联关系
	g.POST("/relation/can_be_related", ginx.WrapBody[ListCanBeRelatedReq](h.ListCanBeRelated))
	g.POST("/relation/diagram", ginx.WrapBody[ListDiagramReq](h.FindDiagram))
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

func (h *Handler) ListCanBeRelated(ctx *gin.Context, req ListCanBeRelatedReq) (ginx.Result, error) {
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
	rn := strings.Split(req.RelationName, "_")
	if rn[0] == req.ModelUid {
		mUid = rn[2]
		excludeIds, err = h.RRSvc.ListSrcRelated(ctx, req.ModelUid, req.RelationName, req.ResourceId)
	} else {
		mUid = rn[0]
		excludeIds, err = h.RRSvc.ListDstRelated(ctx, rn[2], req.RelationName, req.ResourceId)
	}

	// 查看模型字段
	// model_uid = mongo
	fields, err := h.attrSvc.SearchAttributeFieldsByModelUid(ctx, mUid)
	if err != nil {
		return systemErrorResult, err
	}

	// 排除已关联数据, 返回未关联数据
	rrs, err := h.svc.ListExcludeResourceByIds(ctx, fields, mUid, req.Offset, req.Limit, excludeIds)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: rrs,
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
