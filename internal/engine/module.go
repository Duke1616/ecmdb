package engine

import "github.com/Duke1616/ecmdb/internal/engine/internal/web"

type Module struct {
	Svc Service
	Hdl *web.Handler
}
