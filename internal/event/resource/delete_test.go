package resource

import (
	"context"
	"testing"

	"github.com/Duke1616/ecmdb/internal/domain"
	resourcemocks "github.com/Duke1616/ecmdb/internal/service/resource/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestFieldDeleteConsumer_Process(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) *resourcemocks.MockService
		evt     domain.FieldDelete
		wantErr bool
	}{
		{
			name: "成功级联清理字段",
			mock: func(ctrl *gomock.Controller) *resourcemocks.MockService {
				svc := resourcemocks.NewMockService(ctrl)
				svc.EXPECT().UnsetCustomField(gomock.Any(), "host", "legacy_field").
					Return(int64(5), nil)
				return svc
			},
			evt: domain.FieldDelete{
				ModelUid:    "host",
				FieldUid:    "legacy_field",
				TriggerTime: 123456,
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := tc.mock(ctrl)
			consumer := NewFieldDeleteConsumer(nil, svc)
			err := consumer.Process(context.Background(), tc.evt)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
