package strategy

import (
	"os"

	clientv1 "github.com/Duke1616/ecmdb/api/proto/gen/order/v1"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type UserTestSuite struct {
	suite.Suite
	db     *mongox.Mongo
	client clientv1.WorkOrderServiceClient
	ctrl   *gomock.Controller
}

func (s *UserTestSuite) SetupTestSuite() {
	// 加载配置
	s.loadConfig()
}

func (s *UserTestSuite) loadConfig() {
	dir, err := os.Getwd()
	s.Require().NoError(err)
	f, err := os.Open(dir + "/../../../../config/prod.yaml")
	s.Require().NoError(err)
	viper.SetConfigFile(f.Name())
	viper.WatchConfig()
	err = viper.ReadInConfig()
	s.Require().NoError(err)
}
