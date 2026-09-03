package excelx

import "github.com/xuri/excelize/v2"

// StyleSet 现代 SaaS 风格 Excel 样式集
type StyleSet struct {
	// 三层表头：业务主标题 / 技术 Key / 填写约束指引
	TitleRow      int
	KeyRow        int
	ConstraintRow int
	// 数据行：斑马纹奇偶行交替
	OddRow  int
	EvenRow int
}

const (
	fontFamily  = "Microsoft YaHei"
	borderLight = "E2E8F0" // 细腻浅灰边框
	borderDark  = "334155" // 深色表头内边框
	borderDiv   = "CBD5E1" // 表头与数据区分割线

	// 配色（Slate 调色盘，现代企业 SaaS 风格）
	colorTitleBg = "1E293B" // 深邃科技墨蓝（Slate-800）
	colorKeyBg   = "F1F5F9" // 浅灰（Slate-100）
	colorConsBg  = "F8FAFC" // 极淡（Slate-50）
	colorDataEven = "F8FAFC" // 偶数行斑马纹
	colorDataOdd  = "FFFFFF" // 奇数行纯白

	colorTitleText = "FFFFFF" // 主标题白字
	colorKeyText   = "475569" // Key 行深灰字（Slate-600）
	colorConsText  = "64748B" // 约束行冷灰字（Slate-500）
	colorDataText  = "334155" // 数据行暗灰字（Slate-700）
)

func createStyleSet(f *excelize.File) *StyleSet {
	return &StyleSet{
		TitleRow:      newTitleRowStyle(f),
		KeyRow:        newKeyRowStyle(f),
		ConstraintRow: newConstraintRowStyle(f),
		OddRow:        newOddRowStyle(f),
		EvenRow:       newEvenRowStyle(f),
	}
}

// newTitleRowStyle 主标题行：深邃墨蓝底，白字加粗 12pt，体现视觉统领地位
func newTitleRowStyle(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Color:  colorTitleText,
			Size:   12,
			Family: fontFamily,
		},
		Fill:      solidFill(colorTitleBg),
		Alignment: centerAlign(true),
		Border:    uniformBorder(borderDark, 1),
	})
	return style
}

// newKeyRowStyle 技术 Key 行：浅灰底，中深灰字 10pt，弱化机器感
func newKeyRowStyle(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Color:  colorKeyText,
			Size:   10,
			Family: fontFamily,
		},
		Fill:      solidFill(colorKeyBg),
		Alignment: centerAlign(true),
		Border:    uniformBorder(borderLight, 1),
	})
	return style
}

// newConstraintRowStyle 约束指引行：极淡底，柔和冷灰字 9.5pt，底边高亮分割
func newConstraintRowStyle(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Color:  colorConsText,
			Size:   9.5,
			Family: fontFamily,
		},
		Fill:      solidFill(colorConsBg),
		Alignment: centerAlign(true),
		Border: []excelize.Border{
			{Type: "left", Color: borderLight, Style: 1},
			{Type: "right", Color: borderLight, Style: 1},
			{Type: "top", Color: borderLight, Style: 1},
			{Type: "bottom", Color: borderDiv, Style: 2}, // 加粗底边分割线
		},
	})
	return style
}

// newOddRowStyle 奇数数据行：纯白底
func newOddRowStyle(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Color: colorDataText, Size: 10.5, Family: fontFamily},
		Fill:      solidFill(colorDataOdd),
		Alignment: centerAlign(false),
		Border:    uniformBorder(borderLight, 1),
	})
	return style
}

// newEvenRowStyle 偶数数据行：高级斑马纹浅灰底
func newEvenRowStyle(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Color: colorDataText, Size: 10.5, Family: fontFamily},
		Fill:      solidFill(colorDataEven),
		Alignment: centerAlign(false),
		Border:    uniformBorder(borderLight, 1),
	})
	return style
}

// ── 样式辅助函数，消灭重复的 Fill / Alignment / Border 结构体 ──

func solidFill(color string) excelize.Fill {
	return excelize.Fill{Type: "pattern", Color: []string{color}, Pattern: 1}
}

func centerAlign(wrapText bool) *excelize.Alignment {
	return &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: wrapText}
}

func uniformBorder(color string, style int) []excelize.Border {
	sides := []string{"left", "right", "top", "bottom"}
	borders := make([]excelize.Border, len(sides))
	for i, side := range sides {
		borders[i] = excelize.Border{Type: side, Color: color, Style: style}
	}
	return borders
}
