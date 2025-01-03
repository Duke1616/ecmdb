package rule

import "encoding/json"

type Rule struct {
	Type  string                 `json:"type"`
	Field string                 `json:"field"`
	Title string                 `json:"title"`
	Style map[string]interface{} `json:"style"`
}

type Rules []Rule

// ParseRules 解析模版字段
func ParseRules(ruleData interface{}) ([]Rule, error) {
	var rules []Rule
	rulesJson, err := json.Marshal(ruleData)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(rulesJson, &rules)
	return rules, err
}

const (
	SystemProvide = 1
	WechatProvide = 2
)

type Data struct {
	Provide  uint8
	OderData map[string]interface{}
}
