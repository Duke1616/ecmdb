package web

import (
	"fmt"
	"sort"

	"github.com/Duke1616/ecmdb/internal/menu"
	"github.com/Duke1616/ecmdb/internal/policy"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
)

func getPermission(menus []menu.Menu, pMap map[string]policy.Policy) ([]*Menu, []int64, error) {
	var (
		eg       errgroup.Group
		ms       []*Menu
		authzIds []int64
	)
	eg.Go(func() error {
		authzIds = getAuthzIds(menus, pMap)
		return nil
	})

	eg.Go(func() error {
		ms = GetMenusTree(menus)
		return nil
	})
	if err := eg.Wait(); err != nil {
		return ms, authzIds, err
	}

	return ms, authzIds, nil
}

func getAuthzIds(menus []menu.Menu, pMap map[string]policy.Policy) []int64 {
	// 筛选出有权限的节点
	return slice.FilterMap(menus, func(idx int, src menu.Menu) (int64, bool) {
		if len(src.Endpoints) == 0 {
			return 0, false
		}

		for _, endpoint := range src.Endpoints {
			key := fmt.Sprintf("%s:%s", endpoint.Path, endpoint.Method)

			if _, ok := pMap[key]; !ok {
				return 0, false
			}
		}

		return src.Id, true
	})
}

func GetPermissionMenusTree(ms []menu.Menu, ps []policy.Policy) (list []*Menu, err error) {
	// 把所有接口权限生成 map 接口
	pMap := slice.ToMap(ps, func(element policy.Policy) string {
		return fmt.Sprintf("%s:%s", element.Path, element.Method)
	})

	// 过滤拥有权限的路由
	menus := slice.FilterMap(ms, func(idx int, src menu.Menu) (*Menu, bool) {
		if src.Pid == 0 {
			return toVoMenu(src), true
		}

		ok := filterEndpoints(src.Endpoints, pMap)
		if ok {
			return toVoMenu(src), true
		}

		return nil, false
	})

	allMap := map[int64]*Menu{}
	list = []*Menu{}
	for k, cat := range menus {
		menus[k].Children = []*Menu{}
		allMap[cat.Id] = menus[k]
		if cat.Pid == 0 {
			list = append(list, menus[k])
		}
	}

	for k, cat := range menus {
		_, ok := allMap[cat.Pid]
		if ok {
			//如果父级别数据存在，添加到Children
			//利用指针逻辑，map中数据和列表中原始对象为统一指针。指向同一内存地址，如此对map中数据操作，也相当于对原始数据操作。
			allMap[cat.Pid].Children = append(allMap[cat.Pid].Children, menus[k])
		}
	}

	sortMenu(list)
	return
}

func filterEndpoints(endpoints []menu.Endpoint, pMap map[string]policy.Policy) bool {
	// TODO 如果节点为空, 应该如何处理，目前是当作没有权限
	if endpoints == nil || len(endpoints) == 0 {
		return false
	}

	var filtered []menu.Endpoint
	for _, ep := range endpoints {
		key := fmt.Sprintf("%s:%s", ep.Path, ep.Method)
		if _, exists := pMap[key]; exists {
			filtered = append(filtered, ep)
		}
	}

	if len(filtered) == len(endpoints) {
		return true
	}

	return false
}

// GetMenusTreeByButton 带按钮权限的菜单树构建函数
func GetMenusTreeByButton(ms []menu.Menu) []*Menu {
	return buildMenuTree(ms, handleButtonPermission)
}

// GetMenusTree 不带按钮权限的菜单树构建函数
func GetMenusTree(ms []menu.Menu) []*Menu {
	return buildMenuTree(ms, handleNoButtonPermission)
}

func buildMenuTree(ms []menu.Menu, handler func(m *menu.Menu, parent *Menu, allMap map[int64]*Menu)) []*Menu {
	allMap := make(map[int64]*Menu, len(ms))
	var list []*Menu

	// 第一次遍历，初始化所有菜单项并存入 map
	for _, m := range ms {
		voMenu := toVoMenu(m)
		voMenu.Children = []*Menu{}
		allMap[m.Id] = voMenu
		if m.Pid == 0 {
			list = append(list, voMenu)
		}
	}

	// 第二次遍历，根据处理逻辑构建树结构
	for _, m := range ms {
		if parent, exists := allMap[m.Pid]; exists {
			handler(&m, parent, allMap)
		}
	}

	// 对菜单树进行排序
	sortMenu(list)
	return list
}

// 处理按钮权限的处理器
func handleButtonPermission(m *menu.Menu, parent *Menu, allMap map[int64]*Menu) {
	if m.Type == 3 {
		parent.Meta.Buttons = append(parent.Meta.Buttons, m.Name)
	} else {
		parent.Children = append(parent.Children, allMap[m.Id])
	}
}

// 不处理按钮权限的处理器
func handleNoButtonPermission(m *menu.Menu, parent *Menu, allMap map[int64]*Menu) {
	parent.Children = append(parent.Children, allMap[m.Id])
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

func toVoMenu(req menu.Menu) *Menu {
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
			Buttons:     []string{},
		},
		Endpoints: slice.Map(req.Endpoints, func(idx int, src menu.Endpoint) Endpoint {
			return Endpoint{
				Path:   src.Path,
				Method: src.Method,
				Desc:   src.Desc,
			}
		}),
		Children: []*Menu{},
	}
}
