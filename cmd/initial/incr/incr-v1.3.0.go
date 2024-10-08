package incr

import (
	"context"
	"github.com/Duke1616/ecmdb/cmd/initial/ioc"
)

type incrV130 struct {
	App *ioc.App
}

func NewIncrV130(app *ioc.App) InitialIncr {
	return &incrV130{
		App: app,
	}
}

func (i *incrV130) Version() string {
	return "v1.3.0"
}

func (i *incrV130) Commit() error {
	return nil
}

func (i *incrV130) Rollback() error {
	return nil
}

func (i *incrV130) After() error {
	return nil
}

func (i *incrV130) Before() error {
	return i.App.VerSvc.CreateOrUpdateVersion(context.Background(), i.Version())
}
