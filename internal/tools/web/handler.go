package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) PublicRoutes(server *gin.Engine) {
	g := server.Group("/api/tools")
	g.POST("/upload", h.Upload)
}

func (h *Handler) Upload(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无法获取文件"})
		return
	}

	// 定义保存文件的路径
	savePath := filepath.Join("uploads", file.Filename)

	// 保存文件到指定路径
	if err = ctx.SaveUploadedFile(file, savePath); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "文件保存失败"})
		return
	}

	// 返回上传成功的信息
	ctx.JSON(http.StatusOK, gin.H{"message": "文件上传成功", "url": savePath})
}
