package menu

import (
	"encoding/json"
	"testing"

	"github.com/Duke1616/ecmdb/internal/menu"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashCalculator_CalculateMenuHash(t *testing.T) {
	calculator := NewMenuHashCalculator()

	testCases := []struct {
		name     string
		validate func(t *testing.T, hash string, err error)
	}{
		{
			name: "计算菜单哈希成功",
			validate: func(t *testing.T, hash string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, hash)
				assert.Len(t, hash, 32) // MD5 哈希长度应该是 32
			},
		},
		{
			name: "多次调用结果一致",
			validate: func(t *testing.T, hash string, err error) {
				require.NoError(t, err)
				hash2, err2 := calculator.CalculateMenuHash()
				require.NoError(t, err2)
				assert.Equal(t, hash, hash2)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := calculator.CalculateMenuHash()
			tc.validate(t, hash, err)
		})
	}
}

func TestHashCalculator_CalculateMenuDataHash(t *testing.T) {
	calculator := NewMenuHashCalculator()

	testCases := []struct {
		name     string
		menus    []menu.Menu
		validate func(t *testing.T, hash string, err error)
	}{
		{
			name:  "空菜单列表",
			menus: []menu.Menu{},
			validate: func(t *testing.T, hash string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, hash)
				assert.Len(t, hash, 32)
			},
		},
		{
			name:  "默认菜单数据",
			menus: GetInjectMenus(),
			validate: func(t *testing.T, hash string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, hash)
				assert.Len(t, hash, 32)
			},
		},
		{
			name:  "相同数据产生相同哈希",
			menus: GetInjectMenus(),
			validate: func(t *testing.T, hash string, err error) {
				require.NoError(t, err)
				hash2, err2 := calculator.CalculateMenuDataHash(GetInjectMenus())
				require.NoError(t, err2)
				assert.Equal(t, hash, hash2)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := calculator.CalculateMenuDataHash(tc.menus)
			tc.validate(t, hash, err)
		})
	}
}

func TestHashCalculator_CalculateProjectMenuHash(t *testing.T) {
	calculator := NewMenuHashCalculator()

	testCases := []struct {
		name     string
		validate func(t *testing.T, hash string, err error)
	}{
		{
			name: "计算项目菜单哈希成功",
			validate: func(t *testing.T, hash string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, hash)
				assert.Len(t, hash, 32)
			},
		},
		{
			name: "与CalculateMenuHash结果一致",
			validate: func(t *testing.T, hash string, err error) {
				require.NoError(t, err)
				menuHash, err2 := calculator.CalculateMenuHash()
				require.NoError(t, err2)
				assert.Equal(t, hash, menuHash)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := calculator.CalculateProjectMenuHash()
			tc.validate(t, hash, err)
		})
	}
}

func TestHashCalculator_serializeMenuData(t *testing.T) {
	calculator := NewMenuHashCalculator()

	testCases := []struct {
		name     string
		menus    []menu.Menu
		validate func(t *testing.T, data []byte, err error)
	}{
		{
			name:  "空菜单序列化",
			menus: []menu.Menu{},
			validate: func(t *testing.T, data []byte, err error) {
				require.NoError(t, err)
				assert.Equal(t, "[]", string(data))
			},
		},
		{
			name:  "非空菜单序列化",
			menus: GetInjectMenus(),
			validate: func(t *testing.T, data []byte, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, data)
			},
		},
		{
			name:  "序列化结果可以反序列化",
			menus: GetInjectMenus(),
			validate: func(t *testing.T, data []byte, err error) {
				require.NoError(t, err)
				var deserializedMenus []menu.Menu
				err = json.Unmarshal(data, &deserializedMenus)
				require.NoError(t, err)
				assert.Len(t, deserializedMenus, len(GetInjectMenus()))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := calculator.serializeMenuData(tc.menus)
			tc.validate(t, data, err)
		})
	}
}

func TestHashCalculator_Consistency(t *testing.T) {
	calculator := NewMenuHashCalculator()

	testCases := []struct {
		name     string
		validate func(t *testing.T)
	}{
		{
			name: "不同方法计算相同数据的一致性",
			validate: func(t *testing.T) {
				menus := GetInjectMenus()

				hash1, err1 := calculator.CalculateMenuDataHash(menus)
				require.NoError(t, err1)

				hash2, err2 := calculator.CalculateMenuHash()
				require.NoError(t, err2)

				hash3, err3 := calculator.CalculateProjectMenuHash()
				require.NoError(t, err3)

				assert.Equal(t, hash1, hash2)
				assert.Equal(t, hash2, hash3)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.validate(t)
		})
	}
}
