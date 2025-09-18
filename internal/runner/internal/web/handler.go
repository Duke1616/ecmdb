package web

import (
	"context"
	"encoding/json"

	"github.com/Duke1616/ecmdb/internal/codebook"
	"github.com/Duke1616/ecmdb/internal/runner/internal/domain"
	"github.com/Duke1616/ecmdb/internal/runner/internal/service"
	"github.com/Duke1616/ecmdb/internal/worker"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc         service.Service
	workerSvc   worker.Service
	workflowSvc workflow.Service
	codebookSvc codebook.Service
	crypto      cryptox.Crypto[string]
}

func NewHandler(svc service.Service, workerSvc worker.Service, workflowSvc workflow.Service,
	codebookSvc codebook.Service, crypto cryptox.Crypto[string]) *Handler {
	return &Handler{
		svc:         svc,
		workerSvc:   workerSvc,
		workflowSvc: workflowSvc,
		codebookSvc: codebookSvc,
		crypto:      crypto,
	}
}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/runner")
	g.POST("/register", ginx.WrapBody[RegisterRunnerReq](h.Register))
	g.POST("/list", ginx.WrapBody[ListRunnerReq](h.ListRunner))
	g.POST("/list/tags", ginx.Wrap(h.ListTags))
	g.POST("/update", ginx.WrapBody[UpdateRunnerReq](h.UpdateRunner))
	g.POST("/delete", ginx.WrapBody[DeleteRunnerReq](h.DeleteRunner))
	g.POST("/list/by_workflow_id", ginx.WrapBody[ListByWorkflowIdReq](h.ListByWorkflowId))
	g.POST("/list/by_ids", ginx.WrapBody[ListRunnerByIds](h.ListByIds))
}

func (h *Handler) Register(ctx *gin.Context, req RegisterRunnerReq) (ginx.Result, error) {
	// 数据校验
	err := h.validation(ctx, req.CodebookUid, req.CodebookSecret, req.WorkerName)
	if err != nil {
		return systemErrorResult, err
	}

	// 注册
	_, err = h.svc.Register(ctx, h.toDomain(req))
	if err != nil {
		return validationErrorResult, err
	}
	return ginx.Result{
		Msg: "注册成功",
	}, nil
}

func (h *Handler) ListByIds(ctx *gin.Context, req ListRunnerByIds) (ginx.Result, error) {
	rs, err := h.svc.ListByIds(ctx, req.Ids)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "查询 runner 列表成功",
		Data: RetrieveWorkers{
			Total: int64(len(rs)),
			Runners: slice.Map(rs, func(idx int, src domain.Runner) Runner {
				return h.toRunnerVo(src)
			}),
		},
	}, nil
}

func (h *Handler) ListByWorkflowId(ctx *gin.Context, req ListByWorkflowIdReq) (ginx.Result, error) {
	wf, err := h.workflowSvc.Find(ctx, req.WorkflowId)
	if err != nil {
		return systemErrorResult, err
	}

	nodesJSON, err := json.Marshal(wf.FlowData.Nodes)
	if err != nil {
		return systemErrorResult, err
	}
	var nodes []easyflow.Node
	err = json.Unmarshal(nodesJSON, &nodes)
	if err != nil {
		return systemErrorResult, err
	}

	codebookUids := make([]string, 0)
	for _, node := range nodes {
		if node.Type == "automation" {
			property, _ := easyflow.ToNodeProperty[easyflow.AutomationProperty](node)
			codebookUids = append(codebookUids, property.CodebookUid)
		}
	}

	if len(codebookUids) == 0 {
		return ginx.Result{Msg: "此模版暂未绑定 【任务模版】", Code: 500102}, nil
	}

	rs, err := h.svc.ListByCodebookUids(ctx, codebookUids)
	if len(rs) == 0 {
		return ginx.Result{Msg: "此模版暂未绑定 【执行器】", Code: 500103}, nil
	}

	return ginx.Result{
		Msg: "查询 runner 列表成功",
		Data: RetrieveWorkers{
			Total: int64(len(rs)),
			Runners: slice.Map(rs, func(idx int, src domain.Runner) Runner {
				return h.toRunnerVo(src)
			}),
		},
	}, nil
}

func (h *Handler) DeleteRunner(ctx *gin.Context, req DeleteRunnerReq) (ginx.Result, error) {
	count, err := h.svc.Delete(ctx, req.Id)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: count,
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

func (h *Handler) UpdateRunner(ctx *gin.Context, req UpdateRunnerReq) (ginx.Result, error) {
	// 数据校验
	err := h.validation(ctx, req.CodebookUid, req.CodebookSecret, req.WorkerName)
	if err != nil {
		return validationErrorResult, err
	}

	// 注册
	runner, err := h.toUpdateDomain(ctx, req)
	if err != nil {
		return systemErrorResult, err
	}

	_, err = h.svc.Update(ctx, runner)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "修改成功",
	}, nil
}

