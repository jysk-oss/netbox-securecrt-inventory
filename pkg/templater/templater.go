package templater

import (
	"fmt"
	"strings"

	"github.com/jysk-network/netbox-securecrt-inventory/internal/config"
)

type TemplateVariables struct {
	key   string
	value string
}

func GetTemplate(deft string, overwrites []config.ConfigTemplateOverwrite, vars map[string]string) string {
	for _, overwrite := range overwrites {
		if vars[overwrite.Key] == overwrite.Value {
			return overwrite.Template
		}
	}

	return deft
}

func ApplyTemplate(template string, vars map[string]string) string {
	for key, value := range vars {
		template = strings.ReplaceAll(template, fmt.Sprintf("{%s}", key), value)
	}

	return template
}
