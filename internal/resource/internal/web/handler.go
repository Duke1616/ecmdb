package web

import (
	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/internal/resource/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc service.Service
}

func NewHandler(service service.Service) *Handler {
	return &Handler{
		svc: service,
	}
}

func (h *Handler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/resource")

	g.POST("/create/:model_identifies", ginx.WrapBody[CreateResourceReq](h.CreateResource))
}

func (h *Handler) CreateResource(ctx *gin.Context, req CreateResourceReq) (ginx.Result, error) {
	modelIdentifies := ctx.Param("model_identifies")

	id, err := h.svc.CreateResource(ctx, domain.Resource{
		ModelIdentifies: modelIdentifies,
		Data:            req.Data,
	})

	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: id,
		Msg:  "创建资源成功",
	}, nil
}
