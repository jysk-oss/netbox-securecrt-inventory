package main

import (
	"github.com/jysk-network/netbox-securecrt-inventory/internal/config"
	"github.com/jysk-network/netbox-securecrt-inventory/internal/inventory"
	"github.com/jysk-network/netbox-securecrt-inventory/internal/netbox"
	"github.com/jysk-network/netbox-securecrt-inventory/internal/tray"
	"github.com/jysk-network/netbox-securecrt-inventory/pkg/securecrt"
	"github.com/sqweek/dialog"
)

func main() {
	cfgPath, err := config.ParseFlags()
	if err != nil {
		dialog.Message("Error: %v", err).Title("Config Error").Error()
		return
	}

	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		dialog.Message("Error: %v", err).Title("Config Error").Error()
		return
	}

	// setup securecrt config builder, and validate it's installed
	scrt, err := securecrt.New(cfg.DefaultCredential)
	if err != nil {
		dialog.Message("Error: %v", err).Title("Credentials Error").Error()
		return
	}

	// setup the systray, and all menu items
	systray := tray.New(cfg)

	// setup our netbox client, and the inventory client to combine them all
	nb := netbox.New(cfg.NetboxUrl, cfg.NetboxToken)
	invClient := inventory.New(cfg, nb, scrt, systray)

	// handle periodic sync if enabled
	go invClient.SetupPeriodicSync()

	// handle click events
	go func() {
		for menuItem := range systray.ClickedCh {
			if menuItem == "sync" {
				go func() {
					err := invClient.RunSync()
					if err != nil {
						systray.SetStatus(false)
						systray.SetStatusMessage(err.Error())
					}
				}()
			}

			if menuItem == "quit" {
				systray.Quit()
			}
		}
	}()

	// show the systray in a blocking way
	systray.Run()
}
