package web

import (
	"fmt"
	"github.com/Duke1616/ecmdb/internal/pkg/rule"
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

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/template")
	g.POST("/create", ginx.WrapBody[CreateTemplateReq](h.CreateTemplate))
	g.POST("/detail", ginx.WrapBody[DetailTemplateReq](h.DetailTemplate))
	g.POST("/list", ginx.WrapBody[ListTemplateReq](h.ListTemplate))
	g.POST("/delete", ginx.WrapBody[DeleteTemplateReq](h.DeleteTemplate))
	g.POST("/update", ginx.WrapBody[UpdateTemplateReq](h.UpdateTemplate))
	g.POST("/list/pipeline", ginx.Wrap(h.Pipeline))

	g.POST("/by_ids", ginx.WrapBody[FindByTemplateIds](h.FindByTemplateIds))

	// 根据流程ID，获取所有已经被模版，主要为了处理模版
	g.POST("/get_by_workflow_id", ginx.WrapBody[GetTemplatesByWorkFlowIdReq](h.GetTemplatesByWorkflowId))

	g.POST("/rules/by_workflow_id", ginx.WrapBody[GetRulesByWorkFlowIdReq](h.GetRulesByWorkFlowId))
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

func (h *Handler) FindByTemplateIds(ctx *gin.Context, req FindByTemplateIds) (ginx.Result, error) {
	if len(req.Ids) < 0 {
		return systemErrorResult, fmt.Errorf("输入为空，不符合要求")
	}

	ts, err := h.svc.FindByTemplateIds(ctx, req.Ids)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "获取多个用户信息成功",
		Data: RetrieveTemplates{
			Total: int64(len(ts)),
			Templates: slice.Map(ts, func(idx int, src domain.Template) TemplateJson {
				return h.toTemplateJsonVo(src)
			}),
		},
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

func (h *Handler) GetRulesByWorkFlowId(ctx *gin.Context, req GetRulesByWorkFlowIdReq) (ginx.Result, error) {
	wfs, err := h.svc.GetByWorkflowId(ctx, req.WorkFlowId)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "查询流程绑定的模版成功",
		Data: RetrieveTemplateRules{
			TemplateRules: slice.Map(wfs, func(idx int, src domain.Template) TemplateRules {
				rs, _ := rule.ParseRules(src.Rules)
				r := slice.Map(rs, func(idx int, src rule.Rule) Rule {
					return Rule{
						Type:  src.Type,
						Field: src.Field,
						Title: src.Title,
					}
				})

				return TemplateRules{
					Rules: r,
					Id:    src.Id,
					Name:  src.Name,
				}
			}),
		},
	}, nil
}

func (h *Handler) GetTemplatesByWorkflowId(ctx *gin.Context, req GetTemplatesByWorkFlowIdReq) (ginx.Result, error) {
	wfs, err := h.svc.GetByWorkflowId(ctx, req.WorkFlowId)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "查询流程绑定的模版成功",
		Data: RetrieveTemplates{
			Templates: slice.Map(wfs, func(idx int, src domain.Template) TemplateJson {
				return h.toTemplateJsonVo(src)
			}),
		},
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
			Templates: slice.Map(rts, func(idx int, src domain.Template) TemplateJson {
				return h.toTemplateJsonVo(src)
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
	if req.Rules != "" {
		if err := json.Unmarshal([]byte(req.Rules), &rulesData); err != nil {
			return domain.Template{}, err
		}
	}
	var optionsData map[string]interface{}
	if req.Options != "" {
		if err := json.Unmarshal([]byte(req.Options), &optionsData); err != nil {
			return domain.Template{}, err
		}
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

//func (h *Handler) toTemplateVo(req domain.Template) Template {
//	return Template{
//		Id:         req.Id,
//		Name:       req.Name,
//		WorkflowId: req.WorkflowId,
//		GroupId:    req.GroupId,
//		Icon:       req.Icon,
//		Rules:      req.Rules,
//		Options:    req.Options,
//		CreateType: CreateType(req.CreateType),
//		Desc:       req.Desc,
//	}
//}

func (h *Handler) toTemplateVo(req domain.Template) Template {
	rules, _ := json.Marshal(req.Rules)

	options, _ := json.Marshal(req.Options)
	return Template{
		Id:         req.Id,
		Name:       req.Name,
		WorkflowId: req.WorkflowId,
		GroupId:    req.GroupId,
		Icon:       req.Icon,
		Rules:      string(rules),
		Options:    string(options),
		CreateType: CreateType(req.CreateType),
		Desc:       req.Desc,
	}
}

func (h *Handler) toTemplateJsonVo(req domain.Template) TemplateJson {
	return TemplateJson{
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
	if req.Rules != "" {
		if err := json.Unmarshal([]byte(req.Rules), &rulesData); err != nil {
			return domain.Template{}, err
		}
	}
	var optionsData map[string]interface{}
	if req.Options != "" {
		if err := json.Unmarshal([]byte(req.Options), &optionsData); err != nil {
			return domain.Template{}, err
		}
	}

	return domain.Template{
		Id:         req.Id,
		Name:       req.Name,
		Desc:       req.Desc,
		Icon:       req.Icon,
		GroupId:    req.GroupId,
		WorkflowId: req.WorkflowId,
		Rules:      rulesData,
		Options:    optionsData,
	}, nil

}
