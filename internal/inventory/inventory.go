package inventory

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jysk-network/netbox-securecrt-inventory/internal/config"
	"github.com/jysk-network/netbox-securecrt-inventory/internal/netbox"
	"github.com/jysk-network/netbox-securecrt-inventory/pkg/evaluator"
	"github.com/jysk-network/netbox-securecrt-inventory/pkg/securecrt"
)

const (
	STATE_RUNNING = "running"
	STATE_DONE    = "done"
	STATE_ERROR   = "error"
)

type InventorySync struct {
	cfg            *config.Config
	nb             *netbox.NetBox
	scrt           *securecrt.SecureCRT
	stateLogger    func(state string, message string)
	periodicTicker *time.Ticker
	stripRe        *regexp.Regexp
}

func New(cfg *config.Config, nb *netbox.NetBox, scrt *securecrt.SecureCRT, stateLogger func(state string, message string)) *InventorySync {
	inv := InventorySync{
		cfg:            cfg,
		nb:             nb,
		scrt:           scrt,
		stateLogger:    stateLogger,
		periodicTicker: time.NewTicker(time.Minute * time.Duration(*cfg.PeriodicSyncInterval)),
		stripRe:        regexp.MustCompile(`[\\/\?]`),
	}

	return &inv
}

func (i *InventorySync) getSite(sites []netbox.Site, siteID int32) (*netbox.Site, error) {
	for x := 0; x < len(sites); x++ {
		if sites[x].Id == siteID {
			return &sites[x], nil
		}
	}

	return nil, ErrorFailedToFindSite
}

func (i *InventorySync) getRegionName(site *netbox.Site) string {
	if site.Region != nil {
		return site.Region.Name
	}
	return "No Region"
}

func (i *InventorySync) getTenant(device interface{}) string {
	nd, ok := device.(netbox.DeviceWithConfigContext)
	if ok && nd.Tenant != nil {
		return nd.Tenant.Name
	}

	vm, ok := device.(netbox.VirtualMachineWithConfigContext)
	if ok && vm.Tenant != nil {
		return vm.Tenant.Name
	}

	return "No Tenant"
}

func (i *InventorySync) getPrimaryIP(primaryIP *netbox.IPAddress) *string {
	if primaryIP != nil {
		address := primaryIP.Address
		address = strings.Split(address, "/")[0]
		return &address
	}

	return nil
}

func (i *InventorySync) writeSession(session *securecrt.SecureCRTSession) error {
	err := i.scrt.WriteSession(session)
	if err != nil {
		return err
	}

	return nil
}

func (i *InventorySync) getCommonEnvironment(sync_type string) *evaluator.Environment {
	return &evaluator.Environment{
		SessionType:                sync_type,
		Credential:                 i.cfg.Session.SessionOptions.Credential,
		PathTemplate:               i.cfg.Session.Path,
		DeviceNameTemplate:         i.cfg.Session.DeviceName,
		FirewallTemplate:           i.cfg.Session.SessionOptions.Firewall,
		ConnectionProtocolTemplate: i.cfg.Session.SessionOptions.ConnectionProtocol,
	}
}

func (i *InventorySync) getDeviceSessions(devices []netbox.DeviceWithConfigContext, sites []netbox.Site) ([]*securecrt.SecureCRTSession, error) {
	var sessions []*securecrt.SecureCRTSession
	for _, device := range devices {
		site, err := i.getSite(sites, device.Site.Id)
		if err != nil {
			return nil, err
		}

		ipAddress := i.getPrimaryIP(device.PrimaryIp)
		if ipAddress == nil {
			return nil, fmt.Errorf("primary ip is not set on %s", device.Name)
		}

		tenant := i.getTenant(device)
		regionName := i.getRegionName(site)
		siteAddress := strings.ReplaceAll(site.PhysicalAddress, "\r\n", ", ")
		deviceType := device.DeviceType.Display
		siteGroup := ""
		if site.Group != nil {
			siteGroup = site.Group.Slug
		}

		virtualChassisName := ""
		if device.VirtualChassis != nil {
			virtualChassisName = device.VirtualChassis.Name
		}

		env := i.getCommonEnvironment("device")
		env.Device = device
		env.DeviceName = device.Display
		env.DeviceRole = device.Role.Name
		env.DeviceType = deviceType
		env.DeviceIP = *ipAddress
		env.RegionName = regionName
		env.TenantName = tenant
		env.Site = site
		env.SiteName = site.Display
		env.SiteGroup = siteGroup
		env.SiteAddress = siteAddress
		env.VirtualChassisName = virtualChassisName

		err = applyOverrides(i.cfg.Session.Overrides, env)
		if err != nil {
			return nil, err
		}

		path := filepath.Clean(fmt.Sprintf("%s/%s/%s.ini", i.scrt.GetSessionPath(), env.Path, env.DeviceName))
		session := getSessionWithOverrides(path, env)
		sessions = append(sessions, session)
		err = i.writeSession(session)
		if err != nil {
			return nil, err
		}
	}

	return sessions, nil
}

