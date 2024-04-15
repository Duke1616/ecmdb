package web

import (
	"github.com/Duke1616/ecmdb/internal/model/internal/domain"
	"github.com/Duke1616/ecmdb/internal/model/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"time"
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
	g := server.Group("/model")

	g.POST("/group/create", ginx.WrapBody[CreateModelGroupReq](h.CreateGroup))
	g.POST("/create", ginx.WrapBody[CreateModelReq](h.CreateModel))
	g.POST("/detail", ginx.WrapBody[DetailUidModelReq](h.DetailModel))
	g.POST("/list", ginx.WrapBody[ListModelsReq](h.ListModels))
}

func (h *Handler) CreateGroup(ctx *gin.Context, req CreateModelGroupReq) (ginx.Result, error) {
	id, err := h.svc.CreateModelGroup(ctx.Request.Context(), domain.ModelGroup{
		Name: req.Name,
	})

	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: id,
		Msg:  "添加模型分组成功",
	}, nil
}

func (h *Handler) CreateModel(ctx *gin.Context, req CreateModelReq) (ginx.Result, error) {
	id, err := h.svc.CreateModel(ctx, domain.Model{
		Name:    req.Name,
		GroupId: req.GroupId,
		UID:     req.UID,
	})

	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: id,
		Msg:  "添加模型成功",
	}, nil
}

func (h *Handler) DetailModel(ctx *gin.Context, req DetailUidModelReq) (ginx.Result, error) {
	model, err := h.svc.FindModelByUid(ctx, req.uid)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: model,
		Msg:  "模型查找成功",
	}, nil
}

func (h *Handler) ListModels(ctx *gin.Context, req ListModelsReq) (ginx.Result, error) {
	models, total, err := h.svc.ListModels(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: h.toCaseList(models, total),
	}, nil
}

func (h *Handler) toCaseList(data []domain.Model, total int64) ListModelsResp {
	return ListModelsResp{
		Total: total,
		Models: slice.Map(data, func(idx int, m domain.Model) Model {
			return newModel(m)
		}),
	}
}

func newModel(m domain.Model) Model {
	return Model{
		Name:  m.Name,
		UID:   m.UID,
		Ctime: m.Utime.Format(time.DateTime),
		Utime: m.Utime.Format(time.DateTime),
	}
}
