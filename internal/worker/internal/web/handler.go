package web

import (
	"github.com/Duke1616/ecmdb/internal/worker/internal/domain"
	"github.com/Duke1616/ecmdb/internal/worker/internal/service"
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

func (h *Handler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/api/worker")
	g.POST("/list", ginx.WrapBody[ListWorkerReq](h.ListWorker))
	g.POST("/push", ginx.WrapBody[PushMessageReq](h.PushMessage))
}

func (h *Handler) ListWorker(ctx *gin.Context, req ListWorkerReq) (ginx.Result, error) {
	ws, total, err := h.svc.ListWorker(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "查询工单模版列表成功",
		Data: RetrieveWorkers{
			Total: total,
			Workers: slice.Map(ws, func(idx int, src domain.Worker) Worker {
				return h.toWorkerVo(src)
			}),
		},
	}, nil
}

func (h *Handler) PushMessage(ctx *gin.Context, req PushMessageReq) (ginx.Result, error) {
	err := h.svc.Execute(ctx, h.toDomain(req))
	if err != nil {
		return ginx.Result{}, err
	}

	return ginx.Result{
		Msg: "",
	}, nil
}

func (h *Handler) toDomain(req PushMessageReq) domain.Execute {
	return domain.Execute{
		Topic:    req.Topic,
		Name:     req.Name,
		UUID:     req.UUID,
		Language: req.Language,
		Code:     req.Code,
	}
}

func (h *Handler) toWorkerVo(req domain.Worker) Worker {
	return Worker{
		Id:     req.Id,
		Name:   req.Name,
		Desc:   req.Desc,
		Topic:  req.Topic,
		Status: Status(req.Status),
	}
}
