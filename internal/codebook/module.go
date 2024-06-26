package codebook

import "github.com/Duke1616/ecmdb/internal/codebook/internal/service"

type Module struct {
	Hdl *Handler
	Svc service.Service
}
