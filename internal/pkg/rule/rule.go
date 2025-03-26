package rule

import "encoding/json"

type Rule struct {
	Type     string                 `json:"type"`
	Field    string                 `json:"field"`
	Title    string                 `json:"title"`
	Style    map[string]interface{} `json:"style"`
	Children []Rule                 `json:"children"`
	Options  []Options              `json:"options"`
}

type Options struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

type Rules []Rule

// FlattenRules 扁平化处理函数，排除 type 为 "col" 和 "fcRow" 的规则本身，但保留它们的子规则
func FlattenRules(rules []Rule) []Rule {
	var flattened []Rule
	for _, rule := range rules {
		if rule.Type == "col" || rule.Type == "fcRow" {
			if len(rule.Children) > 0 {
				flattened = append(flattened, FlattenRules(rule.Children)...)
			}
			continue
		}

		current := Rule{
			Type:    rule.Type,
			Field:   rule.Field,
			Title:   rule.Title,
			Style:   rule.Style,
			Options: rule.Options,
		}

		// 添加到结果集
		flattened = append(flattened, current)

		// 递归处理子规则（如果有）
		if len(rule.Children) > 0 {
			flattened = append(flattened, FlattenRules(rule.Children)...)
		}
	}
	return flattened
}

// ParseRules 解析模版字段
func ParseRules(ruleData interface{}) ([]Rule, error) {
	var rules []Rule
	rulesJson, err := json.Marshal(ruleData)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(rulesJson, &rules)
	return FlattenRules(rules), err
}

const (
	SystemProvide = 1
	WechatProvide = 2
)

type Data struct {
	Provide  uint8
	OderData map[string]interface{}
}
