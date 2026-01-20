package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
)

// S3Storage S3 存储服务实现(基于 MinIO)
type S3Storage struct {
	client *minio.Client
	bucket string
	prefix string // 文件前缀,如 "exchange/"
}

// S3Config S3 配置
type S3Config struct {
	Bucket string `mapstructure:"bucket"` // Bucket 名称
	Prefix string `mapstructure:"prefix"` // 文件前缀,如 "exchange/"
}

// NewS3Storage 创建 S3 存储服务
func NewS3Storage(client *minio.Client) *S3Storage {
	var cfg S3Config
	if err := viper.UnmarshalKey("s3.exchange", &cfg); err != nil {
		panic(err)
	}

	return &S3Storage{
		client: client,
		bucket: cfg.Bucket,
		prefix: cfg.Prefix,
	}
}

// GenerateUploadURL 生成预签名上传 URL
func (s *S3Storage) GenerateUploadURL(ctx context.Context, fileName string, expireSeconds int) (string, string, error) {
	// 构建完整的文件路径
	fileKey := s.buildFileKey(fileName)

	// 生成预签名 PUT URL
	presignedURL, err := s.client.PresignedPutObject(ctx, s.bucket, fileKey, time.Duration(expireSeconds)*time.Second)
	if err != nil {
		return "", "", fmt.Errorf("生成上传链接失败: %w", err)
	}

	return fileKey, presignedURL.String(), nil
}

// GenerateDownloadURL 生成预签名下载 URL
func (s *S3Storage) GenerateDownloadURL(ctx context.Context, fileKey string, expireSeconds int) (string, error) {
	// 生成预签名 URL
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucket, fileKey, time.Duration(expireSeconds)*time.Second, nil)
	if err != nil {
		return "", fmt.Errorf("生成下载链接失败: %w", err)
	}

	return presignedURL.String(), nil
}

// DeleteFile 删除文件
func (s *S3Storage) DeleteFile(ctx context.Context, fileKey string) error {
	err := s.client.RemoveObject(ctx, s.bucket, fileKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("删除 S3 文件失败: %w", err)
	}

	return nil
}

// buildFileKey 构建完整的文件路径
// NOTE: prefix 作为文件夹路径,例如 "exchange/" 会生成 "exchange/20060102_150405_filename.xlsx"
func (s *S3Storage) buildFileKey(fileName string) string {
	// 确保 prefix 以 / 结尾(作为文件夹)
	prefix := s.prefix
	if prefix != "" && prefix[len(prefix)-1] != '/' {
		prefix = prefix + "/"
	}

	// 添加时间戳避免文件名冲突
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s%s_%s", prefix, timestamp, fileName)
}
