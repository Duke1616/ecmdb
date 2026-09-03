package domain

import (
	"fmt"
	"reflect"
	"strings"
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
