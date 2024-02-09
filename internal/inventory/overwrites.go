package inventory

import (
	"regexp"

	"github.com/jysk-network/netbox-securecrt-inventory/internal/config"
)

func applyNameOverwrites(name string, overwrites []config.ConfigNameOverwrite) string {
	for i := 0; i < len(overwrites); i++ {
		re, err := regexp.Compile(overwrites[i].Regex)
		if err != nil {
			continue
		}

		name = re.ReplaceAllString(name, overwrites[i].Value)
	}

	return name
}