func (i *InventorySync) getVirtualMachineSessions(devices []netbox.VirtualMachineWithConfigContext, sites []netbox.Site) ([]*securecrt.SecureCRTSession, error) {
	var sessions []*securecrt.SecureCRTSession
	for _, device := range devices {
		if device.Site == nil {
			return nil, fmt.Errorf("site is not set on vm: %s", device.Name)
		}

		site, err := i.getSite(sites, device.Site.Id)
		if err != nil {
			return nil, err
		}

		ipAddress := i.getPrimaryIP(device.PrimaryIp)
		if ipAddress == nil {
			return nil, fmt.Errorf("primary ip is not set on %s", device.Name)
		}

		tenant := i.getTenant(device)
		regionName := i.getRegionName(site)
		siteAddress := strings.ReplaceAll(site.PhysicalAddress, "\r\n", ", ")
		deviceType := ""
		if device.Platform != nil {
			deviceType = device.Platform.Display
		}

		siteGroup := ""
		if site.Group != nil {
			siteGroup = site.Group.Slug
		}

		env := i.getCommonEnvironment("virtual_machine")
		env.Device = device
		env.DeviceName = device.Display
		env.DeviceRole = "Virtual Machine"
		env.DeviceType = deviceType
		env.DeviceIP = *ipAddress
		env.RegionName = regionName
		env.TenantName = tenant
		env.Site = site
		env.SiteName = site.Display
		env.SiteGroup = siteGroup
		env.SiteAddress = siteAddress

		err = applyOverrides(i.cfg.Session.Overrides, env)
		if err != nil {
			return nil, err
		}

		path := fmt.Sprintf("%s/%s/%s.ini", i.scrt.GetSessionPath(), env.Path, env.DeviceName)
		session := getSessionWithOverrides(path, env)
		sessions = append(sessions, session)
		err = i.writeSession(session)
		if err != nil {
			return nil, err
		}
	}

	return sessions, nil
}

func (i *InventorySync) runSync() error {
	err := i.nb.TestConnection()
	if err != nil {
		return err
	}

	i.stateLogger(STATE_RUNNING, "Running: Getting sites")
	sites, err := i.nb.GetSites()
	if err != nil {
		return err
	}

	i.stateLogger(STATE_RUNNING, "Running: Getting devices")
	devices, err := i.nb.GetDevices()
	if err != nil {
		return err
	}

	i.stateLogger(STATE_RUNNING, "Running: Getting Virtual Machines")
	vms, err := i.nb.GetVirtualMachines()
	if err != nil {
		return err
	}

	i.stateLogger(STATE_RUNNING, "Running: Writing sessions")
	deviceSessions, err := i.getDeviceSessions(devices, sites)
	if err != nil {
		return err
	}

	vmSessions, err := i.getVirtualMachineSessions(vms, sites)
	if err != nil {
		return err
	}

	i.stateLogger(STATE_RUNNING, "Running: Removing old sessions")
	allSessions := append(deviceSessions, vmSessions...)
	i.scrt.RemoveSessions(allSessions)

	return nil
}

func (i *InventorySync) RunSync() {
	lastSync := time.Now()
	err := i.runSync()
	if err != nil {
		i.stateLogger(STATE_ERROR, err.Error())
	} else {
		i.stateLogger(STATE_DONE, fmt.Sprintf("Status: Last sync @ %s", lastSync.Format("15:04")))
	}
}

func (i *InventorySync) SetupPeriodicSync() {
	for range i.periodicTicker.C {
		if i.cfg.EnablePeriodicSync {
			i.RunSync()
		}
	}
}
