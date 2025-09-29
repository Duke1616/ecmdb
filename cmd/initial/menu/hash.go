package menu

import (
	"crypto/md5"
	"encoding/json"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/menu"
	"github.com/gotomicro/ego/core/elog"
)

// HashCalculator 菜单哈希计算器
type HashCalculator struct {
	logger elog.Component
}

// NewMenuHashCalculator 创建菜单哈希计算器
func NewMenuHashCalculator() *HashCalculator {
	return &HashCalculator{
		logger: *elog.DefaultLogger,
	}
}

// CalculateMenuHash 计算菜单数据的 MD5 哈希值
// 基于菜单内容而不是文件内容，确保在编译后也能正常工作
func (m *HashCalculator) CalculateMenuHash() (string, error) {
	// 获取菜单数据
	menus := GetInjectMenus()

	// 计算菜单数据的哈希值
	return m.CalculateMenuDataHash(menus)
}

// CalculateMenuDataHash 计算菜单数据的哈希值
// 基于菜单内容而不是文件内容
func (m *HashCalculator) CalculateMenuDataHash(menus []menu.Menu) (string, error) {
	// 将菜单数据序列化为字符串
	data, err := m.serializeMenuData(menus)
	if err != nil {
		return "", fmt.Errorf("序列化菜单数据失败: %w", err)
	}

	// 计算 MD5
	hash := md5.Sum(data)
	return fmt.Sprintf("%x", hash), nil
}

// serializeMenuData 序列化菜单数据
func (m *HashCalculator) serializeMenuData(menus []menu.Menu) ([]byte, error) {
	// 使用 JSON 序列化确保稳定性和一致性
	data, err := json.Marshal(menus)
	if err != nil {
		return []byte{}, fmt.Errorf("序列化菜单数据失败: %w", err)
	}
	return data, nil
}

// CalculateProjectMenuHash 计算项目菜单的整体哈希值
// 基于菜单数据内容，确保在编译后也能正常工作
func (m *HashCalculator) CalculateProjectMenuHash() (string, error) {
	return m.CalculateMenuHash()
}
