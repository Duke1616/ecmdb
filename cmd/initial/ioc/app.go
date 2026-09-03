package ioc

import (
	"github.com/Duke1616/ecmdb/cmd/initial/version"
	"github.com/Duke1616/ecmdb/pkg/mongox"
)

type App struct {
	VerSvc version.Service
	DB     *mongox.Mongo
}

