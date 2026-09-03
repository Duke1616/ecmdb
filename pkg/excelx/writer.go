package excelx

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/xuri/excelize/v2"
)

// TableWriter 声明式 Excel 表格生成器
type TableWriter struct {
	sheetName string
	columns   []Column
}

// NewWriter 实例化表格生成器，自动过滤 Hidden 列（如 file 附件类型）
func NewWriter(sheetName string, columns []Column) *TableWriter {
	if sheetName == "" {
		sheetName = "Sheet1"
	}
	return &TableWriter{
		sheetName: sheetName,
		columns:   lo.Filter(columns, func(c Column, _ int) bool { return c.IsVisible() }),
	}
}

// WriteTemplate 生成包含三层表头、100 行格式化录入区与下拉验证的空白导入模板
func (w *TableWriter) WriteTemplate() ([]byte, error) {
	return w.render(nil)
}

// WriteData 生成包含三层表头、数据记录与下拉验证的完整 Excel 字节流
func (w *TableWriter) WriteData(records []map[string]interface{}) ([]byte, error) {
	return w.render(records)
}

// render 内部统一渲染入口
func (w *TableWriter) render(records []map[string]interface{}) ([]byte, error) {
	file := excelize.NewFile()
	defer file.Close()

	w.initSheet(file)

	styles := createStyleSet(file)
	headers := w.buildHeaderRows()
	w.writeHeaders(file, headers, styles)
	w.writeDataRows(file, headers, records, styles)
	w.writeValidations(file, len(records))
	w.setColumnWidths(file, headers, records)
	w.freezeHeaderRows(file)

	buf, err := file.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("生成 Excel 文件字节流失败: %w", err)
	}
	return buf.Bytes(), nil
}

// initSheet 初始化工作表名称
func (w *TableWriter) initSheet(file *excelize.File) {
	if defaultSheet := file.GetSheetName(0); defaultSheet != w.sheetName {
		file.SetSheetName(defaultSheet, w.sheetName)
	}
}

// headerRows 三层表头行数据容器
type headerRows struct {
	titleRow      []string // Row 1: 中文业务标题（必填字段附 " *"）
	keyRow        []string // Row 2: 字段唯一 Key（机器反序列化对齐用）
	constraintRow []string // Row 3: 约束指引说明
}

// buildHeaderRows 根据 Column Schema 用 lo.Map 声明式提取三层表头内容
func (w *TableWriter) buildHeaderRows() headerRows {
	return headerRows{
		titleRow: lo.Map(w.columns, func(col Column, _ int) string {
			if col.Required {
				return col.Title + " *"
			}
			return col.Title
		}),
		keyRow:        lo.Map(w.columns, func(col Column, _ int) string { return col.Key }),
		constraintRow: lo.Map(w.columns, func(col Column, _ int) string { return col.ConstraintText() }),
	}
}

// writeHeaders 将三层表头写入工作表，设置对应样式与行高
func (w *TableWriter) writeHeaders(file *excelize.File, h headerRows, styles *StyleSet) {
	type headerSpec struct {
		data   []string
		style  int
		height float64
	}
	specs := []headerSpec{
		{h.titleRow, styles.TitleRow, headerTitleRowHeight},
		{h.keyRow, styles.KeyRow, headerKeyRowHeight},
		{h.constraintRow, styles.ConstraintRow, headerConstraintRowHeight},
	}
	for rowIdx, spec := range specs {
		excelRow := rowIdx + 1
		for colIdx, val := range spec.data {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, excelRow)
			_ = file.SetCellValue(w.sheetName, cell, val)
			_ = file.SetCellStyle(w.sheetName, cell, cell, spec.style)
		}
		_ = file.SetRowHeight(w.sheetName, excelRow, spec.height)
	}
}

