package web

import (
	"github.com/Duke1616/ecmdb/internal/menu/internal/domain"
	"github.com/ecodeclub/ekit/slice"
	"sort"
)

func GetMenusTree(ms []domain.Menu) []*Menu {
	// 将菜单转换为 *Menu 类型并存入 map
	allMap := make(map[int64]*Menu, len(ms))
	var list []*Menu

	for _, m := range ms {
		voMenu := toVoMenu(m)
		voMenu.Children = []*Menu{}
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
	sortMenu(list)
	return list
}

func sortMenu(menus []*Menu) {
	sort.Slice(menus, func(i, j int) bool {
		return menus[i].Sort < menus[j].Sort
	})

	for _, m := range menus {
		if len(m.Children) > 0 {
			sort.Slice(m.Children, func(i, j int) bool {
				return m.Children[i].Sort < m.Children[j].Sort
			})
			sortMenu(m.Children)
		}
	}
}

func toVoMenu(req domain.Menu) *Menu {
	return &Menu{
		Id:        req.Id,
		Pid:       req.Pid,
		Path:      req.Path,
		Sort:      req.Sort,
		Name:      req.Name,
		Redirect:  req.Redirect,
		Type:      req.Type.ToUint8(),
		Component: req.Component,
		Status:    req.Status.ToUint8(),
		Meta: Meta{
			Title:       req.Meta.Title,
			IsHidden:    req.Meta.IsHidden,
			IsAffix:     req.Meta.IsAffix,
			IsKeepAlive: req.Meta.IsKeepAlive,
			Icon:        req.Meta.Icon,
		},
		Endpoints: slice.Map(req.Endpoints, func(idx int, src domain.Endpoint) Endpoint {
			return Endpoint{
				Path:   src.Path,
				Method: src.Method,
				Desc:   src.Desc,
			}
		}),
		Children: []*Menu{},
	}
}
