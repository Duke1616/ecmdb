package web

import (
	"github.com/Duke1616/ecmdb/internal/template/internal/domain"
	"github.com/Duke1616/ecmdb/internal/template/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
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
	g.POST("/list", ginx.WrapBody[ListTemplateReq](h.ListTemplate))
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
	t, err := h.svc.DetailTemplate(ctx, req.Id)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: h.toTemplateVo(t),
	}, nil
}

func (h *Handler) ListTemplate(ctx *gin.Context, req ListTemplateReq) (ginx.Result, error) {
	rts, total, err := h.svc.ListTemplate(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "查询工单模版列表成功",
		Data: RetrieveTemplates{
			Total: total,
			Templates: slice.Map(rts, func(idx int, src domain.Template) Template {
				return h.toTemplateVo(src)
			}),
		},
	}, nil
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
		Name:       req.Name,
		CreateType: domain.SystemCreate,
		Rules:      rulesData,
		Options:    optionsData,
		Desc:       req.Desc,
	}, nil
}

func (h *Handler) toTemplateVo(req domain.Template) Template {
	return Template{
		Id:         req.Id,
		Name:       req.Name,
		Rules:      req.Rules,
		Options:    req.Options,
		CreateType: CreateType(req.CreateType),
		Desc:       req.Desc,
	}
}
