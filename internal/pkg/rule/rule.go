package rule

import (
	"encoding/json"

	"github.com/mitchellh/mapstructure"
)

type Rule struct {
	Type         string                 `json:"type"`
	Field        string                 `json:"field"`
	Title        string                 `json:"title"`
	Style        map[string]interface{} `json:"style"`
	NotifyHidden bool                   `json:"notify_hidden"`
	Children     []Rule                 `json:"children"`
	Options      []Options              `json:"options"`
}

type Options struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

type Rules []Rule

// IsHidden 统一判断该规则是否属于无需通知展示的隐藏字段
func (r Rule) IsHidden() bool {
	if r.NotifyHidden {
		return true
	}
	if len(r.Style) > 0 {
		if _, ok := r.Style["notify_display"]; ok {
			return true
		}
	}
	return false
}

// FlattenRules 扁平化处理函数，排除 type 为 "col" 和 "fcRow" 的规则本身，但递归保留它们的子规则
func FlattenRules(rules []Rule) []Rule {
	var flattened []Rule
	for _, rule := range rules {
		if rule.Type == "col" || rule.Type == "fcRow" {
			if len(rule.Children) > 0 {
				flattened = append(flattened, FlattenRules(rule.Children)...)
			}
			continue
		}

		leaf := Rule{
			Type:         rule.Type,
			Field:        rule.Field,
			Title:        rule.Title,
			Style:        rule.Style,
			NotifyHidden: rule.NotifyHidden,
			Options:      rule.Options,
		}
		flattened = append(flattened, leaf)

		if len(rule.Children) > 0 {
			flattened = append(flattened, FlattenRules(rule.Children)...)
		}
	}
	return flattened
}

// ParseRules 解析模版字段（兼顾兼容 string, []byte, map 或切片输入）
func ParseRules(ruleData interface{}) ([]Rule, error) {
	if ruleData == nil {
		return nil, nil
	}

	// 若入参为原始 JSON 字节或字符串，先解析
	if str, ok := ruleData.(string); ok {
		ruleData = []byte(str)
	}
	if rawBytes, ok := ruleData.([]byte); ok {
		var unmarshaled interface{}
		if err := json.Unmarshal(rawBytes, &unmarshaled); err != nil {
			return nil, err
		}
		ruleData = unmarshaled
	}

	var rules []Rule
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &rules,
		TagName: "json",
	})
	if err != nil {
		return nil, err
	}

	if err = decoder.Decode(ruleData); err != nil {
		return nil, err
	}

	return FlattenRules(rules), nil
}

const (
	SystemProvide uint8 = 1
	WechatProvide uint8 = 2
)

type Data struct {
	Provide   uint8                  `json:"provide"`
	OrderData map[string]interface{} `json:"order_data"`
}
