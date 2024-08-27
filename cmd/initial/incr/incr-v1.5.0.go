package incr

import "github.com/Duke1616/ecmdb/cmd/initial/ioc"

type incrV150 struct {
	App *ioc.App
}

func NewIncrV150(app *ioc.App) InitialIncr {
	return &incrV150{
		App: app,
	}
}

func (i *incrV150) Version() string {
	return "v1.5.0"
}

func (i *incrV150) Commit() error {
	return nil
}

func (i *incrV150) Rollback() error {
	return nil
}

func (i *incrV150) Before() error {
	return nil
}

func (i *incrV150) After() error {
	return nil
}
