package web

import (
	"github.com/Duke1616/ecmdb/internal/model/internal/domain"
	"github.com/Duke1616/ecmdb/internal/model/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc   service.Service
	mgSvc service.MGService
}

func NewHandler(svc service.Service, groupSvc service.MGService) *Handler {
	return &Handler{
		svc:   svc,
		mgSvc: groupSvc,
	}
}

func (h *Handler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/model")
	// 模型分组
	g.POST("/group/create", ginx.WrapBody[CreateModelGroupReq](h.CreateGroup))

	// 模型操作
	g.POST("/create", ginx.WrapBody[CreateModelReq](h.CreateModel))
	g.POST("/detail", ginx.WrapBody[DetailModelReq](h.DetailModel))
	g.POST("/list", ginx.WrapBody[Page](h.ListModels))
}

func (h *Handler) CreateGroup(ctx *gin.Context, req CreateModelGroupReq) (ginx.Result, error) {
	id, err := h.mgSvc.CreateModelGroup(ctx.Request.Context(), domain.ModelGroup{
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
		Icon:    req.Icon,
	})

	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: id,
		Msg:  "添加模型成功",
	}, nil
}

func (h *Handler) DetailModel(ctx *gin.Context, req DetailModelReq) (ginx.Result, error) {
	model, err := h.svc.FindModelById(ctx, req.ID)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: model,
		Msg:  "模型查找成功",
	}, nil
}

func (h *Handler) ListModels(ctx *gin.Context, req Page) (ginx.Result, error) {
	models, total, err := h.svc.ListModels(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: RetrieveModelsListResp{
			Total: total,
			Models: slice.Map(models, func(idx int, m domain.Model) Model {
				return toModelVo(m)
			}),
		},
	}, nil
}
