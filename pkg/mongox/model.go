package mongox

// IModel 基础数据模型规范接口，AutoIDPlugin 通过此接口检测和设置 ID。
//
// 大多数场景推荐直接内嵌 Model 结构体，自动满足此接口，无需手写方法。
type IModel interface {
	// SetID 设置主键 ID
	SetID(id int64)

	// GetID 获取主键 ID
	GetID() int64
}

// ==========================================
// 开箱即用的基础结构体，内嵌后自动满足 IModel 接口
// ==========================================

// Model 基础模型结构体，内嵌后自动满足 IModel 接口。
//
// 用法:
//
//	type Task struct {
//	    mongox.Model
//	    Name string `bson:"name"`
//	}
//
// Task 自动拥有 SetID/GetID 方法，AutoIDPlugin 可正常识别并注入 ID。
type Model struct {
	ID int64 `bson:"id"`
}

// SetID 实现 IModel 接口
func (m *Model) SetID(id int64) { m.ID = id }

// GetID 实现 IModel 接口
func (m *Model) GetID() int64 { return m.ID }

// TenantModel 租户模型结构体，内嵌后 TenantPlugin 自动识别并注入 TenantID。
//
// 用法:
//
//	type Order struct {
//	    mongox.TenantModel
//	    Title string `bson:"title"`
//	}
//
// 不需要实现任何方法，TenantPlugin 通过反射自动检测和设置 TenantID 字段。
type TenantModel struct {
	Model
	TenantID int64 `bson:"tenant_id"`
}
