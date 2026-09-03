package service

import (
	"context"
	"testing"

	"github.com/Duke1616/ecmdb/internal/domain"
	attributemocks "github.com/Duke1616/ecmdb/internal/service/attribute/mocks"
	resourcemocks "github.com/Duke1616/ecmdb/internal/service/resource/mocks"
	"github.com/Duke1616/ecmdb/pkg/excelx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_ToExcelColumns(t *testing.T) {
	testCases := []struct {
		name     string
		input    []domain.Attribute
		assertFn func(t *testing.T, cols []excelx.Column)
	}{
		{
			name: "领域属性转换为Excel声明式列模型",
			input: []domain.Attribute{
				{
					FieldUid:  "name",
					FieldName: "资产名称",
					FieldType: "string",
					Required:  true,
					Secure:    false,
				},
				{
					FieldUid:  "env",
					FieldName: "部署环境",
					FieldType: "select",
					Required:  false,
					Option:    []string{"dev", "prod"},
				},
			},
			assertFn: func(t *testing.T, cols []excelx.Column) {
				require.Len(t, cols, 2)
				assert.Equal(t, "name", cols[0].Key)
				assert.Equal(t, "资产名称", cols[0].Title)
				assert.True(t, cols[0].Unique)
				assert.True(t, cols[0].Required)

				assert.Equal(t, "env", cols[1].Key)
				assert.Equal(t, []string{"dev", "prod"}, cols[1].Options)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cols := toExcelColumns(tc.input)
			tc.assertFn(t, cols)
		})
	}
}

func Test_SortAttributesByPriority(t *testing.T) {
	testCases := []struct {
		name     string
		input    []domain.Attribute
		expected []string
	}{
		{
			name: "name字段自动置顶且其余字段依SortKey及Index自然排序",
			input: []domain.Attribute{
				{FieldUid: "ip", SortKey: 2000, Index: 2},
				{FieldUid: "desc", SortKey: 3000, Index: 3},
				{FieldUid: "name", SortKey: 9999, Index: 9},
				{FieldUid: "cpu", SortKey: 1000, Index: 1},
			},
			expected: []string{"name", "cpu", "ip", "desc"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			sorted := sortAttributesByPriority(tc.input)
			actual := make([]string, len(sorted))
			for i, attr := range sorted {
				actual[i] = attr.FieldUid
			}
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func Test_DataIO_ExportTemplateAndImportFlow(t *testing.T) {
	testCases := []struct {
		name     string
		modelUID string
		attrs    []domain.Attribute
	}{
		{
			name:     "导出模板并回填导入完整闭环验证",
			modelUID: "host",
			attrs: []domain.Attribute{
				{FieldUid: "name", FieldName: "主机名称", FieldType: "string", Required: true},
				{FieldUid: "cpu", FieldName: "CPU核心", FieldType: "int", Required: false},
				{FieldUid: "os", FieldName: "系统类型", FieldType: "select", Option: []string{"Linux", "Windows"}},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			attrMock := attributemocks.NewMockService(ctrl)
			resMock := resourcemocks.NewMockService(ctrl)

			attrMock.EXPECT().ListAttributes(gomock.Any(), tc.modelUID).
				Return(tc.attrs, int64(len(tc.attrs)), nil).AnyTimes()

			svc := &dataIOService{
				attrSvc: attrMock,
				resSvc:  resMock,
			}

			// 使用 excelx 引擎生成真实 Excel 字节流
			columns := toExcelColumns(tc.attrs)
			excelBytes, err := excelx.NewWriter("主机", columns).WriteData([]map[string]interface{}{
				{
					"name": "web-node-01",
					"cpu":  16,
					"os":   "Linux",
				},
			})
			require.NoError(t, err)
			assert.NotEmpty(t, excelBytes)

			// 验证导入读取并分批写入
			resMock.EXPECT().BatchCreateOrUpdate(gomock.Any(), gomock.Len(1)).
				DoAndReturn(func(_ context.Context, resources []domain.Resource) error {
					assert.Equal(t, "web-node-01", resources[0].Data["name"])
					assert.Equal(t, int64(16), resources[0].Data["cpu"])
					assert.Equal(t, "Linux", resources[0].Data["os"])
					return nil
				})

			count, err := svc.Import(context.Background(), tc.modelUID, excelBytes)
			require.NoError(t, err)
			assert.Equal(t, 1, count)
		})
	}
}
