package web

import (
	"github.com/Duke1616/ecmdb/internal/rota/internal/domain"
	"github.com/Duke1616/ecmdb/internal/rota/internal/service"
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
	g := server.Group("/api/rota")
	// 排班表
	g.POST("/create", ginx.WrapBody[CreateRotaReq](h.Create))
	g.POST("/list", ginx.WrapBody[ListReq](h.List))
	g.POST("/detail", ginx.WrapBody[DetailReq](h.Detail))
	g.POST("/update", ginx.WrapBody[UpdateReq](h.Update))
	g.POST("/delete", ginx.WrapBody[DeleteReq](h.Delete))

	// 常规排班规则
	g.POST("/rule/shift_scheduling/add", ginx.WrapBody[AddRoleReq](h.AddShiftSchedulingRule))
	g.POST("/rule/shift_scheduling/delete", ginx.WrapBody[AddRoleReq](h.AddShiftSchedulingRule))
	g.POST("/rule/shift_scheduling/update", ginx.WrapBody[UpdateShiftRuleReq](h.UpdateShiftSchedulingRole))

	// 临时排班规则
	g.POST("/rule/shift_adjustment/add", ginx.WrapBody[AddOrUpdateAdjustmentRoleReq](h.AddShiftAdjustmentRule))
	g.POST("/rule/shift_adjustment/delete", ginx.WrapBody[DeleteAdjustmentRoleReq](h.DeleteShiftAdjustmentRule))
	g.POST("/rule/shift_adjustment/update", ginx.WrapBody[AddOrUpdateAdjustmentRoleReq](h.UpdateShiftAdjustmentRule))

	// 查看指定规则
	g.POST("/rule/list_by_id", ginx.WrapBody[DetailById](h.GetRuleListById))

	// 排班数据渲染
	g.POST("/schedule/by_my/list", ginx.Wrap(h.AllMySchedules))
	g.POST("/schedule/preview", ginx.WrapBody[GenerateShiftRosteredReq](h.GenerateShiftRostered))
}

// AddShiftSchedulingRule 新增排班规则
func (h *Handler) AddShiftSchedulingRule(ctx *gin.Context, req AddRoleReq) (ginx.Result, error) {
	id, err := h.svc.AddSchedulingRule(ctx, req.Id, h.toRuleDomain(req.RotaRule))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "新增规则成功",
		Data: id,
	}, nil
}

// GetRuleListById 获取常规规则列表
func (h *Handler) GetRuleListById(ctx *gin.Context, req DetailById) (ginx.Result, error) {
	rota, err := h.svc.Detail(ctx, req.Id)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "获取常规规则列表成功",
		Data: slice.Map(rota.Rules, func(idx int, src domain.RotaRule) RotaRule {
			return h.toVoRule(src)
		}),
	}, nil
}

func (h *Handler) DeleteShiftAdjustmentRule(ctx *gin.Context, req DeleteAdjustmentRoleReq) (ginx.Result, error) {
	id, err := h.svc.DeleteAdjustmentRule(ctx, req.Id, req.GroupId)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "删除临时规则成功",
		Data: id,
	}, nil
}

// GenerateShiftRostered 生成排班表
func (h *Handler) GenerateShiftRostered(ctx *gin.Context, req GenerateShiftRosteredReq) (ginx.Result, error) {
	rostered, err := h.svc.GenerateShiftRostered(ctx, req.Id, req.StartTime, req.EndTime)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: RetrieveShiftRostered{
			FinalSchedule: slice.Map(rostered.FinalSchedule, func(idx int, src domain.Schedule) Schedule {
				return h.toVoSchedule(src)
			}),
			CurrentSchedule: h.toVoSchedule(rostered.CurrentSchedule),
			NextSchedule:    h.toVoSchedule(rostered.NextSchedule),
			Members:         rostered.Members,
		},
	}, nil
}

// AddShiftAdjustmentRule 新增临时排班规则
func (h *Handler) AddShiftAdjustmentRule(ctx *gin.Context, req AddOrUpdateAdjustmentRoleReq) (ginx.Result, error) {
	id, err := h.svc.AddAdjustmentRule(ctx, req.Id, h.toAdjustmentRuleDomain(req.RotaRule))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "新增临时排班规则成功",
		Data: id,
	}, nil
}

// UpdateShiftAdjustmentRule 修改临时排班规则
func (h *Handler) UpdateShiftAdjustmentRule(ctx *gin.Context, req AddOrUpdateAdjustmentRoleReq) (ginx.Result, error) {
	rule, err := h.svc.UpdateAdjustmentRule(ctx, req.Id, req.GroupId, h.toAdjustmentRuleDomain(req.RotaRule))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: rule,
	}, nil
}

func (h *Handler) AllMySchedules(ctx *gin.Context) (ginx.Result, error) {
	return ginx.Result{}, nil
}

