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
