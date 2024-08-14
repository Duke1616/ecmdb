package web

import (
	"fmt"
	"github.com/Duke1616/ecmdb/internal/menu"
	"github.com/Duke1616/ecmdb/internal/policy"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
	"sort"
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
		var err error
		ms, err = GetMenusTree(menus)
		return err
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

func GetMenusTree(ms []menu.Menu) (list []*Menu, err error) {
	menus := slice.Map(ms, func(idx int, src menu.Menu) *Menu {
		return toVoMenu(src)
	})

	//生成map， 方便查找获取对象
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
			allMap[cat.Pid].Children = append(allMap[cat.Pid].Children, menus[k])
		}
	}

	sortMenu(list)
	return
}

func sortMenu(menus []*Menu) {
	sort.Slice(menus, func(i, j int) bool {
		return menus[i].Sort < menus[j].Sort
	})
	for _, m := range menus {
		if len(m.Children) > 0 {
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
