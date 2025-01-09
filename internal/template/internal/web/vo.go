package web

type CreateTemplateReq struct {
	Name       string `json:"name"`
	WorkflowId int64  `json:"workflow_id"`
	GroupId    int64  `json:"group_id"`
	Icon       string `json:"icon"`
	Rules      string `json:"rules"`
	Options    string `json:"options"`
	Desc       string `json:"desc"`
}

type DetailTemplateReq struct {
	Id int64 `json:"id"`
}

type Page struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

type ListTemplateReq struct {
	Page
}

type DeleteTemplateReq struct {
	Id int64 `json:"id"`
}

type CreateType uint8

type Template struct {
	Id         int64      `json:"id"`
	Name       string     `json:"name"`
	WorkflowId int64      `json:"workflow_id"`
	Icon       string     `json:"icon"`
	GroupId    int64      `json:"group_id"`
	CreateType CreateType `json:"create_type"`
	//Rules      []map[string]interface{} `json:"rules"`
	//Options    map[string]interface{}   `json:"options"`
	Rules   string `json:"rules"`
	Options string `json:"options"`
	Desc    string `json:"desc"`
}

type TemplateJson struct {
	Id         int64                    `json:"id"`
	Name       string                   `json:"name"`
	WorkflowId int64                    `json:"workflow_id"`
	Icon       string                   `json:"icon"`
	GroupId    int64                    `json:"group_id"`
	CreateType CreateType               `json:"create_type"`
	Rules      []map[string]interface{} `json:"rules"`
	Options    map[string]interface{}   `json:"options"`
	Desc       string                   `json:"desc"`
}

type RetrieveTemplates struct {
	Total     int64          `json:"total"`
	Templates []TemplateJson `json:"templates"`
}

type UpdateTemplateReq struct {
	Id         int64  `json:"id"`
	GroupId    int64  `json:"group_id"`
	Icon       string `json:"icon"`
	WorkflowId int64  `json:"workflow_id"`
	Name       string `json:"name"`
	Desc       string `json:"desc"`
	Rules      string `json:"rules"`
	Options    string `json:"options"`
}

type GetTemplatesByWorkFlowIdReq struct {
	WorkFlowId int64 `json:"workflow_id"`
}

type CreateTemplateGroupReq struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type TemplateGroup struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type RetrieveTemplateGroup struct {
	TemplateGroups []TemplateGroup `json:"template_groups"`
	Total          int64           `json:"total"`
}

// TemplateCombination 一组数据
type TemplateCombination struct {
	Id        int64      `json:"id"`
	Name      string     `json:"name"`
	Icon      string     `json:"icon"`
	Total     int64      `json:"total"`
	Templates []Template `json:"templates"`
}

type RetrieveTemplateCombination struct {
	TemplateCombinations []TemplateCombination `json:"template_combinations"`
}
