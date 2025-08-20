//go:build wireinject

package permission

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/menu"
	"github.com/Duke1616/ecmdb/internal/permission/internal/event"
	"github.com/Duke1616/ecmdb/internal/permission/internal/service"
	"github.com/Duke1616/ecmdb/internal/permission/internal/web"
	"github.com/Duke1616/ecmdb/internal/policy"
	"github.com/Duke1616/ecmdb/internal/role"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
)

func InitModule(db *mongox.Mongo, q mq.MQ, roleModule *role.Module, menuModule *menu.Module, policyModule *policy.Module) (*Module, error) {
	wire.Build(
		web.NewHandler,
		service.NewService,
		InitMenuChangeEventConsumer,
		wire.Struct(new(Module), "*"),
		wire.FieldsOf(new(*menu.Module), "Svc"),
		wire.FieldsOf(new(*role.Module), "Svc"),
		wire.FieldsOf(new(*policy.Module), "Svc"),
	)
	return new(Module), nil
}

func InitMenuChangeEventConsumer(q mq.MQ, svc service.Service) *event.MenuChangeEventConsumer {
	c, err := event.NewMenuChangeEventConsumer(q, svc)
	if err != nil {
		return nil
	}

	c.Start(context.Background())
	return c
}
