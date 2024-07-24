package evaluator

import (
	"fmt"
	"strings"

	"github.com/netbox-community/go-netbox/v4"
)

type Environment struct {
	SessionType                string `expr:"session_type"`
	Description                string `expr:"description"`
	Credential                 string `expr:"credential"`
	Path                       string `expr:"path"`
	PathTemplate               string `expr:"path_template"`
	DeviceName                 string `expr:"device_name"`
	DeviceNameTemplate         string `expr:"device_name_template"`
	Firewall                   string `expr:"firewall"`
	FirewallTemplate           string `expr:"firewall_template"`
	ConnectionProtocol         string `expr:"connection_protocol"`
	ConnectionProtocolTemplate string `expr:"connection_protocol_template"`
	DeviceRole                 string `expr:"device_role"`
	DeviceType                 string `expr:"device_type"`
	DeviceIP                   string `expr:"device_ip"`
	RegionName                 string `expr:"region_name"`
	TenantName                 string `expr:"tenant_name"`
	SiteName                   string `expr:"site_name"`
	SiteGroup                  string `expr:"site_group"`
	SiteAddress                string `expr:"site_address"`

	Device interface{} `expr:"device"`
	Site   interface{} `expr:"site"`
}

func (Environment) FindTag(tags []netbox.NestedTag, label string) *string {
	for i := 0; i < len(tags); i++ {
		if strings.Contains(tags[i].Name, label) {
			result := strings.TrimPrefix(tags[i].Name, fmt.Sprintf("%s:", label))
			return &result
		}
	}

	return nil
}
