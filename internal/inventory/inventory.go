package inventory

import (
	"fmt"
	"strings"
	"time"

	"github.com/jysk-network/netbox-securecrt-inventory/internal/config"
	"github.com/jysk-network/netbox-securecrt-inventory/internal/netbox"
	"github.com/jysk-network/netbox-securecrt-inventory/internal/tray"
	"github.com/jysk-network/netbox-securecrt-inventory/pkg/securecrt"
	"github.com/jysk-network/netbox-securecrt-inventory/pkg/templater"
	"github.com/netbox-community/go-netbox/v3/netbox/models"
)

type InventorySync struct {
	cfg            *config.Config
	nb             *netbox.NetBox
	scrt           *securecrt.SecureCRT
	systray        *tray.SysTray
	periodicTicker *time.Ticker
	LastSync       time.Time
}

func New(cfg *config.Config, nb *netbox.NetBox, scrt *securecrt.SecureCRT, systray *tray.SysTray) *InventorySync {
	inv := InventorySync{
		cfg:            cfg,
		nb:             nb,
		scrt:           scrt,
		systray:        systray,
		periodicTicker: time.NewTicker(time.Minute * time.Duration(*cfg.PeriodicSyncInterval)),
	}

	go inv.setupPeriodicSync()
	return &inv
}

func (i *InventorySync) setupPeriodicSync() {
	for range i.periodicTicker.C {
		if i.cfg.EnablePeriodicSync {
			err := i.RunSync()
			if err != nil {
				i.systray.SetStatus(false)
				i.systray.SetStatusMessage(err.Error())
			}
		}
	}
}

func (i *InventorySync) getSite(sites []*models.Site, siteID int64) (*models.Site, error) {
	for x := 0; x < len(sites); x++ {
		if sites[x].ID == siteID {
			return sites[x], nil
		}
	}

	return nil, ErrorFailedToFindSite
}

func (i *InventorySync) writeSession(sessionType string, site *models.Site, name, ipAddress, siteAddress, deviceType, siteGroup string, extraVars map[string]string) error {
	sessionData := i.scrt.BuildSessionData(ipAddress, "SSH2", *site.Name, siteAddress, deviceType)

	templateVariables := map[string]string{
		"type":        sessionType,
		"tenant_name": *site.Tenant.Name,
		"region_name": *site.Region.Name,
		"site_name":   *site.Name,
		"device_name": name,
		"site_group":  siteGroup,
	}

	if len(extraVars) > 0 {
		for k, v := range extraVars {
			templateVariables[k] = v
		}
	}

	template := templater.GetTemplate(
		i.cfg.SessionPath.Template,
		i.cfg.SessionPath.Overwrites,
		templateVariables,
	)

	sessionPath := templater.ApplyTemplate(fmt.Sprintf("%s/%s/%s.ini", i.cfg.RootPath, template, name), templateVariables)
	err := i.scrt.WriteSession(sessionPath, sessionData)
	if err != nil {
		return err
	}

	return nil
}

func (i *InventorySync) runSync() error {
	err := i.nb.TestConnection()
	if err != nil {
		return err
	}

	i.LastSync = time.Now()
	i.systray.SetStatusMessage("Running: Getting sites")
	sites, err := i.nb.GetSites()
	if err != nil {
		return err
	}

	i.systray.SetStatusMessage("Running: Getting devices")
	devices, err := i.nb.GetDevices()
	if err != nil {
		return err
	}

	i.systray.SetStatusMessage("Running: Getting Virtual Machines")
	vms, err := i.nb.GetVirtualMachines()
	if err != nil {
		return err
	}

	i.systray.SetStatusMessage("Running: Removing old sessions")
	i.scrt.RemoveSessions(i.cfg.RootPath)

	i.systray.SetStatusMessage("Running: Writing sessions")
	for _, device := range devices {
		site, err := i.getSite(sites, device.Site.ID)
		if err != nil {
			return err
		}

		name := applyNameOverwrites(device.Display, i.cfg.NameOverwrites)
		ipAddress := strings.Split(*device.PrimaryIp4.Address, "/")[0]
		siteAddress := strings.ReplaceAll(site.PhysicalAddress, "\r\n", ", ")
		deviceType := device.DeviceType.Display
		siteGroup := ""
		if site.Group != nil {
			siteGroup = *site.Group.Slug
		}

		err = i.writeSession("device", site, name, ipAddress, siteAddress, deviceType, siteGroup, map[string]string{
			"device_role": *device.DeviceRole.Name,
		})
		if err != nil {
			return err
		}
	}

	for _, device := range vms {
		site, err := i.getSite(sites, device.Site.ID)
		if err != nil {
			return err
		}

		name := applyNameOverwrites(device.Display, i.cfg.NameOverwrites)
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

		err = i.writeSession("virtual_machine", site, name, ipAddress, siteAddress, deviceType, siteGroup, map[string]string{
			"device_role": "Virtual Machine",
		})
		if err != nil {
			return err
		}
	}

	i.systray.SetStatus(true)
	i.systray.SetStatusMessage(fmt.Sprintf("Status: Last sync @ %s", i.LastSync.Format("15:04")))
	return nil
}

func (i *InventorySync) RunSync() error {
	i.systray.SetSyncButtonStatus(false)
	i.systray.StartAnimateIcon()
	err := i.runSync()
	i.systray.SetSyncButtonStatus(true)
	i.systray.StopAnimateIcon()
	return err
}
