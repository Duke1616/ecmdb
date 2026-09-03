package resource

import (
	"context"
	"testing"

	"github.com/Duke1616/ecmdb/internal/domain"
	resourcemocks "github.com/Duke1616/ecmdb/internal/service/resource/mocks"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestFieldSecureAttrChangeConsumer_HandleEvent(t *testing.T) {
	testCases := []struct {
		name      string
		mock      func(ctrl *gomock.Controller) *resourcemocks.MockService
		evt       domain.FieldSecureAttrChange
		wantErr   bool
	}{
		{
			name: "成功同步并更新数据",
			mock: func(ctrl *gomock.Controller) *resourcemocks.MockService {
				svc := resourcemocks.NewMockService(ctrl)
				// 第一批查询返回 1 条数据
				svc.EXPECT().ListAndDecryptBeforeUtime(gomock.Any(), int64(1000), []string{"password"}, "host", int64(0), int64(10)).
					Return([]domain.Resource{
						{
							ID:       1,
							Name:     "server-01",
							ModelUID: "host",
							Data:     mongox.MapStr{"password": "secret"},
						},
					}, nil)

				// 执行批量更新
				svc.EXPECT().BatchUpdateResources(gomock.Any(), gomock.Any()).
					Return(int64(1), nil)

				return svc
			},
			evt: domain.FieldSecureAttrChange{
				ModelUid:   "host",
				FieldUid:   "password",
				Secure:     true,
				TiggerTime: 1000,
			},
			wantErr: false,
		},
		{
			name: "空数据直接退出",
			mock: func(ctrl *gomock.Controller) *resourcemocks.MockService {
				svc := resourcemocks.NewMockService(ctrl)
				svc.EXPECT().ListAndDecryptBeforeUtime(gomock.Any(), int64(1000), []string{"password"}, "host", int64(0), int64(10)).
					Return([]domain.Resource{}, nil)
				return svc
			},
			evt: domain.FieldSecureAttrChange{
				ModelUid:   "host",
				FieldUid:   "password",
				Secure:     true,
				TiggerTime: 1000,
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := tc.mock(ctrl)
			consumer := NewFieldSecureAttrChangeConsumer(nil, svc, 10)
			err := consumer.handleEvent(context.Background(), tc.evt)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
