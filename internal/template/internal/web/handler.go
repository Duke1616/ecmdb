package web

import (
	"github.com/Duke1616/ecmdb/internal/template/internal/domain"
	"github.com/Duke1616/ecmdb/internal/template/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
)

type Handler struct {
	svc service.Service
}

func NewHandler(svc service.Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/api/template")
	g.POST("/create", ginx.WrapBody[CreateTemplateReq](h.CreateTemplate))
	g.POST("/detail", ginx.WrapBody[DetailTemplateReq](h.DetailTemplate))
}

func (h *Handler) CreateTemplate(ctx *gin.Context, req CreateTemplateReq) (ginx.Result, error) {
	d, err := h.toDomain(req)
	if err != nil {
		return systemErrorResult, err
	}

	t, err := h.svc.CreateTemplate(ctx, d)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: t,
	}, nil
}

func (h *Handler) DetailTemplate(ctx *gin.Context, req DetailTemplateReq) (ginx.Result, error) {
	return ginx.Result{}, nil
}

func (h *Handler) toDomain(req CreateTemplateReq) (domain.Template, error) {
	var rulesData []map[string]interface{}
	if err := json.Unmarshal([]byte(req.Rules), &rulesData); err != nil {
		return domain.Template{}, err
	}

	var optionsData map[string]interface{}
	if err := json.Unmarshal([]byte(req.Options), &optionsData); err != nil {
		return domain.Template{}, err
	}

	return domain.Template{
		Name:       "表单创建",
		CreateType: domain.SystemCreate,
		Rules:      rulesData,
		Options:    optionsData,
	}, nil
}
