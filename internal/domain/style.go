package domain

import "github.com/xuri/excelize/v2"

// StyleSet Excel 样式集
type StyleSet struct {
	Header  int // 第一行表头样式 (属性/约束)
	Header2 int // 第二行表头样式 (UID)
	Header3 int // 第三行表头样式 (名称)
	OddRow  int // 奇数行样式
	EvenRow int // 偶数行样式
}

// createStyleSet 创建样式集
// NOTE: 必须在同一个 file 实例上创建样式
func createStyleSet(f *excelize.File) *StyleSet {
	return &StyleSet{
		Header:  createHeaderRow1Style(f), // 属性/约束
		Header2: createHeaderRow2Style(f), // UID
		Header3: createHeaderRow3Style(f), // 名称
		OddRow:  createOddRowStyle(f),
		EvenRow: createEvenRowStyle(f),
	}
}

// createHeaderRow1Style 创建第一行表头样式 (属性/约束 - 淡蓝色背景)
func createHeaderRow1Style(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   false,
			Color:  "409EFF",
			Size:   10,
			Family: "Arial",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"ECF5FF"}, // 淡蓝色背景
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "EBEEF5", Style: 1},
			{Type: "right", Color: "EBEEF5", Style: 1},
			{Type: "top", Color: "EBEEF5", Style: 1},
			{Type: "bottom", Color: "EBEEF5", Style: 1},
		},
	})

	return style
}

// createHeaderRow2Style 创建第二行表头样式 (UID - 浅灰色背景)
func createHeaderRow2Style(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Color:  "303133",
			Size:   11,
			Family: "Arial",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"F5F7FA"}, // 浅灰色背景
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "EBEEF5", Style: 1},
			{Type: "right", Color: "EBEEF5", Style: 1},
			{Type: "top", Color: "EBEEF5", Style: 1},
			{Type: "bottom", Color: "EBEEF5", Style: 1},
		},
	})

	return style
}

// createHeaderRow3Style 创建第三行表头样式 (名称 - 浅灰色背景)
func createHeaderRow3Style(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Color:  "303133",
			Size:   11,
			Family: "Arial",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"F5F7FA"}, // 浅灰色背景
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "EBEEF5", Style: 1},
			{Type: "right", Color: "EBEEF5", Style: 1},
			{Type: "top", Color: "EBEEF5", Style: 1},
			{Type: "bottom", Color: "EBEEF5", Style: 1},
		},
	})

	return style
}

// createOddRowStyle 创建奇数行样式
func createOddRowStyle(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   10,
			Color:  "606266",
			Family: "Arial",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "EBEEF5", Style: 1},
			{Type: "right", Color: "EBEEF5", Style: 1},
			{Type: "top", Color: "EBEEF5", Style: 1},
			{Type: "bottom", Color: "EBEEF5", Style: 1},
		},
	})

	return style
}

// createEvenRowStyle 创建偶数行样式(斑马纹)
func createEvenRowStyle(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   10,
			Color:  "606266",
			Family: "Arial",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"FAFAFA"}, // 极浅灰色
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "EBEEF5", Style: 1},
			{Type: "right", Color: "EBEEF5", Style: 1},
			{Type: "top", Color: "EBEEF5", Style: 1},
			{Type: "bottom", Color: "EBEEF5", Style: 1},
		},
	})

	return style
}
