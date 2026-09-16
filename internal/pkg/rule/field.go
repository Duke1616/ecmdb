package rule

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Duke1616/ecmdb/internal/pkg/wechat"
	"github.com/samber/lo"
)

type Field struct {
	IsShort bool   `json:"is_short"`
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

// GetFields 根据表单规则格式化消息通知卡片字段（只读无副作用）
func GetFields(rules []Rule, provide uint8, data map[string]interface{}) []Field {
	if len(data) == 0 {
		return nil
	}

	switch provide {
	case SystemProvide:
		return processSystemFields(rules, data)
	case WechatProvide:
		return processWechatFields(data)
	default:
		return nil
	}
}

// processSystemFields 按规则预设顺序流式生成卡片字段（单次遍历保序，零多余中转）
func processSystemFields(rules []Rule, data map[string]interface{}) []Field {
	var fields []Field
	handled := make(map[string]struct{}, len(data))

	// 1. 优先按 Rules 声明顺序生成已登记的字段
	for _, r := range rules {
		if r.Field == "" {
			continue
		}
		if r.IsHidden() {
			handled[r.Field] = struct{}{} // 标记隐藏字段已处理，避免被末尾兜底误加
			continue
		}
		val, exists := data[r.Field]
		if !exists {
			continue
		}

		handled[r.Field] = struct{}{}
		title := lo.Ternary(r.Title != "", r.Title, r.Field)
		fields = append(fields, newLarkField(title, formatValue(val, r.Options)))
	}

	// 2. 兜底追加未在 Rules 中定义的额外数据字段（按字母序排在末尾）
	extraKeys := lo.Filter(lo.Keys(data), func(k string, _ int) bool {
		_, seen := handled[k]
		return !seen
	})
	slices.Sort(extraKeys)
	for _, k := range extraKeys {
		fields = append(fields, newLarkField(k, formatValue(data[k], nil)))
	}

	return AddRowSpacers(fields)
}

// formatValue 格式化单值或切片，若命中 Options 则自动映射为友好 Label
func formatValue(val any, options []Options) string {
	if val == nil {
		return ""
	}

	// 无选项映射时，直接转为字符串（支持切片展开）
	if len(options) == 0 {
		if items, ok := toStringSlice(val); ok {
			return strings.Join(items, ", ")
		}
		return fmt.Sprint(val)
	}

	// 有选项映射：统一按字符串映射 Label
	optMap := lo.SliceToMap(options, func(o Options) (string, string) {
		return fmt.Sprint(o.Value), o.Label
	})
	mapLabel := func(v any) string {
		str := fmt.Sprint(v)
		if label, ok := optMap[str]; ok {
			return label
		}
		return str
	}

	if items, ok := toAnySlice(val); ok {
		return strings.Join(lo.Map(items, func(item any, _ int) string {
			return mapLabel(item)
		}), ", ")
	}
	return mapLabel(val)
}

func toStringSlice(val any) ([]string, bool) {
	switch v := val.(type) {
	case []string:
		return v, true
	case []any:
		return lo.Map(v, func(item any, _ int) string { return fmt.Sprint(item) }), true
	default:
		return nil, false
	}
}

func toAnySlice(val any) ([]any, bool) {
	switch v := val.(type) {
	case []any:
		return v, true
	case []string:
		return lo.ToAnySlice(v), true
	default:
		return nil, false
	}
}

func newLarkField(title, content string) Field {
	return Field{
		IsShort: true,
		Tag:     "lark_md",
		Content: fmt.Sprintf("**%s:**\n%s", title, content),
	}
}

// processWechatFields 处理企业微信 OA 审批流数据字段
func processWechatFields(data map[string]interface{}) []Field {
	oaData, err := wechat.Unmarshal(data)
	if err != nil {
		return nil
	}

	var fields []Field
	for _, c := range oaData.ApplyData.Contents {
		if len(c.Title) == 0 {
			continue
		}
		var content string
		switch c.Control {
		case "Selector":
			var vals []string
			for _, opt := range c.Value.Selector.Options {
				for _, v := range opt.Value {
					if v.Text != "" {
						vals = append(vals, v.Text)
					}
				}
			}
			content = strings.Join(vals, ", ")
		case "Textarea", "Text":
			content = c.Value.Text
		}

		if content != "" {
			fields = append(fields, newLarkField(c.Title[0].Text, content))
		}
	}
	return fields
}

// AddRowSpacers 为飞书双列排版自适应换行占位（每两个短元素插入一个空 spacer）
func AddRowSpacers(fields []Field) []Field {
	if len(fields) <= 1 {
		return fields
	}

	results := make([]Field, 0, len(fields)+len(fields)/2)
	for i, f := range fields {
		results = append(results, f)
		if (i+1)%2 == 0 {
			results = append(results, Field{
				IsShort: false,
				Tag:     "lark_md",
				Content: "",
			})
		}
	}
	return results
}
