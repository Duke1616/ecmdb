package main

import (
	"github.com/Duke1616/ecmdb/ioc"
)

func main() {
	app, err := ioc.InitApp()
	if err != nil {
		panic(err)
	}

	err = app.Web.Run(":8001")
	panic(err)
}
