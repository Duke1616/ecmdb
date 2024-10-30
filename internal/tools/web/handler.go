package web

import (
	"github.com/Duke1616/ecmdb/pkg/ginx"
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
	g.GET("/download/:filename", h.Download)
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
	ctx.JSON(http.StatusOK, ginx.Result{
		Data: gin.H{"message": "文件上传成功", "filename": file.Filename},
		Msg:  "文件上传成功",
	})
}

func (h *Handler) Download(ctx *gin.Context) {
	filename := ctx.Param("filename")

	// 设置文件的完整路径
	filePath := filepath.Join("uploads", filename)

	// 检查文件是否存在
	if _, err := filepath.Abs(filePath); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "文件未找到"})
		return
	}

	// 设置文件下载的响应头
	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Content-Disposition", "attachment; filename="+filename)
	ctx.Header("Content-Type", "application/octet-stream")

	ctx.File(filePath)
}
