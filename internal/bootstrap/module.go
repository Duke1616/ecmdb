package bootstrap

import "github.com/Duke1616/ecmdb/internal/bootstrap/internal/service"

type Service = service.Loader

type Module struct {
	Svc Service
}
