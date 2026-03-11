package rule

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/Duke1616/ecmdb/internal/pkg/wechat"
	"github.com/ecodeclub/ekit/slice"
	"github.com/xen0n/go-workwx"
)

type Field struct {
	IsShort bool   `json:"is_short"`
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

type FieldProcessor struct {
	rules      []Rule
	provide    uint8
	data       map[string]interface{}
	ruleMap    map[string]Rule
	optionsMap map[string]map[interface{}]string
}

func GetFields(rules []Rule, provide uint8, data map[string]interface{}) []Field {
	fp := &FieldProcessor{
		rules:   rules,
		provide: provide,
		data:    data,
	}

	fp.initialize()
	fp.filterHiddenFields()

	switch provide {
	case SystemProvide:
		return fp.processSystemFields()
	case WechatProvide:
		return fp.processWechatFields()
	default:
		return nil
	}
}

func (fp *FieldProcessor) initialize() {
	fp.ruleMap = slice.ToMap(fp.rules, func(element Rule) string {
		return element.Field
	})

	fp.optionsMap = make(map[string]map[interface{}]string)
	for _, r := range fp.rules {
		if len(r.Options) > 0 {
			fp.optionsMap[r.Field] = slice.ToMapV(r.Options, func(opt Options) (interface{}, string) {
				return opt.Value, opt.Label
			})
		}
	}
}

func (fp *FieldProcessor) filterHiddenFields() {
	for _, rule := range fp.rules {
		if _, ok := rule.Style["notify_display"]; ok {
			delete(fp.data, rule.Field)
		}
	}
}

func (fp *FieldProcessor) processSystemFields() []Field {
	var fields []Field
	keys := fp.getSortedKeys()

	for _, field := range keys {
		value := fp.data[field]
		title := fp.getFieldTitle(field)
		displayValue := fp.getDisplayValue(field, value)

		fields = append(fields, Field{
			IsShort: true,
			Tag:     "lark_md",
			Content: fmt.Sprintf(`**%s:**\n%v`, title, displayValue),
		})
	}

	return AddRowSpacers(fields)
}

func (fp *FieldProcessor) processWechatFields() []Field {
	oaData, err := wechat.Unmarshal(fp.data)
	if err != nil {
		return nil
	}

	var fields []Field

	for _, contents := range oaData.ApplyData.Contents {
		key := contents.Title[0].Text
		content := fp.processWechatContent(contents)

		if content != "" {
			fields = append(fields, Field{
				IsShort: true,
				Tag:     "lark_md",
				Content: fmt.Sprintf(`**%s:**\n%v`, key, content),
			})
		}
	}

	return fields
}

func (fp *FieldProcessor) processWechatContent(contents workwx.OAContent) string {
	switch contents.Control {
	case "Selector":
		switch contents.Value.Selector.Type {
		case "single":
			return contents.Value.Selector.Options[0].Value[0].Text
		case "multi":
			values := slice.Map(contents.Value.Selector.Options, func(_ int, opt workwx.OAContentSelectorOption) string {
				return opt.Value[0].Text
			})
			return strings.Join(values, ", ")
		}
	case "Textarea":
		return contents.Value.Text
	default:
		return ""
	}
	return ""
}

func (fp *FieldProcessor) getSortedKeys() []string {
	keys := make([]string, 0, len(fp.data))
	for key := range fp.data {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (fp *FieldProcessor) getFieldTitle(field string) string {
	if rule, ok := fp.ruleMap[field]; ok {
		return rule.Title
	}
	return field
}

func (fp *FieldProcessor) getDisplayValue(field string, value interface{}) string {
	if value == nil {
		return ""
	}

	if reflect.TypeOf(value).Kind() == reflect.Slice {
		return fp.processSliceValue(field, value)
	}
	return fp.processSingleValue(field, value)
}

func (fp *FieldProcessor) processSliceValue(field string, value interface{}) string {
	sli := reflect.ValueOf(value)
	results := make([]string, 0, sli.Len())

	for i := 0; i < sli.Len(); i++ {
		results = append(results, fp.getOptionLabel(field, sli.Index(i).Interface()))
	}

	return strings.Join(results, ", ")
}

func (fp *FieldProcessor) processSingleValue(field string, value interface{}) string {
	if label := fp.getOptionLabel(field, value); label != "" {
		return label
	}
	return fmt.Sprintf("%v", value)
}

func (fp *FieldProcessor) getOptionLabel(field string, value interface{}) string {
	options, ok := fp.optionsMap[field]
	if !ok {
		return ""
	}

	// 1. 直接通过原始值匹配（针对 interface{} 的 key）
	if label, exists := options[value]; exists {
		return label
	}

	// 2. 尝试数字转换匹配
	if num, err := convertToNumber(value); err == nil {
		if label, exists := options[num]; exists {
			return label
		}
	}

	// 3. 最后退避到字符串匹配
	valueStr := fmt.Sprintf("%v", value)
	if label, exists := options[valueStr]; exists {
		return label
	}

	return ""
}

func convertToNumber(value interface{}) (float64, error) {
	switch v := value.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return reflect.ValueOf(v).Convert(reflect.TypeOf(float64(0))).Float(), nil
	default:
		return 0, fmt.Errorf("not a number")
	}
}

// AddRowSpacers 专门为 rule.Field 提供的排列空行补位函数
// 确保每个飞书短字段在达到双列满排后，插入占位使其下一项换行
func AddRowSpacers(fields []Field) []Field {
	var results []Field
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
