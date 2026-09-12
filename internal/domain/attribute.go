package domain

import (
	"cmp"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/samber/lo"
)


type Attribute struct {
	ID        int64
	GroupId   int64
	ModelUid  string
	FieldUid  string
	FieldName string
	FieldType string
	Required  bool
	Display   bool
	Secure    bool
	Link      bool
	Index     int64
	SortKey   int64 // 拖拽排序键（稀疏索引）
	Option    interface{}
	Version   int64
	Builtin   bool
}

func (a Attribute) ValidateForCreate() error {
	var problems []string
	if a.GroupId <= 0 {
		problems = append(problems, "group_id 不能为空")
	}
	if strings.TrimSpace(a.ModelUid) == "" {
		problems = append(problems, "model_uid 不能为空")
	}
	if strings.TrimSpace(a.FieldUid) == "" {
		problems = append(problems, "field_uid 不能为空")
	}
	if strings.TrimSpace(a.FieldName) == "" {
		problems = append(problems, "field_name 不能为空")
	}
	if strings.TrimSpace(a.FieldType) == "" {
		problems = append(problems, "field_type 不能为空")
	}
	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("%s", strings.Join(problems, "; "))
}

// GetID 实现 Sortable 接口
func (a Attribute) GetID() int64 { return a.ID }

// GetSortKey 实现 Sortable 接口
func (a Attribute) GetSortKey() int64 { return a.SortKey }

// GetOptionStrings 获取选项字符串列表
// NOTE: 仅当字段为 select 或 list 时才解析选项，统一处理 []string、[]interface{}、primitive.A 等类型
func (a *Attribute) GetOptionStrings() []string {
	if a.FieldType != "select" && a.FieldType != "list" {
		return nil
	}
	if a.Option == nil {
		return nil
	}

	switch opts := a.Option.(type) {
	case []string:
		var result []string
		for _, opt := range opts {
			trimmed := strings.TrimSpace(opt)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	case []interface{}:
		result := make([]string, 0, len(opts))
		for _, opt := range opts {
			str := strings.TrimSpace(fmt.Sprint(opt))
			if str != "" && str != "<nil>" {
				result = append(result, str)
			}
		}
		return result
	default:
		// 处理 MongoDB primitive.A 等 slice 类型
		v := reflect.ValueOf(a.Option)
		if v.Kind() == reflect.Slice {
			result := make([]string, 0, v.Len())
			for i := 0; i < v.Len(); i++ {
				str := strings.TrimSpace(fmt.Sprint(v.Index(i).Interface()))
				if str != "" && str != "<nil>" {
					result = append(result, str)
				}
			}
			return result
		}
		return nil
	}
}

type AttributeGroup struct {
	ID       int64
	Name     string
	ModelUid string
	SortKey  int64
}

// GetID 实现 Sortable 接口
func (ag AttributeGroup) GetID() int64 { return ag.ID }

// GetSortKey 实现 Sortable 接口
func (ag AttributeGroup) GetSortKey() int64 { return ag.SortKey }

type AttributePipeline struct {
	GroupId    int64       `bson:"_id"`
	Total      int         `bson:"total"`
	Attributes []Attribute `bson:"attributes"`
}

// AttributeSortItem 属性排序更新项
// NOTE: 用于批量更新时的数据传输
type AttributeSortItem struct {
	ID      int64
	GroupId int64
	SortKey int64
}

// AttributeGroupSortItem 属性组排序更新项
// NOTE: 用于批量更新时的数据传输
type AttributeGroupSortItem struct {
	ID      int64
	SortKey int64
}

const defaultDisplayLimit = 6

// ResolveDisplayFields 解析模型的默认展示列
// 策略 1：如果存在显式配置为 Display == true 的属性，按 Index 升序排列返回
// 策略 2：若未显式配置，进行智能兜底：按 SortKey、Index 升序排列，排除文件与内置字段，返回前 6 个核心属性
func ResolveDisplayFields(attrs []Attribute) []Attribute {
	// 策略 1：优先提取显式配置为展示列的属性，按 Index 升序
	displays := lo.Filter(attrs, func(a Attribute, _ int) bool { return a.Display })
	if len(displays) > 0 {
		slices.SortFunc(displays, func(a, b Attribute) int {
			return cmp.Compare(a.Index, b.Index)
		})
		return displays
	}

	// 策略 2：智能兜底，优先排除文件与内置字段
	candidates := lo.Filter(attrs, func(a Attribute, _ int) bool {
		return a.FieldType != "file" && !a.Builtin
	})
	if len(candidates) == 0 {
		candidates = slices.Clone(attrs)
	}

	// 级联升序排序：优先比较 SortKey，相同则按 Index 排序
	slices.SortFunc(candidates, func(a, b Attribute) int {
		if n := cmp.Compare(a.SortKey, b.SortKey); n != 0 {
			return n
		}
		return cmp.Compare(a.Index, b.Index)
	})

	return lo.Subset(candidates, 0, defaultDisplayLimit)
}

