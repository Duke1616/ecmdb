package web

import (
	"github.com/Duke1616/ecmdb/internal/department/internal/domain"
	"github.com/Duke1616/ecmdb/internal/department/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
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
	g := server.Group("/api/department")

	g.POST("/create", ginx.WrapBody[CreateDepartmentReq](h.CreateDepartment))
	g.POST("/update", ginx.WrapBody[UpdateDepartmentReq](h.UpdateDepartment))
	g.POST("/list/tree", ginx.Wrap(h.ListTreeDepartment))
}

func (h *Handler) CreateDepartment(ctx *gin.Context, req CreateDepartmentReq) (ginx.Result, error) {
	id, err := h.svc.CreateDepartment(ctx, h.toDomain(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: id,
	}, nil
}

func (h *Handler) UpdateDepartment(ctx *gin.Context, req UpdateDepartmentReq) (ginx.Result, error) {
	id, err := h.svc.UpdateDepartment(ctx, h.toUpdateDomain(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: id,
	}, nil
}

func (h *Handler) ListTreeDepartment(ctx *gin.Context) (ginx.Result, error) {
	dms, err := h.svc.ListDepartment(ctx)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "获取部门树成功",
		Data: GetDepartmentsTree(dms),
	}, nil
}

func (h *Handler) toUpdateDomain(req UpdateDepartmentReq) domain.Department {
	return domain.Department{
		Id:         req.Id,
		Pid:        req.Pid,
		Name:       req.Name,
		Sort:       req.Sort,
		Enabled:    req.Enabled,
		Leaders:    req.Leaders,
		MainLeader: req.MainLeader,
	}
}

func (h *Handler) toDomain(req CreateDepartmentReq) domain.Department {
	return domain.Department{
		Pid:        req.Pid,
		Name:       req.Name,
		Sort:       req.Sort,
		Enabled:    req.Enabled,
		Leaders:    req.Leaders,
		MainLeader: req.MainLeader,
	}
}
