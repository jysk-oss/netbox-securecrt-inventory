package netbox

type Region struct {
	Id   int32  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type SiteGroup struct {
	Id   int32  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Site struct {
	Id      int32  `json:"id"`
	Url     string `json:"url"`
	Display string `json:"display"`
	// Full name of the site
	Name            string     `json:"name"`
	Slug            string     `json:"slug"`
	PhysicalAddress string     `json:"physical_address"`
	Description     *string    `json:"description,omitempty"`
	Region          *Region    `json:"region,omitempty"`
	Group           *SiteGroup `json:"group,omitempty"`
}

type Manufacturer struct {
	Id          int32   `json:"id"`
	Url         string  `json:"url"`
	Display     string  `json:"display"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
}

type DeviceType struct {
	Id           int32        `json:"id"`
	Url          string       `json:"url"`
	Display      string       `json:"display"`
	Manufacturer Manufacturer `json:"manufacturer"`
	Model        string       `json:"model"`
	Slug         string       `json:"slug"`
	Description  *string      `json:"description,omitempty"`
}

type DeviceRole struct {
	Id                  int32   `json:"id"`
	Url                 string  `json:"url"`
	Display             string  `json:"display"`
	Name                string  `json:"name"`
	Slug                string  `json:"slug"`
	Description         *string `json:"description,omitempty"`
	VirtualmachineCount int64   `json:"virtualmachine_count"`
}

type VirtualChassis struct {
	Id          int32   `json:"id"`
	Url         string  `json:"url"`
	Display     string  `json:"display"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type Tenant struct {
	Id          int32   `json:"id"`
	Url         string  `json:"url"`
	Display     string  `json:"display"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
}

type Platform struct {
	Id                  int32   `json:"id"`
	Url                 string  `json:"url"`
	Display             string  `json:"display"`
	Name                string  `json:"name"`
	Slug                string  `json:"slug"`
	Description         *string `json:"description,omitempty"`
	VirtualmachineCount int64   `json:"virtualmachine_count"`
}

type Location struct {
	Id          int32   `json:"id"`
	Url         string  `json:"url"`
	Display     string  `json:"display"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	Depth       int32   `json:"_depth"`
}

type Rack struct {
	Id          int32   `json:"id"`
	Url         string  `json:"url"`
	Display     string  `json:"display"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type DeviceStatus struct {
	Value *string `json:"value,omitempty"`
	Label *string `json:"label,omitempty"`
}

type IPAddress struct {
	Id          int32   `json:"id"`
	Url         string  `json:"url"`
	Display     string  `json:"display"`
	Address     string  `json:"address"`
	Description *string `json:"description,omitempty"`
}

type NestedTag struct {
	Id      int32   `json:"id"`
	Url     string  `json:"url"`
	Display string  `json:"display"`
	Name    string  `json:"name"`
	Slug    string  `json:"slug"`
	Color   *string `json:"color,omitempty"`
}

type DeviceWithConfigContext struct {
	Id         int32      `json:"id"`
	Url        string     `json:"url"`
	Display    string     `json:"display"`
	Name       string     `json:"name,omitempty"`
	DeviceType DeviceType `json:"device_type"`
	Role       DeviceRole `json:"role"`
	Tenant     *Tenant    `json:"tenant,omitempty"`
	Platform   *Platform  `json:"platform,omitempty"`
	Serial     *string    `json:"serial,omitempty"`
	AssetTag   *string    `json:"asset_tag,omitempty"`
	Site       Site       `json:"site"`
	Location   *Location  `json:"location,omitempty"`
	Rack       *Rack      `json:"rack,omitempty"`
	Position   *float64   `json:"position,omitempty"`
	// GPS coordinate in decimal format (xx.yyyyyy)
	Latitude *float64 `json:"latitude,omitempty"`
	// GPS coordinate in decimal format (xx.yyyyyy)
	Longitude      *float64        `json:"longitude,omitempty"`
	Status         *DeviceStatus   `json:"status,omitempty"`
	PrimaryIp      *IPAddress      `json:"primary_ip"`
	PrimaryIp4     *IPAddress      `json:"primary_ip4,omitempty"`
	PrimaryIp6     *IPAddress      `json:"primary_ip6,omitempty"`
	OobIp          *IPAddress      `json:"oob_ip,omitempty"`
	Cluster        *IPAddress      `json:"cluster,omitempty"`
	VirtualChassis *VirtualChassis `json:"virtual_chassis,omitempty"`
	VcPosition     *int32          `json:"vc_position,omitempty"`
	// Virtual chassis master election priority
	VcPriority    *int32      `json:"vc_priority,omitempty"`
	Description   *string     `json:"description,omitempty"`
	Comments      *string     `json:"comments,omitempty"`
	ConfigContext interface{} `json:"config_context"`
	// Local config context data takes precedence over source contexts in the final rendered config context
	LocalContextData     interface{}            `json:"local_context_data,omitempty"`
	Tags                 []NestedTag            `json:"tags,omitempty"`
	CustomFields         map[string]interface{} `json:"custom_fields,omitempty"`
	AdditionalProperties map[string]interface{}
}

type Cluster struct {
	Id          int32   `json:"id"`
	Url         string  `json:"url"`
	Display     string  `json:"display"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type VirtualMachineWithConfigContext struct {
	Id          int32         `json:"id"`
	Url         string        `json:"url"`
	Display     string        `json:"display"`
	Name        string        `json:"name"`
	Status      *DeviceStatus `json:"status,omitempty"`
	Site        *Site         `json:"site,omitempty"`
	Cluster     *Cluster      `json:"cluster,omitempty"`
	Role        *DeviceRole   `json:"role,omitempty"`
	Tenant      *Tenant       `json:"tenant,omitempty"`
	Platform    *Platform     `json:"platform,omitempty"`
	PrimaryIp   *IPAddress    `json:"primary_ip"`
	PrimaryIp4  *IPAddress    `json:"primary_ip4,omitempty"`
	PrimaryIp6  *IPAddress    `json:"primary_ip6,omitempty"`
	Vcpus       *float64      `json:"vcpus,omitempty"`
	Memory      *int32        `json:"memory,omitempty"`
	Disk        *int32        `json:"disk,omitempty"`
	Description *string       `json:"description,omitempty"`
	Comments    *string       `json:"comments,omitempty"`
	// Local config context data takes precedence over source contexts in the final rendered config context
	LocalContextData     interface{}            `json:"local_context_data,omitempty"`
	Tags                 []NestedTag            `json:"tags,omitempty"`
	CustomFields         map[string]interface{} `json:"custom_fields,omitempty"`
	ConfigContext        interface{}            `json:"config_context"`
	AdditionalProperties map[string]interface{}
}
