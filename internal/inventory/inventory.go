package inventory

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jysk-network/netbox-securecrt-inventory/internal/config"
	"github.com/jysk-network/netbox-securecrt-inventory/internal/netbox"
	"github.com/jysk-network/netbox-securecrt-inventory/pkg/securecrt"
	"github.com/netbox-community/go-netbox/v3/netbox/models"
)

const (
	STATE_RUNNING = "running"
	STATE_DONE    = "done"
	STATE_ERROR   = "error"
)

type InventorySync struct {
	cfg              *config.Config
	nb               *netbox.NetBox
	scrt             *securecrt.SecureCRT
	syncCallback     func(state string, message string)
	periodicTicker   *time.Ticker
	stripRe          *regexp.Regexp
	sessionGenerator *SessionGenerator
}

func New(cfg *config.Config, nb *netbox.NetBox, scrt *securecrt.SecureCRT, syncCallback func(state string, message string)) *InventorySync {
	inv := InventorySync{
		cfg:              cfg,
		nb:               nb,
		scrt:             scrt,
		syncCallback:     syncCallback,
		periodicTicker:   time.NewTicker(time.Minute * time.Duration(*cfg.PeriodicSyncInterval)),
		stripRe:          regexp.MustCompile(`[\\/\?]`),
		sessionGenerator: NewSessionGenerator(),
	}

	return &inv
}

func (i *InventorySync) getSite(sites []*models.Site, siteID int64) (*models.Site, error) {
	for x := 0; x < len(sites); x++ {
		if sites[x].ID == siteID {
			return sites[x], nil
		}
	}

	return nil, ErrorFailedToFindSite
}

func (i *InventorySync) getTenant(device interface{}) string {
	nd, ok := device.(*models.DeviceWithConfigContext)
	if ok && nd != nil && nd.Tenant != nil {
		return *nd.Tenant.Name
	}

	vm, ok := device.(*models.VirtualMachineWithConfigContext)
	if ok && vm != nil && vm.Tenant != nil {
		return *vm.Tenant.Name
	}

	return "No Tenant"
}

func (i *InventorySync) writeSession(session *securecrt.SecureCRTSession) error {
	sessionData, err := i.scrt.BuildSessionData(session)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/%s/%s.ini", i.cfg.RootPath, session.Path, session.DeviceName)
	err = i.scrt.WriteSession(path, sessionData)
	if err != nil {
		return err
	}

	return nil
}

func (i *InventorySync) getCommonEnvironment(sync_type string) map[string]interface{} {
	return map[string]interface{}{
		"type":                         sync_type,
		"credential":                   i.cfg.Session.SessionOptions.Credential,
		"path_template":                i.cfg.Session.Path,
		"device_name_template":         i.cfg.Session.DeviceName,
		"firewall_template":            i.cfg.Session.SessionOptions.Firewall,
		"connection_protocol_template": i.cfg.Session.SessionOptions.ConnectionProtocol,
	}
}

func (i *InventorySync) syncDevices(devices []*models.DeviceWithConfigContext, sites []*models.Site) error {
	for _, device := range devices {
		site, err := i.getSite(sites, device.Site.ID)
		if err != nil {
			return err
		}

		tenant := i.getTenant(device)
		ipAddress := strings.Split(*device.PrimaryIp4.Address, "/")[0]
		siteAddress := strings.ReplaceAll(site.PhysicalAddress, "\r\n", ", ")
		deviceType := device.DeviceType.Display
		siteGroup := ""
		if site.Group != nil {
			siteGroup = *site.Group.Slug
		}

		env := mergeMaps(i.getCommonEnvironment("device"), map[string]interface{}{
			"device":       device,
			"device_name":  device.Display,
			"device_role":  *device.DeviceRole.Name,
			"device_type":  deviceType,
			"device_ip":    ipAddress,
			"region_name":  *site.Region.Name,
			"tenant_name":  tenant,
			"site":         site,
			"site_name":    site.Display,
			"site_group":   siteGroup,
			"site_address": siteAddress,
		})

		session, err := i.sessionGenerator.GenerateSession(i.cfg.Session.Overrides, env)
		if err != nil {
			return err
		}

		err = i.writeSession(session)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *InventorySync) syncVirtualMachines(devices []*models.VirtualMachineWithConfigContext, sites []*models.Site) error {
	for _, device := range devices {
		site, err := i.getSite(sites, device.Site.ID)
		if err != nil {
			return err
		}

		tenant := i.getTenant(device)
		ipAddress := strings.Split(*device.PrimaryIp4.Address, "/")[0]
		siteAddress := strings.ReplaceAll(site.PhysicalAddress, "\r\n", ", ")
		deviceType := ""
		if device.Platform != nil {
			deviceType = device.Platform.Display
		}

		siteGroup := ""
		if site.Group != nil {
			siteGroup = *site.Group.Slug
		}

		env := mergeMaps(i.getCommonEnvironment("virtual_machine"), map[string]interface{}{
			"device":       device,
			"device_name":  device.Display,
			"device_role":  "Virtual Machine",
			"device_type":  deviceType,
			"device_ip":    ipAddress,
			"region_name":  *site.Region.Name,
			"tenant_name":  tenant,
			"site":         site,
			"site_name":    site.Display,
			"site_group":   siteGroup,
			"site_address": siteAddress,
		})

		session, err := i.sessionGenerator.GenerateSession(i.cfg.Session.Overrides, env)
		if err != nil {
			return err
		}

		err = i.writeSession(session)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *InventorySync) runSync() error {
	err := i.nb.TestConnection()
	if err != nil {
		return err
	}

	i.syncCallback(STATE_RUNNING, "Running: Getting sites")
	sites, err := i.nb.GetSites()
	if err != nil {
		return err
	}

	i.syncCallback(STATE_RUNNING, "Running: Getting devices")
	devices, err := i.nb.GetDevices()
	if err != nil {
		return err
	}

	i.syncCallback(STATE_RUNNING, "Running: Getting Virtual Machines")
	vms, err := i.nb.GetVirtualMachines()
	if err != nil {
		return err
	}

	i.syncCallback(STATE_RUNNING, "Running: Removing old sessions")
	i.scrt.RemoveSessions(i.cfg.RootPath)

	i.syncCallback(STATE_RUNNING, "Running: Writing sessions")
	err = i.syncDevices(devices, sites)
	if err != nil {
		return err
	}

	err = i.syncVirtualMachines(vms, sites)
	if err != nil {
		return err
	}

	return nil
}

func (i *InventorySync) RunSync() {
	lastSync := time.Now()
	err := i.runSync()
	if err != nil {
		i.syncCallback(STATE_ERROR, err.Error())
	} else {
		i.syncCallback(STATE_DONE, fmt.Sprintf("Status: Last sync @ %s", lastSync.Format("15:04")))
	}
}

func (i *InventorySync) SetupPeriodicSync() {
	for range i.periodicTicker.C {
		if i.cfg.EnablePeriodicSync {
			i.RunSync()
		}
	}
}

func mergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
