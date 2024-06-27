package web

import (
	"github.com/Duke1616/ecmdb/internal/codebook"
	"github.com/Duke1616/ecmdb/internal/runner/internal/domain"
	"github.com/Duke1616/ecmdb/internal/runner/internal/service"
	"github.com/Duke1616/ecmdb/internal/worker"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc         service.Service
	workerSvc   worker.Service
	codebookSvc codebook.Service
}

func NewHandler(svc service.Service, workerSvc worker.Service, codebookSvc codebook.Service) *Handler {
	return &Handler{
		svc:         svc,
		workerSvc:   workerSvc,
		codebookSvc: codebookSvc,
	}
}

func (h *Handler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/api/runner")
	g.POST("/register", ginx.WrapBody[RegisterRunnerReq](h.Register))
	g.POST("/list", ginx.WrapBody[ListRunnerReq](h.ListRunner))
}

func (h *Handler) Register(ctx *gin.Context, req RegisterRunnerReq) (ginx.Result, error) {
	//  验证代码模版密钥是否正确
	exist, err := h.codebookSvc.ValidationSecret(ctx, req.TaskIdentifier, req.TaskSecret)
	if exist != true {
		return systemErrorResult, err
	}

	// 验证节点是否存在
	exist, err = h.workerSvc.ValidationByName(ctx, req.WorkName)
	if exist != true {
		return validationErrorResult, err
	}

	_, err = h.svc.Register(ctx, h.toDomain(req))
	if err != nil {
		return validationErrorResult, err
	}
	return ginx.Result{
		Msg: "注册成功",
	}, nil
}

func (h *Handler) ListRunner(ctx *gin.Context, req ListRunnerReq) (ginx.Result, error) {
	ws, total, err := h.svc.ListRunner(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "查询 runner 列表成功",
		Data: RetrieveWorkers{
			Total: total,
			Runners: slice.Map(ws, func(idx int, src domain.Runner) Runner {
				return h.toRunnerVo(src)
			}),
		},
	}, nil
}

func (h *Handler) toDomain(req RegisterRunnerReq) domain.Runner {
	return domain.Runner{
		Name:           req.Name,
		TaskIdentifier: req.TaskIdentifier,
		TaskSecret:     req.TaskSecret,
		WorkName:       req.WorkName,
		Tags:           req.Tags,
		Action:         domain.Action(REGISTER),
	}
}

func (h *Handler) toRunnerVo(req domain.Runner) Runner {
	return Runner{
		Id:     req.Id,
		Name:   req.Name,
		Tags:   req.Tags,
		Desc:   req.Desc,
		Worker: req.WorkName,
	}
}
