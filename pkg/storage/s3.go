package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
)

// S3Storage S3 存储服务实现(基于 MinIO)
type S3Storage struct {
	client *minio.Client
}

// NewS3Storage 创建 S3 存储服务
func NewS3Storage(client *minio.Client) *S3Storage {
	return &S3Storage{
		client: client,
	}
}

// GenerateUploadURL 生成预签名上传 URL
func (s *S3Storage) GenerateUploadURL(ctx context.Context, bucket string, prefix string, fileName string, expireSeconds int) (string, string, error) {
	// 确定 Prefix
	if prefix != "" && prefix[len(prefix)-1] != '/' {
		prefix = prefix + "/"
	}

	// 构建文件路径: prefix/YYYY-MM-DD/fileName
	// 例如: import/2023-10-27/169837482_report.xlsx
	dateDir := time.Now().Format("2006-01-02")
	fileKey := fmt.Sprintf("%s%s/%s", prefix, dateDir, fileName)

	// 生成预签名 PUT URL
	presignedURL, err := s.client.PresignedPutObject(ctx, bucket, fileKey, time.Duration(expireSeconds)*time.Second)
	if err != nil {
		return "", "", fmt.Errorf("生成上传链接失败: %w", err)
	}

	return fileKey, presignedURL.String(), nil
}

// GenerateDownloadURL 生成预签名下载 URL
func (s *S3Storage) GenerateDownloadURL(ctx context.Context, bucket string, fileKey string, expireSeconds int) (string, error) {
	// 生成预签名 URL
	presignedURL, err := s.client.PresignedGetObject(ctx, bucket, fileKey, time.Duration(expireSeconds)*time.Second, nil)
	if err != nil {
		return "", fmt.Errorf("生成下载链接失败: %w", err)
	}

	return presignedURL.String(), nil
}

// DeleteFile 删除文件
func (s *S3Storage) DeleteFile(ctx context.Context, bucket string, fileKey string) error {
	err := s.client.RemoveObject(ctx, bucket, fileKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("删除 S3 文件失败: %w", err)
	}

	return nil
}

// GetFile 获取文件内容
// NOTE: 用于 Excel 导入等场景,直接下载文件到内存
func (s *S3Storage) GetFile(ctx context.Context, bucket string, fileKey string) ([]byte, error) {
	// 获取对象
	object, err := s.client.GetObject(ctx, bucket, fileKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 S3 文件失败: %w", err)
	}
	defer object.Close()

	// 读取文件内容
	var buf []byte
	buf = make([]byte, 0, 1024*1024) // 预分配 1MB
	tmpBuf := make([]byte, 4096)

	for {
		n, err1 := object.Read(tmpBuf)
		if err1 != nil && !errors.Is(err1, io.EOF) {
			return nil, fmt.Errorf("读取 S3 文件内容失败: %w", err1)
		}

		if n > 0 {
			buf = append(buf, tmpBuf[:n]...)
		}

		if err1 != nil {
			break
		}
	}

	return buf, nil
}
