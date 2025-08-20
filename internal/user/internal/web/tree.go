package web

import (
	"sort"

	"github.com/Duke1616/ecmdb/internal/department"
	"github.com/Duke1616/ecmdb/internal/user/internal/domain"
	"github.com/ecodeclub/ekit/slice"
)

func GenerateDepartmentUserTree(pipeline []domain.UserCombination, departments []department.Department) ([]*UserDepartmentCombination, error) {
	// Step 1: 生成部门到用户的映射
	uMap := slice.ToMapV(pipeline, func(element domain.UserCombination) (int64, []*UserDepartmentCombination) {
		return element.DepartMentId, slice.Map(element.Users, func(idx int, src domain.User) *UserDepartmentCombination {
			return &UserDepartmentCombination{
				Id:          src.Id,
				Disabled:    false,
				DisplayName: src.DisplayName,
				Type:        "person",
				Name:        src.Username,
				Children:    []*UserDepartmentCombination{},
			}
		})
	})

	// Step 2: 初始化部门并构建树
	allMap := make(map[int64]*UserDepartmentCombination, len(departments))
	var tree []*UserDepartmentCombination

	for _, d := range departments {
		depNode := toVoMenu(d)
		depNode.Children = []*UserDepartmentCombination{}

		// 如果该部门有用户，加入用户列表
		if users, exists := uMap[d.Id]; exists {
			depNode.Children = append(depNode.Children, users...)
		}

		allMap[d.Id] = depNode

		// 如果部门的父级ID为0，说明是顶级部门，加入树的根节点
		if d.Pid == 0 {
			tree = append(tree, depNode)
		}
	}

	// Step 3: 构建部门树，将子部门加入对应的父部门中
	for _, d := range departments {
		if parent, exists := allMap[d.Pid]; exists {
			parent.Children = append(parent.Children, allMap[d.Id])
		}
	}

	// Step 4: 对生成的树进行排序
	sortDepartment(tree)

	return tree, nil
}

// 将 department.Department 转换为 UserDepartmentCombination
func toVoMenu(d department.Department) *UserDepartmentCombination {
	return &UserDepartmentCombination{
		Id:          d.Id,
		Disabled:    true,
		DisplayName: d.Name,
		Type:        "department",
		Name:        d.Name,
		Sort:        d.Sort,
		Children:    []*UserDepartmentCombination{},
	}
}

// 对部门和用户进行排序
func sortDepartment(deps []*UserDepartmentCombination) {
	sort.Slice(deps, func(i, j int) bool {
		return deps[i].Sort < deps[j].Sort
	})

	for _, dep := range deps {
		if len(dep.Children) > 0 {
			sortDepartment(dep.Children)
		}
	}
}
