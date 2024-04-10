package web

import (
	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/internal/resource/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/gin-gonic/gin"
	"strconv"
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

	g.POST("/create/:model_id", ginx.WrapBody[CreateResourceReq](h.CreateResource))
}

func (h *Handler) CreateResource(ctx *gin.Context, req CreateResourceReq) (ginx.Result, error) {
	modelStr := ctx.Param("model_id")
	modelID, err := strconv.ParseInt(modelStr, 10, 64)
	if err != nil {
		return urlPathErrorResult, err
	}

	id, err := h.svc.CreateResource(ctx, domain.Resource{
		ModelID: modelID,
		Data:    req.Data,
	})

	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: id,
		Msg:  "创建资源成功",
	}, nil
}