// writeDataRows 写入数据行与格式化录入区（数据行从第 4 行开始）
func (w *TableWriter) writeDataRows(file *excelize.File, h headerRows, records []map[string]interface{}, styles *StyleSet) {
	dataRowCount := len(records)
	totalRenderRows := computeRenderRows(dataRowCount, records == nil)

	for rIdx := 0; rIdx < totalRenderRows; rIdx++ {
		excelRow := 4 + rIdx
		rowStyle := lo.Ternary(rIdx%2 == 0, styles.OddRow, styles.EvenRow)

		for cIdx, col := range w.columns {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, excelRow)
			_ = file.SetCellStyle(w.sheetName, cell, cell, rowStyle)

			if rIdx < dataRowCount && records != nil {
				if v, ok := records[rIdx][col.Key]; ok && v != nil {
					_ = file.SetCellValue(w.sheetName, cell, v)
				}
			}
		}
		_ = file.SetRowHeight(w.sheetName, excelRow, dataRowHeight)
	}
}

// computeRenderRows 计算需要渲染格式的总行数
// 模板时固定渲染 100 行，数据导出时最少保留 50 行格式，避免视觉断层
func computeRenderRows(dataRowCount int, isTemplate bool) int {
	if isTemplate {
		return templateRenderRows
	}
	return max(dataRowCount, exportMinRenderRows)
}

// writeValidations 为 select / list 列挂载下拉列表数据验证（覆盖至第 1000 行）
func (w *TableWriter) writeValidations(file *excelize.File, dataRowCount int) {
	endRow := max(validationMaxRow, dataRowCount+100)

	for cIdx, col := range w.columns {
		if !col.NeedsValidation() {
			continue
		}
		colName, _ := excelize.ColumnNumberToName(cIdx + 1)
		dv := excelize.NewDataValidation(true)
		dv.Sqref = fmt.Sprintf("%s4:%s%d", colName, colName, endRow)
		_ = dv.SetDropList(col.Options)
		_ = file.AddDataValidation(w.sheetName, dv)
	}
}

// setColumnWidths 根据表头与数据内容自适应计算列宽，消灭"细面条"局促感
func (w *TableWriter) setColumnWidths(file *excelize.File, h headerRows, records []map[string]interface{}) {
	allHeaderRows := [][]string{h.titleRow, h.keyRow, h.constraintRow}
	// 抽样上限 100 行
	sample := records[:min(len(records), 100)]

	for cIdx, col := range w.columns {
		// 表头各行宽度取最大值
		headerWidth := lo.Max(lo.FilterMap(allHeaderRows, func(row []string, _ int) (float64, bool) {
			if cIdx >= len(row) {
				return 0, false
			}
			return textWidth(row[cIdx]), true
		}))

		// 数据抽样宽度取最大值
		dataWidth := lo.Max(lo.FilterMap(sample, func(rec map[string]interface{}, _ int) (float64, bool) {
			v, ok := rec[col.Key]
			if !ok || v == nil {
				return 0, false
			}
			return textWidth(fmt.Sprint(v)), true
		}))

		width := min(max(baseColWidth(col), headerWidth, dataWidth), colWidthMax) + colWidthPadding

		colName, _ := excelize.ColumnNumberToName(cIdx + 1)
		_ = file.SetColWidth(w.sheetName, colName, colName, width)
	}
}

// freezeHeaderRows 冻结前三行表头，固定标题在滚动时始终可见
func (w *TableWriter) freezeHeaderRows(file *excelize.File) {
	_ = file.SetPanes(w.sheetName, &excelize.Panes{
		Freeze:      true,
		XSplit:      0,
		YSplit:      3,
		TopLeftCell: "A4",
		ActivePane:  "bottomLeft",
	})
}

// ── 列宽计算辅助函数 ──

// baseColWidth 根据列属性确定基础最小宽度
func baseColWidth(col Column) float64 {
	if col.Unique || col.Key == "name" || col.Type == "string" {
		return colWidthWide
	}
	return colWidthDefault
}

// textWidth 估算字符串在 Excel 中的显示宽度（中文宽度约为英文的 2 倍）
func textWidth(s string) float64 {
	width := 0.0
	for _, r := range strings.TrimSpace(s) {
		if r < 128 {
			width += 1.0
		} else {
			width += 2.0
		}
	}
	return width * 1.25
}
