//go:build wireinject

package startup

import (
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/google/wire"
)

func InitRMHandler() (*relation.RMHandler, error) {
	wire.Build(InitMongoDB,
		relation.InitModule,
		wire.FieldsOf(new(*relation.Module), "RMHdl"),
	)
	return new(relation.RMHandler), nil
}

func InitRRHandler() (*relation.RRHandler, error) {
	wire.Build(InitMongoDB,
		relation.InitModule,
		wire.FieldsOf(new(*relation.Module), "RRHdl"),
	)
	return new(relation.RRHandler), nil
}

func InitRRSvc() relation.RRSvc {
	wire.Build(InitMongoDB, relation.InitRRService)
	return nil
}

func InitRMSvc() relation.RMSvc {
	wire.Build(InitMongoDB, relation.InitRMService)
	return nil
}
