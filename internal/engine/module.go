package engine

import (
	engineEvent "github.com/Duke1616/ecmdb/internal/engine/internal/event/engine"
	"github.com/Duke1616/ecmdb/internal/engine/internal/web"
)

type Module struct {
	Svc   Service
	Hdl   *web.Handler
	event *engineEvent.ProcessEvent
}
