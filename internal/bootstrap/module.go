package bootstrap

import "github.com/Duke1616/ecmdb/internal/service/bootstrap"

type Service = service.Loader

type Module struct {
	Svc Service
}
