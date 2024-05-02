package tray

import (
	"time"

	"fyne.io/systray"
	"github.com/jysk-network/netbox-securecrt-inventory/internal/config"
	"github.com/jysk-network/netbox-securecrt-inventory/internal/tray/assets"
)

type SysTray struct {
	mStatus         *systray.MenuItem
	mSyncNow        *systray.MenuItem
	mQuit           *systray.MenuItem
	mPeriodicSync   *systray.MenuItem
	cfg             *config.Config
	animationTicker *time.Ticker
	ClickedCh       chan string
}

func New(cfg *config.Config) *SysTray {
	return &SysTray{
		ClickedCh: make(chan string),
		cfg:       cfg,
	}
}

func (s *SysTray) onExit() {
	close(s.ClickedCh)
}

func (s *SysTray) onStartup() {
	systray.SetTemplateIcon(assets.Icon, assets.Icon)
	systray.SetTooltip("Sync devices from NetBox to SecureCRT")

	s.mStatus = systray.AddMenuItem("", "Sync Status")
	s.mStatus.Disable()
	systray.AddSeparator()

	s.mSyncNow = systray.AddMenuItem("Sync Inventory Now", "Start a manual sync now")

	systray.AddSeparator()
	mSettings := systray.AddMenuItem("Settings", "View Settings")
	s.mPeriodicSync = mSettings.AddSubMenuItemCheckbox("Periodic Sync", "Toggle periodic sync on/off", s.cfg.EnablePeriodicSync)

	s.mQuit = systray.AddMenuItem("Quit", "Quit the whole app")

	s.SetStatus(true)
	s.SetStatusMessage("Status: Not synced yet")
	s.togglePeriodicSync()
	s.handleClicks()
}

func (s *SysTray) handleClicks() {
	for {
		select {
		case <-s.mQuit.ClickedCh:
			s.ClickedCh <- "quit"
		case <-s.mSyncNow.ClickedCh:
			s.ClickedCh <- "sync"
		case <-s.mPeriodicSync.ClickedCh:
			s.ClickedCh <- "periodic-sync"
			s.cfg.EnablePeriodicSync = !s.cfg.EnablePeriodicSync
			s.cfg.Save()
			s.togglePeriodicSync()
		}
	}
}

func (s *SysTray) togglePeriodicSync() {
	if s.cfg.EnablePeriodicSync {
		s.mPeriodicSync.SetTitle("Periodic Sync: Enabled")
		s.mPeriodicSync.Check()
	} else {
		s.mPeriodicSync.SetTitle("Periodic Sync: Disabled")
		s.mPeriodicSync.Uncheck()
	}
}

func (s *SysTray) Run() {
	systray.Run(s.onStartup, s.onExit)
}

func (s *SysTray) Quit() {
	if s.animationTicker != nil {
		s.animationTicker.Stop()
	}

	systray.Quit()
}

func (s *SysTray) StartAnimateIcon() {
	imageCount := 6
	currentFrame := 0
	s.animationTicker = time.NewTicker(time.Second / 5)
	icons := [][]byte{
		assets.AnimateIcon1,
		assets.AnimateIcon2,
		assets.AnimateIcon3,
		assets.AnimateIcon4,
		assets.AnimateIcon5,
		assets.AnimateIcon6,
	}

	go func() {
		for range s.animationTicker.C {
			if currentFrame == imageCount-1 {
				currentFrame = 0
			}

			systray.SetTemplateIcon(icons[currentFrame], icons[currentFrame])
			currentFrame = currentFrame + 1
		}
	}()
}

func (s *SysTray) StopAnimateIcon() {
	if s.animationTicker != nil {
		s.animationTicker.Stop()
	}

	systray.SetTemplateIcon(assets.Icon, assets.Icon)
}

func (s *SysTray) SetStatus(status bool) {
	if status {
		s.mStatus.SetIcon(assets.StatusIconGreen)
	} else {
		s.mStatus.SetIcon(assets.StatusIconRed)
	}
}

func (s *SysTray) SetSyncButtonStatus(status bool) {
	if status {
		s.mSyncNow.Enable()
	} else {
		s.mSyncNow.Disable()
	}
}

func (s *SysTray) SetStatusMessage(message string) {
	s.mStatus.SetTitle(message)
}
