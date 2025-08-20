package service

import (
	"context"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
)

type Service interface {
	GetPresignedUrl(ctx context.Context, bucketName string, objectName string) (*url.URL, error)
	PutPresignedUrl(ctx context.Context, bucketName string, objectName string) (*url.URL, error)
	RemoveObject(ctx context.Context, bucketName string, objetName string) error
}

type service struct {
	minioClient *minio.Client
	expires     time.Duration
}

func NewService(minioClient *minio.Client) Service {
	return &service{
		minioClient: minioClient,
		expires:     time.Minute * 2,
	}
}

func (s *service) PutPresignedUrl(ctx context.Context, bucketName string, objectName string) (*url.URL, error) {
	return s.minioClient.PresignedPutObject(ctx, bucketName, objectName, s.expires)
}

func (s *service) GetPresignedUrl(ctx context.Context, bucketName string, objectName string) (*url.URL, error) {
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", "attachment; filename="+objectName)
	return s.minioClient.PresignedGetObject(ctx, bucketName, objectName, s.expires, reqParams)
}

func (s *service) RemoveObject(ctx context.Context, bucketName string, objetName string) error {
	return s.minioClient.RemoveObject(ctx, bucketName, objetName, minio.RemoveObjectOptions{
		ForceDelete:      true,
		GovernanceBypass: true,
	})
}
