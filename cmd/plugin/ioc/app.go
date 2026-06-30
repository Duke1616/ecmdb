package ioc

import (
	pluginSvc "github.com/Duke1616/ecmdb/internal/service/plugin"
)

type App struct {
	PluginSvc      pluginSvc.Service
	TenantProvider pluginSvc.BuiltinTenantProvider
}
