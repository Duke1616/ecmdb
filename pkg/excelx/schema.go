package excelx

import "strings"

// 表格布局常量，统一管理所有魔法数字避免散落各处
const (
	// 表头行高（单位：磅）
	headerTitleRowHeight      = 38.0
	headerKeyRowHeight        = 26.0
	headerConstraintRowHeight = 26.0

	// 数据区行高
	dataRowHeight = 26.0

	// 数据区渲染行数：模板预设 100 行，导出数据最少补齐 50 行
	templateRenderRows = 100
	exportMinRenderRows = 50

	// 下拉验证最大保护行数（保证用户在第 1000 行以内填写均有效）
	validationMaxRow = 1000

	// 列宽：基础最小宽度 / 名称或唯一键宽度 / 上限 / padding
	colWidthDefault  = 24.0
	colWidthWide     = 28.0
	colWidthMax      = 60.0
	colWidthPadding  = 4.0
)

// Column 声明式列 Schema 定义
// 驱动表格模板生成、数据导出渲染与导入校验反序列化，一次声明，导入导出共享
type Column struct {
	Key      string   // 字段唯一标识（供机器读写精准对齐，如 field_uid）
	Title    string   // 字段中文展示名称（供人类阅读，如 field_name）
	Type     string   // 数据类型：string / int / float / bool / select / list
	Required bool     // 是否必填约束
	Unique   bool     // 是否唯一约束（如资产名称 name）
	Secure   bool     // 是否敏感加密标记
	Options  []string // 预设合法枚举选项（select/list 类型生成下拉验证与值域校验）
	Hidden   bool     // 是否跳过导出（如 file 类型附件不参与 Excel 交换）
}

// ConstraintText 生成规范的约束描述文本，用于写入第三行约束指引表头
func (c Column) ConstraintText() string {
	tags := make([]string, 0, 4)
	if c.Unique {
		tags = append(tags, "唯一索引")
	}
	if c.Required {
		tags = append(tags, "必填")
	}
	if c.Secure {
		tags = append(tags, "加密")
	}
	// select/list 且配置了合法选项：提示用户从下拉中选择；其余字段统一提示手动输入
	if (c.Type == "select" || c.Type == "list") && len(c.Options) > 0 {
		tags = append(tags, "由用户选择")
	} else {
		tags = append(tags, "由用户输入")
	}
	return strings.Join(tags, " | ")
}

// NeedsValidation 判断当前列是否需要挂载下拉框数据验证
func (c Column) NeedsValidation() bool {
	return (c.Type == "select" || c.Type == "list") && len(c.Options) > 0
}

// IsVisible 判断是否应当参与导出（Hidden 类型如 file 附件跳过）
func (c Column) IsVisible() bool {
	return !c.Hidden
}
