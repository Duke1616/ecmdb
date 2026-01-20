package excel

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// Builder Excel 构建器
// NOTE: 封装 excelize,提供流式、声明式的 API
type Builder struct {
	file          *excelize.File
	sheetName     string
	headers       []string
	use3RowHeader bool
	headerRows    [][]string // 3 行表头数据
	rows          [][]interface{}
	styles        *StyleSet
	autoWidth     bool // 是否自动调整列宽
}

// NewBuilder 创建 Excel 构建器
func NewBuilder(sheetName string) *Builder {
	file := excelize.NewFile()
	return &Builder{
		file:      file,
		sheetName: sheetName,
		rows:      make([][]interface{}, 0),
		styles:    createStyleSet(file),
		autoWidth: true, // 默认启用自动列宽
	}
}

// WithHeaders 设置单行表头
func (b *Builder) WithHeaders(headers ...string) *Builder {
	b.headers = headers
	b.use3RowHeader = false
	return b
}

// With3RowHeaders 设置 3 行表头
// row1: 字段 UID, row2: 字段名称, row3: 字段约束
func (b *Builder) With3RowHeaders(row1, row2, row3 []string) *Builder {
	b.headerRows = [][]string{row1, row2, row3}
	b.use3RowHeader = true
	return b
}

// AddRow 添加数据行
func (b *Builder) AddRow(data ...interface{}) *Builder {
	b.rows = append(b.rows, data)
	return b
}

// AddRows 批量添加数据行
func (b *Builder) AddRows(rows [][]interface{}) *Builder {
	b.rows = append(b.rows, rows...)
	return b
}

// WithStyle 设置样式
func (b *Builder) WithStyle(style *StyleSet) *Builder {
	b.styles = style
	return b
}

// WithColumnWidths 设置列宽
func (b *Builder) WithColumnWidths(widths map[int]float64) *Builder {
	b.autoWidth = false // 手动设置列宽时禁用自动列宽
	for colIdx, width := range widths {
		col, _ := excelize.ColumnNumberToName(colIdx + 1)
		b.file.SetColWidth(b.sheetName, col, col, width)
	}
	return b
}

// WithValidation 添加数据验证
func (b *Builder) WithValidation(colIdx int, options []string, startRow, endRow int) *Builder {
	if len(options) == 0 {
		return b
	}

	col, _ := excelize.ColumnNumberToName(colIdx + 1)
	dv := excelize.NewDataValidation(true)
	dv.Sqref = fmt.Sprintf("%s%d:%s%d", col, startRow, col, endRow)
	dv.SetDropList(options)
	b.file.AddDataValidation(b.sheetName, dv)
	return b
}

// Build 构建 Excel 文件
func (b *Builder) Build() error {
	// 1. 写入表头
	if b.use3RowHeader {
		if err := b.write3RowHeaders(); err != nil {
			return err
		}
	} else if len(b.headers) > 0 {
		if err := b.writeHeaders(); err != nil {
			return err
		}
	}

	// 2. 写入数据行
	if err := b.writeRows(); err != nil {
		return err
	}

	// 3. 自动调整列宽
	if b.autoWidth {
		if err := b.autoAdjustColumnWidths(); err != nil {
			return err
		}
	}

	// 4. 冻结首行
	if err := b.freezeFirstRow(); err != nil {
		return err
	}

	return nil
}

// ToBytes 导出为字节数组
func (b *Builder) ToBytes() ([]byte, error) {
	if err := b.Build(); err != nil {
		return nil, err
	}

	buf, err := b.file.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("生成 Excel 文件失败: %w", err)
	}

	return buf.Bytes(), nil
}

// Close 关闭文件
func (b *Builder) Close() error {
	return b.file.Close()
}

// writeHeaders 写入单行表头
func (b *Builder) writeHeaders() error {
	for colIdx, header := range b.headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		if err := b.file.SetCellValue(b.sheetName, cell, header); err != nil {
			return err
		}
		if err := b.file.SetCellStyle(b.sheetName, cell, cell, b.styles.Header); err != nil {
			return err
		}
	}

	// 设置表头行高
	if err := b.file.SetRowHeight(b.sheetName, 1, 30); err != nil {
		return err
	}

	return nil
}