func (h *Handler) validation(ctx context.Context, codebookUid, codebookSecret, workerName string) error {
	//  验证代码模版密钥是否正确
	exist, err := h.codebookSvc.ValidationSecret(ctx, codebookUid, codebookSecret)
	if exist != true {
		return err
	}

	// 验证节点是否存在
	exist, err = h.workerSvc.ValidationByName(ctx, workerName)
	if exist != true {
		return err
	}

	return nil
}

func (h *Handler) ListTags(ctx *gin.Context) (ginx.Result, error) {
	tags, err := h.svc.ListTagsPipelineByCodebookUid(ctx)
	if err != nil {
		return ginx.Result{}, err
	}

	codeUids := slice.Map(tags, func(idx int, src domain.RunnerTags) string {
		return src.CodebookUid
	})
	codes, err := h.codebookSvc.FindByUids(ctx, codeUids)
	if err != nil {
		return ginx.Result{}, err
	}
	codeMaps := slice.ToMapV(codes, func(element codebook.Codebook) (string, string) {
		return element.Identifier, element.Name
	})

	return ginx.Result{
		Msg: "查询 runner tags 列表成功",
		Data: RetrieveRunnerTags{
			RunnerTags: slice.Map(tags, func(idx int, src domain.RunnerTags) RunnerTags {
				codeName, _ := codeMaps[src.CodebookUid]
				return RunnerTags{
					TagsMappingTopic: src.TagsMappingTopic,
					CodebookUid:      src.CodebookUid,
					CodebookName:     codeName,
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
		Topic:          req.Topic,
		WorkerName:     req.WorkerName,
		Tags:           req.Tags,
		Variables:      h.toVariablesDomain(req.Variables),
		Action:         domain.Action(REGISTER),
	}
}

func (h *Handler) toUpdateDomain(ctx context.Context, req UpdateRunnerReq) (domain.Runner, error) {
	runner, err := h.svc.Detail(ctx, req.Id)
	if err != nil {
		return domain.Runner{}, err
	}

	oldVars := slice.ToMap(runner.Variables, func(element domain.Variables) string {
		return element.Key
	})

	return domain.Runner{
		Id:             req.Id,
		Name:           req.Name,
		CodebookSecret: req.CodebookSecret,
		CodebookUid:    req.CodebookUid,
		WorkerName:     req.WorkerName,
		Topic:          req.Topic,
		Tags:           req.Tags,
		Variables:      h.toUpdateVariablesDomain(oldVars, req.Variables),
		Action:         domain.Action(REGISTER),
	}, nil
}

func (h *Handler) toUpdateVariablesDomain(oldVars map[string]domain.Variables, req []Variables) []domain.Variables {
	return slice.Map(req, func(idx int, src Variables) domain.Variables {
		value := src.Value
		if src.Secret {
			val, ok := oldVars[src.Key]
			if ok && src.Value == "" {
				value = val.Value
			} else {
				aesVal, err := h.crypto.Encrypt(src.Value)
				if err != nil {
					return domain.Variables{}
				}
				value = aesVal
			}
		}

		return domain.Variables{
			Key:    src.Key,
			Secret: src.Secret,
			Value:  value,
		}
	})
}

func (h *Handler) toVariablesDomain(req []Variables) []domain.Variables {
	return slice.Map(req, func(idx int, src Variables) domain.Variables {
		val := src.Value
		if src.Secret {
			// 如果加密失败就存储原始存储
			aesVal, err := h.crypto.Encrypt(src.Value)
			if err != nil {
				return domain.Variables{}
			}

			val = aesVal
		}

		return domain.Variables{
			Key:    src.Key,
			Secret: src.Secret,
			Value:  val,
		}
	})
}

func (h *Handler) toRunnerVo(req domain.Runner) Runner {
	return Runner{
		Id:          req.Id,
		Name:        req.Name,
		CodebookUid: req.CodebookUid,
		Tags:        req.Tags,
		Desc:        req.Desc,
		Variables: slice.Map(req.Variables, func(idx int, src domain.Variables) Variables {
			if src.Secret {
				return Variables{
					Key:    src.Key,
					Secret: src.Secret,
				}
			}
			return Variables{
				Key:    src.Key,
				Secret: src.Secret,
				Value:  src.Value,
			}
		}),
		WorkerName: req.WorkerName,
	}
}
