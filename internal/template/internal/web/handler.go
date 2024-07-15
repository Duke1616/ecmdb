package web

import (
	"github.com/Duke1616/ecmdb/internal/template/internal/domain"
	"github.com/Duke1616/ecmdb/internal/template/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/gotomicro/ego/core/elog"
	"sort"
)

type Handler struct {
	groupSvc service.GroupService
	svc      service.Service
	logger   *elog.Component
}

func NewHandler(svc service.Service, groupSvc service.GroupService) *Handler {
	return &Handler{
		svc:      svc,
		groupSvc: groupSvc,
		logger:   elog.DefaultLogger,
	}
}

func (h *Handler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/api/template")
	g.POST("/create", ginx.WrapBody[CreateTemplateReq](h.CreateTemplate))
	g.POST("/detail", ginx.WrapBody[DetailTemplateReq](h.DetailTemplate))
	g.POST("/list", ginx.WrapBody[ListTemplateReq](h.ListTemplate))
	g.POST("/delete", ginx.WrapBody[DeleteTemplateReq](h.DeleteTemplate))
	g.POST("/update", ginx.WrapBody[UpdateTemplateReq](h.UpdateTemplate))
	g.POST("/list/pipeline", ginx.Wrap(h.Pipeline))
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

func (h *Handler) DeleteTemplate(ctx *gin.Context, req DeleteTemplateReq) (ginx.Result, error) {
	count, err := h.svc.DeleteTemplate(ctx, req.Id)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: count,
	}, nil
}

func (h *Handler) UpdateTemplate(ctx *gin.Context, req UpdateTemplateReq) (ginx.Result, error) {
	d, err := h.toUpdateDomain(req)
	if err != nil {
		return systemErrorResult, err
	}

	t, err := h.svc.UpdateTemplate(ctx, d)

	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: t,
	}, nil
}

func (h *Handler) Pipeline(ctx *gin.Context) (ginx.Result, error) {
	// 根据 组ID 聚合查询所有数据
	pipeline, err := h.svc.Pipeline(ctx)
	if err != nil {
		return systemErrorResult, err
	}

	// 获取组信息
	ids := slice.Map(pipeline, func(idx int, src domain.TemplateCombination) int64 {
		return src.Id
	})
	gs, err := h.groupSvc.ListByIds(ctx, ids)
	if err != nil {
		return systemErrorResult, err
	}
	gsMap := slice.ToMap(gs, func(element domain.TemplateGroup) int64 {
		return element.Id
	})

	// 组合前端返回数据
	tc := slice.Map(pipeline, func(idx int, src domain.TemplateCombination) TemplateCombination {
		val, _ := gsMap[src.Id]
		return TemplateCombination{
			Id:    src.Id,
			Name:  val.Name,
			Icon:  val.Icon,
			Total: src.Id,
			Templates: slice.Map(src.Templates, func(idx int, src domain.Template) Template {
				return h.toTemplateVo(src)
			}),
		}
	})

	sort.Slice(tc, func(i, j int) bool {
		// 根据需要的排序逻辑进行排序，这里假设你有一个字段可以用来排序，比如 id
		return tc[i].Total < tc[j].Total
	})

	return ginx.Result{
		Data: RetrieveTemplateCombination{TemplateCombinations: tc},
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
		WorkflowId: req.WorkflowId,
		GroupId:    req.GroupId,
		Icon:       req.Icon,
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
		WorkflowId: req.WorkflowId,
		GroupId:    req.GroupId,
		Icon:       req.Icon,
		Rules:      req.Rules,
		Options:    req.Options,
		CreateType: CreateType(req.CreateType),
		Desc:       req.Desc,
	}
}

func (h *Handler) toUpdateDomain(req UpdateTemplateReq) (domain.Template, error) {
	var rulesData []map[string]interface{}
	if err := json.Unmarshal([]byte(req.Rules), &rulesData); err != nil {
		return domain.Template{}, err
	}

	var optionsData map[string]interface{}
	if err := json.Unmarshal([]byte(req.Options), &optionsData); err != nil {
		return domain.Template{}, err
	}

	return domain.Template{
		Id:         req.Id,
		Name:       req.Name,
		Icon:       req.Icon,
		GroupId:    req.GroupId,
		WorkflowId: req.WorkflowId,
		Rules:      rulesData,
		Options:    optionsData,
	}, nil

}
