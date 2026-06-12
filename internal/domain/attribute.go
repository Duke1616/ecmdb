package domain

import (
	"fmt"
	"reflect"
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
	Builtin   bool
}

// GetID 实现 Sortable 接口
func (a Attribute) GetID() int64 { return a.ID }

// GetSortKey 实现 Sortable 接口
func (a Attribute) GetSortKey() int64 { return a.SortKey }

// GetOptionStrings 获取选项字符串列表
// NOTE: 统一处理 []string、[]interface{}、primitive.A 等类型
func (a *Attribute) GetOptionStrings() []string {
	if a.Option == nil {
		return nil
	}

	switch opts := a.Option.(type) {
	case []string:
		return opts
	case []interface{}:
		result := make([]string, 0, len(opts))
		for _, opt := range opts {
			result = append(result, fmt.Sprint(opt))
		}
		return result
	default:
		// 处理 MongoDB primitive.A 等 slice 类型
		v := reflect.ValueOf(a.Option)
		if v.Kind() == reflect.Slice {
			result := make([]string, v.Len())
			for i := 0; i < v.Len(); i++ {
				result[i] = fmt.Sprint(v.Index(i).Interface())
			}
			return result
		}
		return []string{fmt.Sprint(a.Option)}
	}
}

// IsSelectType 判断是否为选择类型字段
// NOTE: select 和 list 类型需要下拉列表验证
func (a *Attribute) IsSelectType() bool {
	return a.FieldType == "select" || a.FieldType == "list"
}

// NeedsValidation 判断是否需要数据验证
func (a *Attribute) NeedsValidation() bool {
	return a.IsSelectType() && len(a.GetOptionStrings()) > 0
}

// ToExcelRow 转换为 Excel 行数据
func (a *Attribute) ToExcelRow() []interface{} {
	return []interface{}{
		a.FieldName,
		a.FieldUid,
		a.FieldType,
		a.Required,
		a.Display,
		a.Secure,
		fmt.Sprintf("%v", a.Option),
		a.Index,
	}
}

// GetConstraintDescription 获取字段约束描述
// NOTE: 用于 Excel 模板的第三行表头
func (a *Attribute) GetConstraintDescription() string {
	constraints := []string{}

	// 模型唯一索引列
	if a.FieldUid == "name" {
		constraints = append(constraints, "唯一索引")
	}

	// 必填约束
	if a.Required {
		constraints = append(constraints, "必填")
	}

	// 加密约束
	if a.Secure {
		constraints = append(constraints, "加密")
	}

	// 选择类型约束
	if a.IsSelectType() {
		constraints = append(constraints, "由用户选择")
	}

	if len(constraints) == 0 {
		return ""
	}

	return fmt.Sprintf("%s", constraints[0]) + func() string {
		if len(constraints) > 1 {
			return " | " + constraints[1]
		}
		if len(constraints) > 2 {
			return " | " + constraints[2]
		}
		return ""
	}()
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
