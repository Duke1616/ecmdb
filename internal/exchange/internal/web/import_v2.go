package web

import (
	"fmt"
	"io"

	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/gin-gonic/gin"
)

// ImportV2 导入数据 V2 (直接上传文件)
// NOTE: 用于快速测试,前端直接上传文件,不经过 S3
func (h *Handler) ImportV2(ctx *gin.Context) (ginx.Result, error) {
	// 1. 获取 model_uid 参数
	modelUID := ctx.PostForm("model_uid")
	if modelUID == "" {
		return ginx.Result{
			Code: 400,
			Msg:  "model_uid 参数不能为空",
		}, nil
	}

	// 2. 获取上传的文件
	file, err := ctx.FormFile("file")
	if err != nil {
		return systemErrorResult, fmt.Errorf("获取上传文件失败: %w", err)
	}

	// 3. 读取文件内容
	f, err := file.Open()
	if err != nil {
		return systemErrorResult, fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	fileData, err := io.ReadAll(f)
	if err != nil {
		return systemErrorResult, fmt.Errorf("读取文件内容失败: %w", err)
	}

	// 4. 调用 Service 导入数据
	importedCount, err := h.svc.Import(ctx.Request.Context(), modelUID, fileData)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "导入成功",
		Data: map[string]interface{}{
			"imported_count": importedCount,
			"file_name":      file.Filename,
		},
	}, nil
}
