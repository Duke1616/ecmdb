package web

import (
	"fmt"

	"github.com/Duke1616/ecmdb/internal/strategy/internal/domain"
	"github.com/Duke1616/ecmdb/internal/strategy/internal/service"
	"github.com/Duke1616/ecmdb/internal/template"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"github.com/xen0n/go-workwx"
)

type Handler struct {
	svc         service.Service
	templateSvc template.Service
}

func NewHandler(svc service.Service, templateSvc template.Service) *Handler {
	return &Handler{
		svc:         svc,
		templateSvc: templateSvc,
	}
}

func (h *Handler) PublicRoutes(server *gin.Engine) {
	g := server.Group("/api/strategy")
	g.POST("/get_specified_template", ginx.WrapBody[GetSpecifiedTemplate](h.GetSpecifiedTemplate))
	g.POST("/create", ginx.WrapBody[GetSpecifiedTemplate](h.GetSpecifiedTemplate))
}

// GetSpecifiedTemplate 获取指定模版下可用的规则选项
func (h *Handler) GetSpecifiedTemplate(ctx *gin.Context, req GetSpecifiedTemplate) (ginx.Result, error) {
	t, err := h.templateSvc.DetailTemplate(ctx, req.Id)
	if err != nil {
		return systemErrorResult, err
	}

	var val []domain.Strategy
	for _, controls := range t.WechatOAControls.Controls {
		switch controls.Property.Control {
		case "Selector":
			key := controls.Property.Title[0].Text
			value := slice.Map(controls.Config.Selector.Options, func(idx int, src workwx.OATemplateControlConfigSelectorOption) string {
				return src.Value[0].Text
			})

			val = append(val, domain.Strategy{
				Key:   key,
				Value: value,
			})
		case "default":
			fmt.Println("不符合筛选规则")
		}
	}
	return ginx.Result{
		Data: val,
	}, nil
}

func (h *Handler) Validation(ctx *gin.Context, req GetSpecifiedTemplate) (ginx.Result, error) {
	return ginx.Result{}, nil
}
