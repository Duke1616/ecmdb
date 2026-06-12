package service

import (
	"context"
	"time"

	"github.com/Duke1616/ecmdb/pkg/storage"
)

type Service interface {
	GetPresignedUrl(ctx context.Context, bucketName string, objectName string) (string, error)
	PutPresignedUrl(ctx context.Context, bucketName string, prefix string, objectName string) (string, string, error)
	RemoveObject(ctx context.Context, bucketName string, objetName string) error
}

type service struct {
	storage *storage.S3Storage
	expires time.Duration
}

func NewService(storage *storage.S3Storage) Service {
	return &service{
		storage: storage,
		expires: time.Minute * 2,
	}
}

func (s *service) PutPresignedUrl(ctx context.Context, bucketName string, prefix string, objectName string) (string, string, error) {
	key, url, err := s.storage.GenerateUploadURL(ctx, bucketName, prefix, objectName, int(s.expires.Seconds()))
	if err != nil {
		return "", "", err
	}
	return key, url, nil
}

func (s *service) GetPresignedUrl(ctx context.Context, bucketName string, objectName string) (string, error) {
	return s.storage.GenerateDownloadURL(ctx, bucketName, objectName, int(s.expires.Seconds()))
}

func (s *service) RemoveObject(ctx context.Context, bucketName string, objetName string) error {
	return s.storage.DeleteFile(ctx, bucketName, objetName)
}
