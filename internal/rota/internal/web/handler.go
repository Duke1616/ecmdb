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
	g.POST("/rule/shift_scheduling/add", ginx.WrapBody[AddRoleReq](h.AddShiftSchedulingRole))
	g.POST("/rule/shift_scheduling/delete", ginx.WrapBody[AddRoleReq](h.AddShiftSchedulingRole))
	g.POST("/rule/shift_scheduling/update", ginx.WrapBody[AddRoleReq](h.AddShiftSchedulingRole))

	// 临时排班规则
	g.POST("/rule/shift_adjustment/add", ginx.WrapBody[AddRoleReq](h.AddShiftAdjustmentRole))
	g.POST("/rule/shift_adjustment/delete", ginx.WrapBody[AddRoleReq](h.AddShiftAdjustmentRole))
	g.POST("/rule/shift_adjustment/update", ginx.WrapBody[AddRoleReq](h.AddShiftAdjustmentRole))

	// 查看指定规则
	g.POST("/rule/list_by_id", ginx.WrapBody[DetailById](h.GetRuleListById))

	// 排班数据渲染
	g.POST("/schedule/by_my/list", ginx.Wrap(h.AllMySchedules))
	g.POST("/schedule/preview", ginx.Wrap(h.AllMySchedules))
}

// AddShiftSchedulingRole 新增排班规则
func (h *Handler) AddShiftSchedulingRole(ctx *gin.Context, req AddRoleReq) (ginx.Result, error) {
	id, err := h.svc.AddSchedulingRole(ctx, req.Id, h.toRuleDomain(req))
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

// AddShiftAdjustmentRole 新增临时排班规则
func (h *Handler) AddShiftAdjustmentRole(ctx *gin.Context, req AddRoleReq) (ginx.Result, error) {
	id, err := h.svc.AddSchedulingRole(ctx, req.Id, h.toRuleDomain(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "新增临时排班规则成功",
		Data: id,
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

func (h *Handler) Update(ctx *gin.Context, req UpdateReq) (ginx.Result, error) {
	return ginx.Result{}, nil
}

func (h *Handler) toRuleDomain(req AddRoleReq) domain.RotaRule {
	return domain.RotaRule{
		RotaGroups: slice.Map(req.RotaRule.RotaGroups, func(idx int, src RotaGroup) domain.RotaGroup {
			return domain.RotaGroup{
				Id:      src.Id,
				Name:    src.Name,
				Members: src.Members,
			}
		}),
		Rotate: domain.Rotate{
			TimeUnit:     domain.TimeUnit(req.RotaRule.Rotate.TimeUnit),
			TimeDuration: req.RotaRule.Rotate.TimeDuration,
		},
		StartTime: req.RotaRule.StartTime,
		EndTime:   req.RotaRule.EndTime,
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
		TempRules: slice.Map(req.TempRules, func(idx int, src domain.RotaRule) RotaRule {
			return h.toVoRule(src)
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
