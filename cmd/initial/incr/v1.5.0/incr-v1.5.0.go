package v150

import (
	"context"

	"github.com/Duke1616/ecmdb/cmd/initial/incr"
	"github.com/Duke1616/ecmdb/cmd/initial/ioc"
)

type incrV150 struct {
	App *ioc.App
}

func NewIncrV150(app *ioc.App) incr.InitialIncr {
	return &incrV150{
		App: app,
	}
}

func (i *incrV150) Version() string {
	return "v1.5.0"
}

func (i *incrV150) Commit(ctx context.Context) error {
	return nil
}

func (i *incrV150) Rollback(ctx context.Context) error {
	return nil
}

func (i *incrV150) Before(ctx context.Context) error {
	return nil
}

func (i *incrV150) After(ctx context.Context) error {
	return i.App.VerSvc.CreateOrUpdateVersion(context.Background(), i.Version())
}
