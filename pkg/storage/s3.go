package storage

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

// IStorage 对象存储接口
type IStorage interface {
	GenerateUploadURL(ctx context.Context, bucket, prefix, fileName string, expireSeconds int) (string, string, error)
	GenerateDownloadURL(ctx context.Context, bucket, fileKey string, expireSeconds int) (string, error)
	DeleteFile(ctx context.Context, bucket, fileKey string) error
	GetFile(ctx context.Context, bucket, fileKey string) ([]byte, error)
}

// S3Storage 兼容类型别名
type S3Storage = IStorage

type s3Storage struct {
	client *minio.Client
}

// NewS3Storage 创建 S3 存储服务实现
func NewS3Storage(client *minio.Client) IStorage {
	return &s3Storage{
		client: client,
	}
}

// GenerateUploadURL 生成预签名上传 URL
func (s *s3Storage) GenerateUploadURL(ctx context.Context, bucket, prefix, fileName string, expireSeconds int) (string, string, error) {
	fileKey := buildObjectKey(prefix, fileName, time.Now())

	presignedURL, err := s.client.PresignedPutObject(ctx, bucket, fileKey, time.Duration(expireSeconds)*time.Second)
	if err != nil {
		return "", "", fmt.Errorf("生成上传链接失败: %w", err)
	}

	return fileKey, presignedURL.String(), nil
}

// buildObjectKey 构建规范化的 S3 对象存储路径: prefix/YYYY-MM-DD/fileName
func buildObjectKey(prefix, fileName string, now time.Time) string {
	cleanPrefix := strings.Trim(prefix, "/")
	dateDir := now.Format("2006-01-02")
	return path.Join(cleanPrefix, dateDir, fileName)
}

// GenerateDownloadURL 生成预签名下载 URL
func (s *s3Storage) GenerateDownloadURL(ctx context.Context, bucket, fileKey string, expireSeconds int) (string, error) {
	presignedURL, err := s.client.PresignedGetObject(ctx, bucket, fileKey, time.Duration(expireSeconds)*time.Second, nil)
	if err != nil {
		return "", fmt.Errorf("生成下载链接失败: %w", err)
	}

	return presignedURL.String(), nil
}

// DeleteFile 删除文件
func (s *s3Storage) DeleteFile(ctx context.Context, bucket, fileKey string) error {
	if err := s.client.RemoveObject(ctx, bucket, fileKey, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("删除 S3 文件失败: %w", err)
	}

	return nil
}

// GetFile 获取文件内容
func (s *s3Storage) GetFile(ctx context.Context, bucket, fileKey string) ([]byte, error) {
	object, err := s.client.GetObject(ctx, bucket, fileKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 S3 文件失败: %w", err)
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("读取 S3 文件内容失败: %w", err)
	}

	return data, nil
}
