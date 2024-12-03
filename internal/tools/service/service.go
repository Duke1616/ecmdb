package service

import "github.com/minio/minio-go/v7"

type Service interface {
}

type service struct {
	minioClient *minio.Client
}

func NewService(minioClient *minio.Client) Service {
	return service{minioClient: minioClient}
}
