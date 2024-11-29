package web

import (
	"github.com/Duke1616/ecmdb/internal/codebook/internal/domain"
	"github.com/Duke1616/ecmdb/internal/codebook/internal/service"
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
	g := server.Group("/api/codebook")
	g.POST("/create", ginx.WrapBody[CreateCodebookReq](h.CreateCodebook))
	g.POST("/list", ginx.WrapBody[ListCodebookReq](h.ListCodebook))
	g.POST("/detail", ginx.WrapBody[DetailCodebookReq](h.DetailCodebook))
	g.POST("/update", ginx.WrapBody[UpdateCodebookReq](h.UpdateCodebook))
	g.POST("/delete", ginx.WrapBody[DeleteCodebookReq](h.DeleteCodebook))
}

func (h *Handler) CreateCodebook(ctx *gin.Context, req CreateCodebookReq) (ginx.Result, error) {
	t, err := h.svc.CreateCodebook(ctx, h.toDomain(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: t,
	}, nil
}

func (h *Handler) DetailCodebook(ctx *gin.Context, req DetailCodebookReq) (ginx.Result, error) {
	t, err := h.svc.DetailCodebook(ctx, req.Id)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: h.toCodebookVo(t),
	}, nil
}

func (h *Handler) ListCodebook(ctx *gin.Context, req ListCodebookReq) (ginx.Result, error) {
	rts, total, err := h.svc.ListCodebook(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "查询工单模版列表成功",
		Data: RetrieveCodebooks{
			Total: total,
			Codebooks: slice.Map(rts, func(idx int, src domain.Codebook) Codebook {
				return h.toCodebookVo(src)
			}),
		},
	}, nil
}

func (h *Handler) UpdateCodebook(ctx *gin.Context, req UpdateCodebookReq) (ginx.Result, error) {
	t, err := h.svc.UpdateCodebook(ctx, h.toUpdateDomain(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: t,
	}, nil
}

func (h *Handler) DeleteCodebook(ctx *gin.Context, req DeleteCodebookReq) (ginx.Result, error) {
	count, err := h.svc.DeleteCodebook(ctx, req.Id)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: count,
	}, nil
}

func (h *Handler) toDomain(req CreateCodebookReq) domain.Codebook {
	return domain.Codebook{
		Name:       req.Name,
		Owner:      req.Owner,
		Code:       req.Code,
		Language:   req.Language,
		Identifier: req.Identifier,
	}
}

func (h *Handler) toUpdateDomain(req UpdateCodebookReq) domain.Codebook {
	return domain.Codebook{
		Id:       req.Id,
		Name:     req.Name,
		Owner:    req.Owner,
		Code:     req.Code,
		Language: req.Language,
	}
}

func (h *Handler) toCodebookVo(req domain.Codebook) Codebook {
	return Codebook{
		Id:         req.Id,
		Name:       req.Name,
		Owner:      req.Owner,
		Code:       req.Code,
		Language:   req.Language,
		Secret:     req.Secret,
		Identifier: req.Identifier,
	}
}
