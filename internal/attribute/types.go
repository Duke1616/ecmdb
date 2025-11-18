package attribute

import (
	"github.com/Duke1616/ecmdb/internal/attribute/internal/domain"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/event"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/service"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/web"
)

type Handler = web.Handler

type Service = service.Service

type Attribute = domain.Attribute

type AttributeGroup = domain.AttributeGroup

type Event = event.FieldSecureAttrChange
