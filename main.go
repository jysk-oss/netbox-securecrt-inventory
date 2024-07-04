package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/jysk-network/netbox-securecrt-inventory/internal/config"
	"github.com/jysk-network/netbox-securecrt-inventory/internal/inventory"
	"github.com/jysk-network/netbox-securecrt-inventory/internal/netbox"
	"github.com/jysk-network/netbox-securecrt-inventory/internal/tray"
	"github.com/jysk-network/netbox-securecrt-inventory/pkg/securecrt"
	"github.com/sqweek/dialog"
)

func main() {
	// make sure our config is valid
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

	// setup logging
	appDataDir, err := os.UserConfigDir()
	if err != nil {
		dialog.Message("Error: %v", err).Title("Config Error").Error()
		return
	}

	logPath := fmt.Sprintf("%s/%s/%s", appDataDir, "securecrt-inventory", "securecrt-inventory.log")
	err = os.MkdirAll(filepath.Dir(logPath), 0755)
	if err != nil {
		dialog.Message("Error: %v", err).Title("Logging Setup Error").Error()
		return
	}

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		dialog.Message("Error: %v", err).Title("Logging Setup Error").Error()
		return
	}

	logLevel := slog.LevelError
	if cfg.LogLevel == "DEBUG" {
		logLevel = slog.LevelDebug
	}
	if cfg.LogLevel == "INFO" {
		logLevel = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(file, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	// setup securecrt config builder, and validate it's installed
	scrt, err := securecrt.New()
	if err != nil {
		slog.Error("Failed to load securecrt config", slog.String("error", err.Error()))
		dialog.Message("Error: %v", err).Title("Config Error").Error()
		return
	}

	// setup the systray, and all menu items
	systray := tray.New(cfg)
	syncCallback := func(state string, message string) {
		if state == inventory.STATE_RUNNING {
			systray.SetSyncButtonStatus(false)
			systray.StartAnimateIcon()
			systray.SetStatus(true)
		} else {
			systray.SetSyncButtonStatus(true)
			systray.StopAnimateIcon()
		}

		if state == inventory.STATE_ERROR {
			systray.SetStatus(false)
		}

		systray.SetStatusMessage(message)
	}

	// setup our netbox client, and the inventory client to combine them all
	nb := netbox.New(cfg.NetboxUrl, cfg.NetboxToken)
	invClient := inventory.New(cfg, nb, scrt, syncCallback)

	// handle periodic sync if enabled
	go invClient.SetupPeriodicSync()

	// handle click events
	go func() {
		for menuItem := range systray.ClickedCh {
			if menuItem == "sync" {
				go func() {
					slog.Info("Running manual sync")
					invClient.RunSync()
				}()
			}

			if menuItem == "open-log" {
				err := openFile(logPath)
				if err != nil {
					slog.Error("failed to open log")
				}
			}

			if menuItem == "quit" {
				systray.Quit()
			}
		}
	}()

	// show the systray in a blocking way
	systray.Run()
}

func openFile(file string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", file).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", file).Start()
	case "darwin":
		err = exec.Command("open", file).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	return err
}
