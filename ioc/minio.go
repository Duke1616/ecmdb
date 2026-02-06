package ioc

import (
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/viper"
)

func InitMinioClient() *minio.Client {
	type Config struct {
		Endpoint        string `mapstructure:"endpoint"`
		AccessKeyId     string `mapstructure:"access_key_id"`
		SecretAccessKey string `mapstructure:"secret_access_key"`
		UseSSL          bool   `mapstructure:"use_ssl"`
	}
	var cfg Config
	if err := viper.UnmarshalKey("minio", &cfg); err != nil {
		panic(fmt.Errorf("unable to decode into structure: %v", err))
	}

	// 初始化 Minio 客户端
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyId, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})

	if err != nil {
		panic(err)
	}
	return minioClient
}
