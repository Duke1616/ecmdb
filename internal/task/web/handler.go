package web

import (
	"github.com/Bunny3th/easy-workflow/example/process"
	"github.com/Bunny3th/easy-workflow/workflow/web_api/router"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/gin-gonic/gin"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(server *gin.Engine) {
	g := router.NewRouter(server, "/api/task", false, "")
	g.POST("/create", ginx.WrapBody[CreateReq](h.CreateTask))
}

func (h *Handler) CreateTask(ctx *gin.Context, req CreateReq) (ginx.Result, error) {
	process.CreateExampleProcess()
	return ginx.Result{}, nil
}
