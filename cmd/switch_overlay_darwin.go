//go:build darwin

package cmd

import (
	"fmt"
	"sync"

	"github.com/getlantern/systray"
)

// On macOS the progress of a switch shows as the menu bar title beside the
// icon, so it stays visible without a window taking focus. The newest card
// owns the title until it closes.

type switchOverlay struct{ id uint64 }

var menuBarProgress struct {
	sync.Mutex
	current, next uint64
}

func startSwitchOverlay(target string) *switchOverlay {
	return startTerminalCard("claude switch "+target, "closing Claude", 0.04)
}

func startAddPreparationOverlay() *switchOverlay {
	return startTerminalCard("claude account add", "preparing a clean sign-in", 0.1)
}

func startAddCompletionOverlay() *switchOverlay {
	return startTerminalCard("claude account add", "saving the new account", 0.5)
}

func startLoginReopenOverlay() *switchOverlay {
	return startTerminalCard("claude open", "reopening Claude", 0.1)
}

func startAddSuccessOverlay() *switchOverlay {
	return startTerminalCard("claude account add", "account added", 1)
}

func startBackupPreparationOverlay() *switchOverlay {
	return startTerminalCard("claude backup", "preparing the backup", 0.1)
}

func startInitialAccountSaveOverlay() *switchOverlay {
	return startTerminalCard("claude account save", "saving the current account", 0.1)
}

func terminalCardUnless(quiet bool, command, step string, progress float64) *switchOverlay {
	if quiet {
		return &switchOverlay{}
	}
	return startTerminalCard(command, step, progress)
}

func startTerminalCard(_, step string, progress float64) *switchOverlay {
	menuBarProgress.Lock()
	defer menuBarProgress.Unlock()
	menuBarProgress.next++
	overlay := &switchOverlay{id: menuBarProgress.next}
	menuBarProgress.current = overlay.id
	showMenuBarProgress(step, progress)
	return overlay
}

func (o *switchOverlay) Step(step string, progress float64) {
	if o == nil || o.id == 0 {
		return
	}
	menuBarProgress.Lock()
	defer menuBarProgress.Unlock()
	if menuBarProgress.current == o.id {
		showMenuBarProgress(step, progress)
	}
}

func (o *switchOverlay) Close() {
	if o == nil || o.id == 0 {
		return
	}
	menuBarProgress.Lock()
	defer menuBarProgress.Unlock()
	if menuBarProgress.current == o.id {
		menuBarProgress.current = 0
		if trayStarted.Load() {
			systray.SetTitle(trayTitle)
		}
	}
}

func showMenuBarProgress(step string, progress float64) {
	if !trayStarted.Load() {
		return
	}
	systray.SetTitle(fmt.Sprintf("%d%% %s", int(min(max(progress, 0), 1)*100), step))
}
