package rule

import (
	"fmt"
	"strings"
)

func GenerateTitle(displayName string, templateName string) string {
	if strings.HasSuffix(templateName, "申请") {
		return fmt.Sprintf("%s的%s", displayName, templateName)
	}
	return fmt.Sprintf("%s的%s申请", displayName, templateName)
}

func GenerateCCTitle(displayName string, templateName string) string {
	if strings.HasSuffix(templateName, "申请") {
		return fmt.Sprintf("%s抄送了%s给你", displayName, templateName)
	}
	return fmt.Sprintf("%s抄送了%s申请给你", displayName, templateName)
}
