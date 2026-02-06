//go:build wireinject

package ioc

import (
	notificationv1 "github.com/Duke1616/ecmdb/api/proto/gen/notification/v1"
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/codebook"
	"github.com/Duke1616/ecmdb/internal/dataio"
	"github.com/Duke1616/ecmdb/internal/department"
	"github.com/Duke1616/ecmdb/internal/discovery"
	"github.com/Duke1616/ecmdb/internal/endpoint"
	"github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/event"
	"github.com/Duke1616/ecmdb/internal/menu"
	"github.com/Duke1616/ecmdb/internal/model"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/permission"
	"github.com/Duke1616/ecmdb/internal/pkg/middleware"
	"github.com/Duke1616/ecmdb/internal/policy"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/Duke1616/ecmdb/internal/role"
	"github.com/Duke1616/ecmdb/internal/rota"
	"github.com/Duke1616/ecmdb/internal/runner"
	"github.com/Duke1616/ecmdb/internal/strategy"
	"github.com/Duke1616/ecmdb/internal/task"
	"github.com/Duke1616/ecmdb/internal/template"
	"github.com/Duke1616/ecmdb/internal/terminal"
	"github.com/Duke1616/ecmdb/internal/tools"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/worker"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/pkg/storage"
	grpcpkg "github.com/Duke1616/ework-runner/pkg/grpc"
	"github.com/Duke1616/ework-runner/pkg/grpc/registry"
	"github.com/google/wire"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

var BaseSet = wire.NewSet(InitMongoDB, InitMySQLDB, InitRedis, InitMinioClient, InitMQ,
	InitRedisSearch, InitEtcdClient, InitWorkWx, InitLarkClient, InitModuleCrypto, InitRegistry)

func InitApp() (*App, error) {
	wire.Build(wire.Struct(new(App), "*"),
		BaseSet,
		InitSession,
		InitCasbin,
		InitLdapConfig,
		storage.NewS3Storage,
		InitEALERTGrpcClient,
		InitNotificationServiceClient,
		model.InitModule,
		wire.FieldsOf(new(*model.Module), "Hdl"),
		attribute.InitModule,
		wire.FieldsOf(new(*attribute.Module), "Hdl"),
		resource.InitModule,
		wire.FieldsOf(new(*resource.Module), "Hdl"),
		relation.InitModule,
		wire.FieldsOf(new(*relation.Module), "RRHdl", "RMHdl", "RTHdl"),
		user.InitModule,
		wire.FieldsOf(new(*user.Module), "Hdl", "RpcServer"),
		template.InitModule,
		wire.FieldsOf(new(*template.Module), "Hdl", "GroupHdl"),
		codebook.InitModule,
		wire.FieldsOf(new(*codebook.Module), "Hdl"),
		worker.InitModule,
		wire.FieldsOf(new(*worker.Module), "Hdl"),
		runner.InitModule,
		wire.FieldsOf(new(*runner.Module), "Hdl"),
		order.InitModule,
		wire.FieldsOf(new(*order.Module), "Hdl", "RpcServer"),
		strategy.InitModule,
		wire.FieldsOf(new(*strategy.Module), "Hdl"),
		workflow.InitModule,
		wire.FieldsOf(new(*workflow.Module), "Hdl"),
		engine.InitModule,
		wire.FieldsOf(new(*engine.Module), "Hdl"),
		event.InitModule,
		wire.FieldsOf(new(*event.Module), "Event"),
		task.InitModule,
		wire.FieldsOf(new(*task.Module), "Hdl", "StartTaskJob", "PassProcessTaskJob"),
		policy.InitModule,
		wire.FieldsOf(new(*policy.Module), "Hdl", "Svc", "RpcServer"),
		menu.InitModule,
		wire.FieldsOf(new(*menu.Module), "Hdl"),
		endpoint.InitModule,
		wire.FieldsOf(new(*endpoint.Module), "Hdl", "Svc", "RpcServer"),
		department.InitModule,
		wire.FieldsOf(new(*department.Module), "Hdl"),
		role.InitModule,
		wire.FieldsOf(new(*role.Module), "Hdl"),
		permission.InitModule,
		wire.FieldsOf(new(*permission.Module), "Hdl"),
		rota.InitModule,
		wire.FieldsOf(new(*rota.Module), "Hdl", "RpcServer"),
		discovery.InitModule,
		wire.FieldsOf(new(*discovery.Module), "Hdl"),
		tools.InitModule,
		terminal.InitModule,
		dataio.InitModule,
		wire.FieldsOf(new(*dataio.Module), "Hdl"),
		middleware.NewCheckPolicyMiddlewareBuilder,
		middleware.NewCheckLoginMiddlewareBuilder,
		initCronJobs,
		InitWebServer,
		InitGrpcServer,
		InitGinMiddlewares)
	return new(App), nil
}

// InitEALERTGrpcClient 初始化 EALERT gRPC 客户端
func InitEALERTGrpcClient(reg registry.Registry) grpc.ClientConnInterface {
	var cfg grpcpkg.ClientConfig
	if err := viper.UnmarshalKey("grpc.client.ealert", &cfg); err != nil {
		panic(err)
	}

	cc, err := grpcpkg.NewClientConn(
		reg,
		grpcpkg.WithServiceName(cfg.Name),
		grpcpkg.WithClientJWTAuth(cfg.AuthToken),
	)
	if err != nil {
		panic(err)
	}

	return cc
}

// InitNotificationServiceClient 初始化 notification 服务客户端
func InitNotificationServiceClient(cc grpc.ClientConnInterface) notificationv1.NotificationServiceClient {
	return notificationv1.NewNotificationServiceClient(cc)
}