// write3RowHeaders 写入 3 行表头
func (b *Builder) write3RowHeaders() error {
	// 为每行选择不同的样式
	styles := []int{b.styles.Header, b.styles.Header2, b.styles.Header3}

	for rowIdx, rowData := range b.headerRows {
		for colIdx, value := range rowData {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1)
			if err := b.file.SetCellValue(b.sheetName, cell, value); err != nil {
				return err
			}
			// 使用对应行的样式
			if err := b.file.SetCellStyle(b.sheetName, cell, cell, styles[rowIdx]); err != nil {
				return err
			}
		}
	}

	// 设置 3 行表头的行高
	for i := 1; i <= 3; i++ {
		if err := b.file.SetRowHeight(b.sheetName, i, 30); err != nil {
			return err
		}
	}

	return nil
}

// writeRows 写入数据行
func (b *Builder) writeRows() error {
	// 计算数据起始行(单行表头从第 2 行开始,3 行表头从第 4 行开始)
	startRow := 2
	if b.use3RowHeader {
		startRow = 4
	}

	for rowIdx, row := range b.rows {
		actualRow := startRow + rowIdx

		// 选择样式(斑马纹)
		style := b.styles.OddRow
		if rowIdx%2 == 1 {
			style = b.styles.EvenRow
		}

		for colIdx, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, actualRow)
			if err := b.file.SetCellValue(b.sheetName, cell, value); err != nil {
				return err
			}
			if err := b.file.SetCellStyle(b.sheetName, cell, cell, style); err != nil {
				return err
			}
		}
	}

	return nil
}

// autoAdjustColumnWidths 自动调整列宽
func (b *Builder) autoAdjustColumnWidths() error {
	// 计算每列的最大宽度
	colCount := 0
	if b.use3RowHeader && len(b.headerRows) > 0 {
		colCount = len(b.headerRows[0])
	} else if len(b.headers) > 0 {
		colCount = len(b.headers)
	}

	for colIdx := 0; colIdx < colCount; colIdx++ {
		maxWidth := 10.0 // 最小宽度

		// 检查表头宽度
		if b.use3RowHeader {
			for _, row := range b.headerRows {
				if colIdx < len(row) {
					width := calculateStringWidth(row[colIdx])
					if width > maxWidth {
						maxWidth = width
					}
				}
			}
		} else if colIdx < len(b.headers) {
			width := calculateStringWidth(b.headers[colIdx])
			if width > maxWidth {
				maxWidth = width
			}
		}

		// 检查数据行宽度(只检查前 100 行以提高性能)
		checkRows := len(b.rows)
		if checkRows > 100 {
			checkRows = 100
		}
		for rowIdx := 0; rowIdx < checkRows; rowIdx++ {
			if colIdx < len(b.rows[rowIdx]) {
				width := calculateStringWidth(fmt.Sprintf("%v", b.rows[rowIdx][colIdx]))
				if width > maxWidth {
					maxWidth = width
				}
			}
		}

		// 设置列宽(最大不超过 50)
		if maxWidth > 50 {
			maxWidth = 50
		}

		// 添加一些边距
		maxWidth += 2

		col, _ := excelize.ColumnNumberToName(colIdx + 1)
		if err := b.file.SetColWidth(b.sheetName, col, col, maxWidth); err != nil {
			return err
		}
	}

	return nil
}

// calculateStringWidth 计算字符串显示宽度
// NOTE: 中文字符宽度约为英文的 2 倍
func calculateStringWidth(s string) float64 {
	width := 0.0
	for _, r := range s {
		if r < 128 {
			// ASCII 字符
			width += 1.0
		} else {
			// 中文等宽字符
			width += 2.0
		}
	}
	// 转换为 Excel 列宽单位(大约)
	return width * 1.2
}

// freezeFirstRow 冻结首行
func (b *Builder) freezeFirstRow() error {
	// 3 行表头时冻结前 3 行
	ySplit := 1
	topLeftCell := "A2"
	if b.use3RowHeader {
		ySplit = 3
		topLeftCell = "A4"
	}

	return b.file.SetPanes(b.sheetName, &excelize.Panes{
		Freeze:      true,
		XSplit:      0,
		YSplit:      ySplit,
		TopLeftCell: topLeftCell,
		ActivePane:  "bottomLeft",
	})
}
