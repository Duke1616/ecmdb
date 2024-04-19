package web

import (
	"fmt"
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/internal/resource/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/gin-gonic/gin"
	"strings"
)

type Handler struct {
	svc          service.Service
	attributeSvc attribute.Service
	RRSvc        relation.RRSvc
}

func NewHandler(service service.Service, attributeSvc attribute.Service, RRSvc relation.RRSvc) *Handler {
	return &Handler{
		svc:          service,
		attributeSvc: attributeSvc,
		RRSvc:        RRSvc,
	}
}

func (h *Handler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/resource")
	g.POST("/create", ginx.WrapBody[CreateResourceReq](h.CreateResource))

	// 查看资产详情信息
	g.POST("/detail", ginx.WrapBody[DetailResourceReq](h.DetailResource))

	// 根据模型查看资产列表
	g.POST("/list", ginx.WrapBody[ListResourceReq](h.ListResource))

	// 根据 ids 查询模型资产列表
	g.POST("/list/ids", ginx.WrapBody[ListResourceIdsReq](h.ListResourceByIds))

	// 查询可以关联的节点
	g.POST("/list/related", ginx.WrapBody[ListRelatedReq](h.ListRelated))
}

func (h *Handler) CreateResource(ctx *gin.Context, req CreateResourceReq) (ginx.Result, error) {
	id, err := h.svc.CreateResource(ctx, domain.Resource{
		Name:     req.Name,
		ModelUID: req.ModelUid,
		Data:     req.Data,
	})

	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: id,
		Msg:  "创建资源成功",
	}, nil
}

func (h *Handler) DetailResource(ctx *gin.Context, req DetailResourceReq) (ginx.Result, error) {
	projection, err := h.attributeSvc.SearchAttributeFiled(ctx, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}

	resp, err := h.svc.FindResourceById(ctx, projection, req.ID)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: resp,
		Msg:  "查看资源详情成功",
	}, nil
}

func (h *Handler) ListResource(ctx *gin.Context, req ListResourceReq) (ginx.Result, error) {
	projection, err := h.attributeSvc.SearchAttributeFiled(ctx, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}

	fmt.Print(projection)
	resp, err := h.svc.ListResource(ctx, projection, req.ModelUid, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: resp,
		Msg:  "查看资源列表成功",
	}, nil
}

func (h *Handler) ListResourceByIds(ctx *gin.Context, req ListResourceIdsReq) (ginx.Result, error) {
	filed, err := h.attributeSvc.SearchAttributeFiled(ctx, req.ModelUid)
	if err != nil {
		return ginx.Result{}, err
	}

	rrs, err := h.svc.ListResourceByIds(ctx, filed, req.Ids)
	if err != nil {
		return ginx.Result{}, err
	}

	return ginx.Result{Data: rrs}, nil
}

func (h *Handler) ListRelated(ctx *gin.Context, req ListRelatedReq) (ginx.Result, error) {
	var (
		mUid       string
		err        error
		excludeIds []int64
	)
	// 查询已经关联的数据
	// "host_run_mysql"
	rn := strings.Split(req.RelationName, "_")
	if rn[0] == req.ModelUid {
		mUid = rn[2]
		excludeIds, err = h.RRSvc.ListDstRelated(ctx, rn[2], req.RelationName, req.ResourceId)
	} else {
		mUid = rn[0]
		excludeIds, err = h.RRSvc.ListSrcRelated(ctx, rn[0], req.RelationName, req.ResourceId)
	}

	// 查看模型字段
	filed, err := h.attributeSvc.SearchAttributeFiled(ctx, mUid)
	if err != nil {
		return systemErrorResult, err
	}

	// 排除已关联数据, 返回未关联数据
	rrs, err := h.svc.ListExcludeResource(ctx, filed, mUid, req.Offset, req.Limit, excludeIds)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: rrs,
	}, nil
}
