package web

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	service "github.com/Duke1616/ecmdb/internal/service/tools"
	"github.com/Duke1616/eiam/pkg/web/capability"
	"github.com/ecodeclub/ginx"
	"github.com/gin-gonic/gin"
)

var _ ginx.Handler = &Handler{}

const DefaultBucket = "ecmdb"

type Handler struct {
	svc service.Service
	capability.IRegistry
}

func NewHandler(svc service.Service) *Handler {
	return &Handler{
		svc:       svc,
		IRegistry: capability.NewRegistry("cmdb", "tools", "资产仓库/文件管理"),
	}
}

func (h *Handler) PublicRoutes(server *gin.Engine) {
	g := server.Group("/api/tools")
	g.GET("/download/:filename", h.Download)
}

func (h *Handler) IdentifyRoutes(_ *gin.Engine) {}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/tools")
	g.POST("/upload", h.Define("文件上传", "upload").
		Bind(h.Upload),
	)

	g.POST("/minio/get_presigned_url", h.Define("获取下载预签名", "get_presigned_url").
		NoSync().
		Bind(ginx.B[GetPresignedUrl](h.GetPresignedUrl)),
	)
	g.POST("/minio/put_presigned_url", h.Define("获取上传预签名", "put_presigned_url").
		NoSync().
		Bind(ginx.B[PutPresignedUrl](h.PutPresignedUrl)),
	)
	g.POST("/minio/object/remove", h.Define("删除对象", "remove").
		Bind(ginx.B[RemoveObjectReq](h.RemoveObject)),
	)
}

func (h *Handler) Upload(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无法获取文件"})
		return
	}

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "无法打开文件"})
		return
	}

	defer src.Close()

	// 定义保存路径
	dst := filepath.Join("./uploads", file.Filename)
	if err = os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建目录"})
		return
	}

	out, err := os.Create(dst)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建文件"})
		return
	}
	defer out.Close()

	// 缓冲区读取并写入 4MB
	buffer := make([]byte, 4*1024*1024)
	for {
		n, readErr := src.Read(buffer)
		if n > 0 {
			if _, writeErr := out.Write(buffer[:n]); writeErr != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "写入文件失败"})
				return
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
			return
		}
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

func (h *Handler) PutPresignedUrl(ctx *ginx.Context, req PutPresignedUrl) (ginx.Result, error) {
	bucket := h.bucket(req.Bucket)
	key, url, err := h.svc.PutPresignedUrl(ctx.Context, bucket, req.Prefix, req.ObjectName)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: map[string]string{
			"url":         url,
			"bucket":      bucket,
			"object_name": key,
		},
	}, nil
}

func (h *Handler) GetPresignedUrl(ctx *ginx.Context, req GetPresignedUrl) (ginx.Result, error) {
	bucket := h.bucket(req.Bucket)
	url, err := h.svc.GetPresignedUrl(ctx.Context, bucket, req.ObjectName)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: map[string]string{
			"url":    url,
			"bucket": bucket,
		},
	}, nil
}

func (h *Handler) RemoveObject(ctx *ginx.Context, req RemoveObjectReq) (ginx.Result, error) {
	bucket := h.bucket(req.Bucket)
	err := h.svc.RemoveObject(ctx.Context, bucket, req.ObjectName)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: true,
		Msg:  "删除对象成功",
	}, nil
}

func (h *Handler) bucket(bucket string) string {
	if bucket == "" {
		bucket = DefaultBucket
	}

	return bucket
}

