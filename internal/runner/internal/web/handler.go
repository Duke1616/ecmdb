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
	g.POST("/list/tags", ginx.Wrap(h.ListTags))
}

func (h *Handler) Register(ctx *gin.Context, req RegisterRunnerReq) (ginx.Result, error) {
	//  验证代码模版密钥是否正确
	exist, err := h.codebookSvc.ValidationSecret(ctx, req.CodebookUid, req.CodebookSecret)
	if exist != true {
		return systemErrorResult, err
	}

	// 验证节点是否存在
	exist, err = h.workerSvc.ValidationByName(ctx, req.WorkerName)
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

func (h *Handler) ListTags(ctx *gin.Context) (ginx.Result, error) {
	tags, err := h.svc.ListTagsPipelineByCodebookUid(ctx)
	if err != nil {
		return ginx.Result{}, err
	}

	return ginx.Result{
		Msg: "查询 runner tags 列表成功",
		Data: RetrieveRunnerTags{
			RunnerTags: slice.Map(tags, func(idx int, src domain.RunnerTags) RunnerTags {
				return RunnerTags{
					CodebookUid: src.CodebookUid,
					Tags:        src.Tags,
				}
			}),
		},
	}, nil
}

func (h *Handler) toDomain(req RegisterRunnerReq) domain.Runner {
	return domain.Runner{
		Name:           req.Name,
		CodebookSecret: req.CodebookSecret,
		CodebookUid:    req.CodebookUid,
		WorkerName:     req.WorkerName,
		Tags:           req.Tags,
		Variables: slice.Map(req.Variables, func(idx int, src Variables) domain.Variables {
			return domain.Variables{
				Key:   src.Key,
				Value: src.Value,
			}
		}),
		Action: domain.Action(REGISTER),
	}
}

func (h *Handler) toRunnerVo(req domain.Runner) Runner {
	return Runner{
		Id:          req.Id,
		Name:        req.Name,
		CodebookUid: req.CodebookUid,
		Tags:        req.Tags,
		Desc:        req.Desc,
		Variables: slice.Map(req.Variables, func(idx int, src domain.Variables) Variables {
			return Variables{
				Key:   src.Key,
				Value: src.Value,
			}
		}),
		WorkerName: req.WorkerName,
	}
}
