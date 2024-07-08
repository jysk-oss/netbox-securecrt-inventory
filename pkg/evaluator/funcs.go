package evaluator

import (
	"fmt"
	"strings"

	"github.com/netbox-community/go-netbox/v3/netbox/models"
)

func FindTag(tags []*models.NestedTag, label string) *string {
	for i := 0; i < len(tags); i++ {
		if tags[i].Name != nil && strings.Contains(*tags[i].Name, label) {
			result := strings.TrimPrefix(*tags[i].Name, fmt.Sprintf("%s:", label))
			return &result
		}
	}

	return nil
}
