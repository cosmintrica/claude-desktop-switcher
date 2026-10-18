//go:build windows

package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/TinoA/claude-desktop-switcher/internal/platform"
	"github.com/TinoA/claude-desktop-switcher/internal/profile"
	"github.com/getlantern/systray"
)

const (
	separateFolderName  = "Claude-Accounts"
	separateStartWait   = 15 * time.Second
	separateCardMinimum = 1500 * time.Millisecond
)

type separateWindowState struct {
	open         bool
	otherAccount bool
}

// Each account's separate window keeps its own sign-in in a folder of its own,
// outside the Store package's redirected Roaming folder.
func separateWindowFolder(name string) (string, error) {
	local, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(local, separateFolderName, name), nil
}

func (s *trayState) readSeparateWindows(profiles []profile.Meta) map[string]separateWindowState {
	runner, ok := s.platform.(platform.SeparateWindowRunner)
	if !ok || s.backgroundReadsPaused() {
		return nil
	}
	open, err := runner.SeparateWindows()
	if err != nil {
		return nil
	}
	states := make(map[string]separateWindowState, len(profiles))
	for _, meta := range profiles {
		folder, err := separateWindowFolder(meta.Name)
		if err != nil {
			continue
		}
		signedIn := profile.AccountUUIDHashAt(folder)
		states[meta.Name] = separateWindowState{
			open:         open[normalizedFolder(folder)],
			otherAccount: signedIn != "" && meta.AccountUUIDHash != "" && signedIn != meta.AccountUUIDHash,
		}
	}
	return states
}

func normalizedFolder(folder string) string {
	if absolute, err := filepath.Abs(folder); err == nil {
		folder = absolute
	}
	return strings.ToLower(filepath.Clean(folder))
}

func (s *trayState) updateSeparateMenuLocked(profiles []profile.Meta, active string, running, enabled bool, states map[string]separateWindowState) {
	if s.separateMenu == nil {
		return
	}
	if _, ok := s.platform.(platform.SeparateWindowRunner); !ok || len(profiles) == 0 {
		s.separateMenu.Hide()
		return
	}
	s.separateMenu.Show()
	seen := make(map[string]bool, len(profiles))
	for _, meta := range profiles {
		seen[meta.Name] = true
		item, ok := s.separateItems[meta.Name]
		if !ok {
			item = s.separateMenu.AddSubMenuItem(menuLabel(s.profileLabels[meta.Name]), "Open this account in a separate Claude window")
			s.separateItems[meta.Name] = item
			stop := make(chan struct{})
			s.separateStops[meta.Name] = stop
			go s.watchSeparateItem(item, meta.Name, stop)
		}
		state, known := states[meta.Name]
		if !known {
			state = s.separateShown[meta.Name]
		}
		s.separateShown[meta.Name] = state
		title, clickable := separateMenuState(s.profileLabels[meta.Name], running && active == meta.Name, state, enabled)
		item.SetTitle(title)
		if clickable {
			item.Enable()
		} else {
			item.Disable()
		}
		item.Show()
	}
	for name := range s.separateItems {
		if !seen[name] {
			s.retireSeparateItemLocked(name)
		}
	}
}

func separateMenuState(label string, inMainWindow bool, state separateWindowState, enabled bool) (string, bool) {
	title := menuLabel(label)
	switch {
	case inMainWindow:
		return title + " (main window)", false
	case state.open:
		title = "✓ " + title + " (open)"
	}
	if state.otherAccount {
		title += " - another account is signed in there"
	}
	return title, enabled
}

func (s *trayState) disableSeparateItemsLocked() {
	for _, item := range s.separateItems {
		item.Disable()
	}
}

func (s *trayState) retireSeparateItemLocked(name string) {
	item, ok := s.separateItems[name]
	if !ok {
		return
	}
	retireMenuItem(item)
	stopMenuWatcher(s.separateStops, name)
	delete(s.separateItems, name)
	delete(s.separateShown, name)
}

func (s *trayState) watchSeparateItem(item *systray.MenuItem, name string, stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			return
		case _, ok := <-item.ClickedCh:
			if !ok {
				return
			}
		}
		s.openSeparateWindow(name)
	}
}

func (s *trayState) openSeparateWindow(name string) {
	if s.initializingSnapshot() || s.switchingSnapshot() || s.workflowSnapshot() != nil {
		s.setStatus("Wait for the current operation to finish")
		return
	}
	runner, ok := s.platform.(platform.SeparateWindowRunner)
	if !ok {
		return
	}
	label := s.cachedProfileDisplayName(name)
	if running, active := s.claudeStateSnapshot(); running && active == name {
		trayWarning(ProductName, label+" is open in the main Claude window.")
		return
	}
	folder, err := separateWindowFolder(name)
	if err != nil {
		trayWarning("Could not open a separate window", err.Error())
		return
	}
	if _, err := os.Stat(folder); errors.Is(err, os.ErrNotExist) {
		title, message := separateWindowPrompt(label)
		if confirmed, err := nativeTrayConfirm(title, message); err != nil || !confirmed {
			return
		}
	}
	s.setStatus("Opening a separate window for " + label)
	card := startTerminalCard("claude window "+label, "starting Claude", 0.2)
	shown := time.Now()
	err = runner.OpenSeparateWindow(folder)
	if err == nil {
		card.Step("waiting for the window", 0.5)
		deadline := time.Now().Add(separateStartWait)
		for time.Now().Before(deadline) {
			if open, readErr := runner.SeparateWindows(); readErr == nil && open[normalizedFolder(folder)] {
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		card.Step("window opened", 1)
	}
	if wait := separateCardMinimum - time.Since(shown); wait > 0 {
		time.Sleep(wait)
	}
	card.Close()
	if err != nil {
		s.setStatus("Could not open a separate window: " + err.Error())
		trayWarning("Could not open a separate window", err.Error())
		return
	}
	s.refreshAccounts()
}

func separateWindowPrompt(label string) (string, string) {
	return "Open " + label + " in a separate window?",
		"A second Claude Desktop window opens for " + label + ", next to the main one.\n\n" +
			"The first time it opens signed out: choose “Sign in with code” and sign in as " + label + ".\n\n" +
			"It uses about 2 GB of memory while it is open."
}
