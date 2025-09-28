package web

import (
	"github.com/Duke1616/ecmdb/internal/endpoint/internal/domain"
	"github.com/Duke1616/ecmdb/internal/endpoint/internal/service"
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

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/endpoint")
	g.POST("/register", ginx.WrapBody[RegisterEndpointReq](h.RegisterEndpoint))
	g.POST("/register/multi", ginx.WrapBody[RegisterEndpointsReq](h.BatchRegister))
	g.POST("/list", ginx.WrapBody[FilterPathReq](h.ListEndpoint))
}

func (h *Handler) RegisterEndpoint(ctx *gin.Context, req RegisterEndpointReq) (ginx.Result, error) {
	eId, err := h.svc.RegisterEndpoint(ctx, h.toDomain(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: eId,
	}, nil
}

func (h *Handler) BatchRegister(ctx *gin.Context, req RegisterEndpointsReq) (ginx.Result, error) {
	// 将请求中的端点转换为 domain 对象
	endpoints := make([]domain.Endpoint, len(req.RegisterEndpoint))
	for i, ep := range req.RegisterEndpoint {
		endpoints[i] = h.toDomain(ep)
	}
	
	// 直接调用 service 的按 Resource 注册方法
	count, err := h.svc.BatchRegisterByResource(ctx, req.Resource, endpoints)
	if err != nil {
		return systemErrorResult, err
	}
	
	return ginx.Result{
		Msg:  "批量注册端点成功",
		Data: count,
	}, nil
}

func (h *Handler) ListEndpoint(ctx *gin.Context, req FilterPathReq) (ginx.Result, error) {
	rts, total, err := h.svc.ListEndpoints(ctx, req.Offset, req.Limit, req.Path)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "查询API接口列表成功",
		Data: RetrieveEndpoints{
			Total: total,
			Endpoints: slice.Map(rts, func(idx int, src domain.Endpoint) Endpoint {
				return h.toEndpointVo(src)
			}),
		},
	}, nil
}

func (h *Handler) toDomain(req RegisterEndpointReq) domain.Endpoint {
	return domain.Endpoint{
		Path:         req.Path,
		Method:       req.Method,
		Resource:     req.Resource,
		Desc:         req.Desc,
		IsAuth:       req.IsAuth,
		IsAudit:      req.IsAudit,
		IsPermission: req.IsPermission,
	}
}

func (h *Handler) toEndpointVo(req domain.Endpoint) Endpoint {
	return Endpoint{
		Id:           req.Id,
		Path:         req.Path,
		Method:       req.Method,
		Resource:     req.Resource,
		Desc:         req.Desc,
		IsAuth:       req.IsAuth,
		IsAudit:      req.IsAudit,
		IsPermission: req.IsPermission,
	}
}
