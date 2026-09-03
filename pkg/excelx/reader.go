package excelx

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"github.com/xuri/excelize/v2"
)

// ── Strategy 模式：类型转换策略 ────────────────────────────────────────────────

// coercer 类型转换策略函数签名
// raw: 原始字符串（含空格，用于错误信息）; trimmed: 已 TrimSpace 的值; col / lineNum 用于构造定位错误
type coercer func(raw, trimmed string, col Column, lineNum int) (interface{}, error)

// coercerRegistry 类型 → 转换策略注册表
// 遵循开闭原则：扩展新类型只需在此注册，核心调度逻辑 coerce() 永不改动
var coercerRegistry = map[string]coercer{
	"int":     coerceInt,
	"integer": coerceInt,

	"float":  coerceFloat,
	"double": coerceFloat,
	"number": coerceFloat,

	"bool":    coerceBool,
	"boolean": coerceBool,

	"select": coerceEnum,
	"list":   coerceEnum,
}

func coerceInt(raw, val string, col Column, lineNum int) (interface{}, error) {
	v, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("第 %d 行: 字段【%s(%s)】值 '%s' 不是合法的整数", lineNum, col.Title, col.Key, raw)
	}
	return v, nil
}

func coerceFloat(raw, val string, col Column, lineNum int) (interface{}, error) {
	v, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return nil, fmt.Errorf("第 %d 行: 字段【%s(%s)】值 '%s' 不是合法的浮点数", lineNum, col.Title, col.Key, raw)
	}
	return v, nil
}

func coerceBool(raw, val string, col Column, lineNum int) (interface{}, error) {
	switch strings.ToLower(val) {
	case "true", "是", "1", "yes", "y":
		return true, nil
	case "false", "否", "0", "no", "n":
		return false, nil
	default:
		return nil, fmt.Errorf("第 %d 行: 字段【%s(%s)】值 '%s' 不是合法的布尔值(支持 是/否、true/false)", lineNum, col.Title, col.Key, raw)
	}
}

func coerceEnum(raw, val string, col Column, lineNum int) (interface{}, error) {
	if len(col.Options) > 0 && !lo.Contains(col.Options, val) {
		return nil, fmt.Errorf("第 %d 行: 字段【%s(%s)】值 '%s' 不在允许的预设选项列表中 %v", lineNum, col.Title, col.Key, raw, col.Options)
	}
	return val, nil
}

// ── TableReader ───────────────────────────────────────────────────────────────

// TableReader 声明式 Excel 反序列化与校验引擎
type TableReader struct {
	columns []Column
}

// NewReader 实例化读取器，自动过滤 Hidden 列
func NewReader(columns []Column) *TableReader {
	return &TableReader{
		columns: lo.Filter(columns, func(c Column, _ int) bool { return c.IsVisible() }),
	}
}

// Read 从 Excel 二进制流中按三层表头契约解析数据行，返回强类型校验后的记录集
func (r *TableReader) Read(fileData []byte) ([]map[string]interface{}, error) {
	rows, err := r.loadRows(fileData)
	if err != nil {
		return nil, err
	}

	colIndexMap, err := r.buildColIndexMap(rows[1])
	if err != nil {
		return nil, err
	}

	return r.parseDataRows(rows[3:], colIndexMap)
}

// loadRows 打开 Excel 文件并加载第一个工作表的所有行
func (r *TableReader) loadRows(fileData []byte) ([][]string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileData))
	if err != nil {
		return nil, fmt.Errorf("解析 Excel 文件失败: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("Excel 文件未包含有效工作表")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取 Excel 数据失败: %w", err)
	}
	if len(rows) < 4 {
		return nil, fmt.Errorf("Excel 格式错误：至少需包含 3 行规范表头与 1 行有效数据")
	}
	return rows, nil
}

// buildColIndexMap 读取第二行（Key 行），建立列索引到 Column Schema 的映射
func (r *TableReader) buildColIndexMap(keyRow []string) (map[int]Column, error) {
	colByKey := lo.KeyBy(r.columns, func(c Column) string { return c.Key })
	indexMap := make(map[int]Column, len(keyRow))

	for colIdx, key := range keyRow {
		if col, ok := colByKey[strings.TrimSpace(key)]; ok {
			indexMap[colIdx] = col
		}
	}

	if len(indexMap) == 0 {
		return nil, fmt.Errorf("未在第 2 行表头中匹配到任何已知字段 Key，请使用规范模板上传")
	}
	return indexMap, nil
}

// parseDataRows 逐行解析数据（第 4 行起），收集校验错误后统一报告
func (r *TableReader) parseDataRows(dataRows [][]string, colIndexMap map[int]Column) ([]map[string]interface{}, error) {
	var (
		records   []map[string]interface{}
		parseErrs []error
	)

	for rIdx, row := range dataRows {
		lineNum := rIdx + 4 // 对应 Excel 实际行号
		record, hasContent, rowErrs := r.parseRow(row, colIndexMap, lineNum)
		parseErrs = append(parseErrs, rowErrs...)
		if hasContent {
			records = append(records, record)
		}
	}

	if len(parseErrs) > 0 {
		return nil, aggregateErrors(parseErrs)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("未在文件中检测到有效数据行，请填写后重试")
	}
	return records, nil
}

// parseRow 解析单行数据：执行类型转换、必填校验，返回记录、是否有内容标记与行级错误列表
func (r *TableReader) parseRow(row []string, colIndexMap map[int]Column, lineNum int) (map[string]interface{}, bool, []error) {
	record := make(map[string]interface{}, len(colIndexMap))
	var errs []error
	hasContent := false

	for cIdx, rawVal := range row {
		col, ok := colIndexMap[cIdx]
		if !ok {
			continue
		}
		val, err := coerce(rawVal, col, lineNum)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if val != nil {
			record[col.Key] = val
			hasContent = true
		}
	}

	// 补全检查：尾部截断时必填列可能缺失
	for cIdx, col := range colIndexMap {
		if cIdx >= len(row) && col.Required {
			errs = append(errs, fmt.Errorf("第 %d 行: 必填字段【%s(%s)】缺失", lineNum, col.Title, col.Key))
		}
	}

	return record, hasContent, errs
}

// coerce 单元格值调度入口：按字段类型从注册表中查找对应策略函数执行转换
// 未注册的类型（如 string）直接返回字符串原值
func coerce(raw string, col Column, lineNum int) (interface{}, error) {
	trimmed := strings.TrimSpace(raw)

	if trimmed == "" {
		if col.Required {
			return nil, fmt.Errorf("第 %d 行: 必填字段【%s(%s)】不能为空", lineNum, col.Title, col.Key)
		}
		return nil, nil
	}

	if fn, ok := coercerRegistry[strings.ToLower(col.Type)]; ok {
		return fn(raw, trimmed, col, lineNum)
	}
	return trimmed, nil
}

// aggregateErrors 聚合行级错误，最多上报前 5 条，超出部分给出剩余数量提示
func aggregateErrors(errs []error) error {
	const maxReport = 5
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Excel 数据校验失败（共发现 %d 处错误）:\n", len(errs)))

	reportCount := min(len(errs), maxReport)
	for i := 0; i < reportCount; i++ {
		sb.WriteString(fmt.Sprintf("  - %s\n", errs[i].Error()))
	}
	if len(errs) > maxReport {
		sb.WriteString(fmt.Sprintf("  ... 其余 %d 处错误已省略，请修正后重试", len(errs)-maxReport))
	}
	return fmt.Errorf("%s", sb.String())
}
