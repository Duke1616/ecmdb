package rule

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/Duke1616/ecmdb/internal/pkg/wechat"
	"github.com/Duke1616/enotify/notify/feishu/card"
	"github.com/ecodeclub/ekit/slice"
	"github.com/xen0n/go-workwx"
)

type FieldProcessor struct {
	rules      []Rule
	provide    uint8
	data       map[string]interface{}
	ruleMap    map[string]Rule
	optionsMap map[string]map[interface{}]string
}

func GetFields(rules []Rule, provide uint8, data map[string]interface{}) []card.Field {
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
	for _, rule := range fp.rules {
		if len(rule.Options) > 0 {
			fieldOptions := make(map[interface{}]string)
			for _, opt := range rule.Options {
				fieldOptions[opt.Value] = opt.Label
			}
			fp.optionsMap[rule.Field] = fieldOptions
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

func (fp *FieldProcessor) processSystemFields() []card.Field {
	var fields []card.Field
	keys := fp.getSortedKeys()

	for i, field := range keys {
		value := fp.data[field]
		title := fp.getFieldTitle(field)
		displayValue := fp.getDisplayValue(field, value)

		fields = append(fields, card.Field{
			IsShort: true,
			Tag:     "lark_md",
			Content: fmt.Sprintf(`**%s:**\n%v`, title, displayValue),
		})

		if (i+1)%2 == 0 {
			fields = append(fields, card.Field{
				IsShort: false,
				Tag:     "lark_md",
				Content: "",
			})
		}
	}

	return fields
}

func (fp *FieldProcessor) processWechatFields() []card.Field {
	oaData, err := wechat.Unmarshal(fp.data)
	if err != nil {
		return nil
	}

	var fields []card.Field

	for i, contents := range oaData.ApplyData.Contents {
		key := contents.Title[0].Text
		content := fp.processWechatContent(contents)

		if content != "" {
			fields = append(fields, card.Field{
				IsShort: true,
				Tag:     "lark_md",
				Content: fmt.Sprintf(`**%s:**\n%v`, key, content),
			})

			if (i+1)%2 == 0 {
				fields = append(fields, card.Field{
					IsShort: false,
					Tag:     "lark_md",
					Content: "",
				})
			}
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
	var results []string

	for i := 0; i < sli.Len(); i++ {
		elem := sli.Index(i).Interface()
		results = append(results, fp.getOptionLabel(field, elem))
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

	valueStr := fmt.Sprintf("%v", value)
	if label, exists := options[valueStr]; exists {
		return label
	}

	if num, err := convertToNumber(value); err == nil {
		if label, exists := options[num]; exists {
			return label
		}
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