// Create 创建排班表
func (h *Handler) Create(ctx *gin.Context, req CreateRotaReq) (ginx.Result, error) {
	id, err := h.svc.Create(ctx, h.toDomain(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "新增值班表成功",
		Data: id,
	}, nil
}

func (h *Handler) List(ctx *gin.Context, req ListReq) (ginx.Result, error) {
	rts, total, err := h.svc.List(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "查询排班列表成功",
		Data: RetrieveRotas{
			Total: total,
			Rotas: slice.Map(rts, func(idx int, src domain.Rota) Rota {
				return h.toVoRota(src)
			}),
		},
	}, nil
}

func (h *Handler) Detail(ctx *gin.Context, req DetailReq) (ginx.Result, error) {
	return ginx.Result{}, nil
}

func (h *Handler) Delete(ctx *gin.Context, req DeleteReq) (ginx.Result, error) {
	return ginx.Result{}, nil
}

func (h *Handler) UpdateShiftSchedulingRole(ctx *gin.Context, req UpdateShiftRuleReq) (ginx.Result, error) {
	role, err := h.svc.UpdateSchedulingRule(ctx, req.Id, slice.Map(req.RotaRules,
		func(idx int, src RotaRule) domain.RotaRule {
			return h.toRuleDomain(src)
		}))
	if err != nil {
		return systemErrorResult, nil
	}

	return ginx.Result{
		Data: role,
	}, nil
}

func (h *Handler) Update(ctx *gin.Context, req UpdateReq) (ginx.Result, error) {
	return ginx.Result{}, nil
}

func (h *Handler) toRuleDomain(req RotaRule) domain.RotaRule {
	return domain.RotaRule{
		RotaGroups: slice.Map(req.RotaGroups, func(idx int, src RotaGroup) domain.RotaGroup {
			return domain.RotaGroup{
				Id:      src.Id,
				Name:    src.Name,
				Members: src.Members,
			}
		}),
		Rotate: domain.Rotate{
			TimeUnit:     domain.TimeUnit(req.Rotate.TimeUnit),
			TimeDuration: req.Rotate.TimeDuration,
		},
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}
}

func (h *Handler) toAdjustmentRuleDomain(req RotaAdjustmentRule) domain.RotaAdjustmentRule {
	return domain.RotaAdjustmentRule{
		RotaGroup: domain.RotaGroup{
			Id:      req.RotaGroup.Id,
			Name:    req.RotaGroup.Name,
			Members: req.RotaGroup.Members,
		},
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}
}

func (h *Handler) toUpdateRuleDomain(req RotaRule) domain.RotaRule {
	return domain.RotaRule{
		RotaGroups: slice.Map(req.RotaGroups, func(idx int, src RotaGroup) domain.RotaGroup {
			return domain.RotaGroup{
				Id:      src.Id,
				Name:    src.Name,
				Members: src.Members,
			}
		}),
		Rotate: domain.Rotate{
			TimeUnit:     domain.TimeUnit(req.Rotate.TimeUnit),
			TimeDuration: req.Rotate.TimeDuration,
		},
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}
}

func (h *Handler) toDomain(req CreateRotaReq) domain.Rota {
	return domain.Rota{
		Name:    req.Name,
		Desc:    req.Desc,
		Owner:   req.Owner,
		Enabled: req.Enabled,
	}
}

func (h *Handler) toVoRota(req domain.Rota) Rota {
	return Rota{
		Id:      req.Id,
		Name:    req.Name,
		Desc:    req.Desc,
		Enabled: req.Enabled,
		Owner:   req.Owner,
		Rules: slice.Map(req.Rules, func(idx int, src domain.RotaRule) RotaRule {
			return h.toVoRule(src)
		}),
		AdjustmentRules: slice.Map(req.AdjustmentRules, func(idx int, src domain.RotaAdjustmentRule) RotaAdjustmentRule {
			return h.toVoAdjustmentRule(src)
		}),
	}
}

func (h *Handler) toVoRule(req domain.RotaRule) RotaRule {
	return RotaRule{
		RotaGroups: slice.Map(req.RotaGroups, func(idx int, src domain.RotaGroup) RotaGroup {
			return RotaGroup{
				Id:      src.Id,
				Name:    src.Name,
				Members: src.Members,
			}
		}),
		Rotate: Rotate{
			TimeUnit:     req.Rotate.TimeUnit.ToUint8(),
			TimeDuration: req.Rotate.TimeDuration,
		},
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}
}

func (h *Handler) toVoAdjustmentRule(req domain.RotaAdjustmentRule) RotaAdjustmentRule {
	return RotaAdjustmentRule{
		RotaGroup: RotaGroup{
			Id:      req.RotaGroup.Id,
			Name:    req.RotaGroup.Name,
			Members: req.RotaGroup.Members,
		},
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}
}

func (h *Handler) toVoSchedule(req domain.Schedule) Schedule {
	return Schedule{
		Title:     req.Title,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		RotaGroup: RotaGroup{
			Id:      req.RotaGroup.Id,
			Name:    req.RotaGroup.Name,
			Members: req.RotaGroup.Members,
		},
	}
}
