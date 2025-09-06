package menu

import (
	"github.com/Duke1616/ecmdb/internal/menu/internal/domain"
	"github.com/Duke1616/ecmdb/internal/menu/internal/event"
	"github.com/Duke1616/ecmdb/internal/menu/internal/service"
	"github.com/Duke1616/ecmdb/internal/menu/internal/web"
)

type Handler = web.Handler

type Service = service.Service

type Menu = domain.Menu

type Meta = domain.Meta

type Type = domain.Type

type Status = domain.Status

type Endpoint = domain.Endpoint

type EventMenuQueue = event.MenuEvent

type EventMenu = event.Menu

type EventEndpoint = event.Endpoint
