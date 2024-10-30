package web

import (
	"encoding/json"
	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Duke1616/ecmdb/internal/engine/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
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
	g := server.Group("/api/engine")
	g.POST("/task/pass", ginx.WrapBody[Pass](h.Pass))
}

func (h *Handler) Pass(ctx *gin.Context, req Pass) (ginx.Result, error) {
	variables, err := json.Marshal(req.Variables)
	if err != nil {
		return systemErrorResult, err
	}

	err = engine.TaskPass(req.TaskId, req.Comment, string(variables), false)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{}, nil
}
