package web

import (
	"sort"

	"github.com/Duke1616/ecmdb/internal/department/internal/domain"
)

func GetDepartmentsTree(ms []domain.Department) []*Department {
	// 将菜单转换为 *Menu 类型并存入 map
	allMap := make(map[int64]*Department, len(ms))
	var list []*Department

	for _, m := range ms {
		voMenu := toVoMenu(m)
		voMenu.Children = []*Department{}
		allMap[m.Id] = voMenu
		if m.Pid == 0 {
			list = append(list, voMenu)
		}
	}

	// 构建菜单树
	for _, m := range ms {
		if parent, exists := allMap[m.Pid]; exists {
			parent.Children = append(parent.Children, allMap[m.Id])
		}
	}

	// 对菜单树进行排序
	sortDepartment(list)
	return list
}

func sortDepartment(menus []*Department) {
	sort.Slice(menus, func(i, j int) bool {
		return menus[i].Sort < menus[j].Sort
	})

	for _, m := range menus {
		if len(m.Children) > 0 {
			sort.Slice(m.Children, func(i, j int) bool {
				return m.Children[i].Sort < m.Children[j].Sort
			})
			sortDepartment(m.Children)
		}
	}
}

func toVoMenu(req domain.Department) *Department {
	return &Department{
		Id:         req.Id,
		Pid:        req.Pid,
		Sort:       req.Sort,
		Name:       req.Name,
		Enabled:    req.Enabled,
		Leaders:    req.Leaders,
		MainLeader: req.MainLeader,
		Children:   []*Department{},
	}
}
