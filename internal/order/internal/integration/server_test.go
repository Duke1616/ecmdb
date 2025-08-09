package integration

import (
	"context"
	clientv1 "github.com/Duke1616/ecmdb/api/proto/gen/order/v1"
	evtmocks "github.com/Duke1616/ecmdb/internal/order/internal/event/mocks"
	grpc2 "github.com/Duke1616/ecmdb/internal/order/internal/grpc"
	"github.com/Duke1616/ecmdb/internal/order/internal/integration/startup"
	"github.com/Duke1616/ecmdb/internal/order/internal/repository"
	"github.com/Duke1616/ecmdb/internal/order/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/order/internal/service"
	"github.com/Duke1616/ecmdb/pkg/grpcx"
	jwtpkg "github.com/Duke1616/ecmdb/pkg/grpcx/interceptors/jwt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/metadata"
	"log"
	"os"
	"testing"
	"time"
)

type GRPCServerTestSuite struct {
	BaseGRPCServerTestSuite
}

func (s *GRPCServerTestSuite) SetupSuite() {
	s.BaseGRPCServerTestSuite.SetupTestSuite()
}

func TestGRPCServerWithSuccessMock(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(GRPCServerTestSuite))
}

func (s *GRPCServerTestSuite) TearDownSuite() {
	s.BaseGRPCServerTestSuite.TearDownTestSuite()
}

func (s *GRPCServerTestSuite) TestCreateWorkOrder() {
	testCases := []struct {
		name          string
		before        func(t *testing.T)
		req           *clientv1.CreateOrderRequest
		setupContext  func(context.Context) context.Context
		wantResp      *clientv1.Response
		errAssertFunc assert.ErrorAssertionFunc
	}{
		{
			name: "参数传递错误",
			req: &clientv1.CreateOrderRequest{
				Order: &clientv1.Order{
					TemplateName: "",
				},
			},
			setupContext:  s.contextWithJWT,
			errAssertFunc: assert.Error,
		},
		{
			name: "创建工单成功",
			req: &clientv1.CreateOrderRequest{
				Order: &clientv1.Order{
					TemplateName: "",
				},
			},
			before: func(t *testing.T) {
				s.producer.EXPECT().Produce(gomock.Any(), gomock.Any()).Return(nil)
			},
			setupContext:  s.contextWithJWT,
			errAssertFunc: assert.Error,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)

			// 添加JWT认证或其他上下文设置
			ctx := tc.setupContext(t.Context())

			// 调用服务
			log.Printf("发送请求: %+v", tc.req)
			resp, err := s.client.CreateWorkOrder(ctx, tc.req)

			if err != nil {
				log.Printf("发送请求失败: %v", err)
			} else {
				log.Printf("接收响应: %+v", resp)
			}

			tc.errAssertFunc(t, err)
			if err != nil {
				return
			}
		})
	}
}

type BaseGRPCServerTestSuite struct {
	suite.Suite
	db               *mongox.Mongo
	clientGRPCServer *grpcx.Server
	client           clientv1.WorkOrderServiceClient
	producer         *evtmocks.MockCreateProcessEventProducer
	ctrl             *gomock.Controller
}

func (s *BaseGRPCServerTestSuite) SetupTestSuite() {
	// 加载配置
	dir, err := os.Getwd()
	s.Require().NoError(err)
	f, err := os.Open(dir + "/../../../../config/prod.yaml")
	s.Require().NoError(err)
	viper.SetConfigFile(f.Name())
	viper.WatchConfig()
	err = viper.ReadInConfig()
	s.Require().NoError(err)

	// 初始化数据库
	s.db = startup.InitMongoDB()
	time.Sleep(1 * time.Second)

	// 初始化 mock
	s.ctrl = gomock.NewController(s.T())
	s.producer = evtmocks.NewMockCreateProcessEventProducer(s.ctrl)

	// 初始化注册中心
	etcdClient := startup.InitEtcdClient()

	orderDao := dao.NewOrderDAO(s.db)
	orderRepository := repository.NewOrderRepository(orderDao)
	orderServer := grpc2.NewWorkOrderServer(service.NewService(orderRepository, s.producer))
	s.clientGRPCServer = startup.InitGrpcServer(orderServer, etcdClient)

	// 创建服务器
	setupCtx, setupCancelFunc := context.WithCancel(context.Background())
	go func() {
		setupCancelFunc()
		err = s.clientGRPCServer.Serve()
		s.NoError(err)
	}()
	// 等待服务启动
	log.Printf("等待服务启动...\n")
	select {
	case <-setupCtx.Done():
		time.Sleep(1 * time.Second)
	case <-time.After(10 * time.Second):
		s.Fail("服务启动超时")
	}

	// 创建客户端
	conn := startup.InitEcmdbClient(etcdClient)
	s.client = clientv1.NewWorkOrderServiceClient(conn)
}

func (s *BaseGRPCServerTestSuite) TearDownTestSuite() {
	s.ctrl.Finish()
	s.clientGRPCServer.Stop()
}

func (s *BaseGRPCServerTestSuite) TearDownTest() {
}

// 添加JWT认证到context
func (s *BaseGRPCServerTestSuite) contextWithJWT(ctx context.Context) context.Context {
	// 使用项目已有的JWT包创建令牌
	jwtAuth := jwtpkg.NewJwtAuth("1234567890")

	// 创建包含业务ID的声明
	claims := jwt.MapClaims{
		"biz_id": float64(1),
	}

	// 使用JWT认证包的Encode方法生成令牌
	tokenString, _ := jwtAuth.Encode(claims)

	// 创建带有授权信息的元数据
	md := metadata.New(map[string]string{
		"Authorization": "Bearer " + tokenString,
	})
	return metadata.NewOutgoingContext(ctx, md)
}
