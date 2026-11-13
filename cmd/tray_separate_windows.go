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
	"github.com/TinoA/claude-desktop-switcher/internal/sessions"
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

func (s *trayState) closeSeparateWindow(name string) {
	folder, err := separateWindowFolder(name)
	if err != nil {
		return
	}
	if home, err := os.UserHomeDir(); err == nil {
		if working := sessions.Working(filepath.Join(home, ".claude"), folder); len(working) > 0 &&
			!confirmStoppingWork("Close the separate window of "+s.cachedProfileDisplayName(name)+"?", "You can continue these sessions after it opens again.", "Close anyway?", working) {
			return
		}
	}
	s.closeSeparateWindowNow(name, false)
}

// closeSeparateWindowNow closes the window without asking; quiet leaves the
// progress to the aquarium.
func (s *trayState) closeSeparateWindowNow(name string, quiet bool) {
	if s.initializingSnapshot() || s.switchingSnapshot() || s.workflowSnapshot() != nil {
		s.setStatus("Wait for the current operation to finish")
		return
	}
	runner, ok := s.platform.(platform.SeparateWindowRunner)
	if !ok {
		return
	}
	folder, err := separateWindowFolder(name)
	if err != nil {
		return
	}
	label := s.cachedProfileDisplayName(name)
	s.setStatus("Closing the separate window of " + label)
	card := terminalCardUnless(quiet, "claude quit "+label, "closing the window", 0.1)
	err = runner.CloseSeparateWindow(folder)
	card.Close()
	if err != nil {
		s.setStatus("Could not close the separate window: " + err.Error())
		trayWarning("Could not close the separate window", err.Error())
		return
	}
	s.refreshAccounts()
}

func (s *trayState) openSeparateWindow(name string) {
	s.openSeparateWindowWith(name, false)
}

// openSeparateWindowWith opens the account's separate window; quiet leaves the
// progress to the aquarium.
func (s *trayState) openSeparateWindowWith(name string, quiet bool) {
	if s.initializingSnapshot() || s.switchingSnapshot() || s.workflowSnapshot() != nil {
		s.setStatus("Wait for the current operation to finish")
		return
	}
	runner, ok := s.platform.(platform.SeparateWindowRunner)
	if !ok {
		return
	}
	label := s.cachedProfileDisplayName(name)
	running, active := s.claudeStateSnapshot()
	switch separateOpenChoice(running, active == name, s.mainWindowHolds(name)) {
	case refuseInMain:
		trayWarning(ProductName, label+" is open in the main Claude window.")
		return
	case openInMain:
		s.switchTo(name, false)
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
	if open, err := runner.SeparateWindows(); err == nil && !open[normalizedFolder(folder)] {
		if err := s.prepareSeparateWindow(name, folder); err != nil {
			s.setStatus("Could not open a separate window: " + err.Error())
			trayWarning("Could not open a separate window", "The sessions of "+label+" could not be prepared for the window.\n\n"+err.Error())
			return
		}
	}
	s.setStatus("Opening a separate window for " + label)
	card := terminalCardUnless(quiet, "claude window "+label, "starting Claude", 0.2)
	shown := time.Now()
	err = runner.OpenSeparateWindow(folder, "")
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

// Separate windows share the main window's session records, so an account
// signed in to one of them must not open in the main window too: both windows
// would change the same sessions.
func (s *trayState) separateWindowWith(name string) (string, bool) {
	profiles, err := s.store.List()
	if err != nil {
		return "", false
	}
	return s.separateWindowHolding(profiles, profileAccount(profiles, name))
}

func (s *trayState) separateWindowHolding(profiles []profile.Meta, account string) (string, bool) {
	runner, ok := s.platform.(platform.SeparateWindowRunner)
	if !ok || account == "" {
		return "", false
	}
	open, err := runner.SeparateWindows()
	if err != nil || len(open) == 0 {
		return "", false
	}
	var folders []string
	for _, meta := range profiles {
		if folder, err := separateWindowFolder(meta.Name); err == nil {
			folders = append(folders, folder)
		}
	}
	return folderHolding(folders, open, account)
}

// Each window's own sign-in decides which account it holds, not the profile
// the switcher believes is active: that one may be unverified.
func folderHolding(folders []string, open map[string]bool, account string) (string, bool) {
	if account == "" {
		return "", false
	}
	for _, folder := range folders {
		if open[normalizedFolder(folder)] && profile.AccountUUIDHashAt(folder) == account {
			return folder, true
		}
	}
	return "", false
}

type separateOpen int

const (
	openSeparately separateOpen = iota
	refuseInMain
	openInMain
)

// The account the main folder belongs to opens in the main window, where its
// sidebar groups and pins live: a separate window would show it without them.
func separateOpenChoice(running, active, mainHolds bool) separateOpen {
	switch {
	case running && (active || mainHolds):
		return refuseInMain
	case !running && mainHolds:
		return openInMain
	}
	return openSeparately
}

func (s *trayState) mainWindowHolds(name string) bool {
	appData, err := s.platform.AppDataPath()
	if err != nil {
		return false
	}
	profiles, err := s.store.List()
	if err != nil {
		return false
	}
	account := profileAccount(profiles, name)
	return account != "" && profile.AccountUUIDHashAt(appData) == account
}

// separateWindowForMain finds the separate window holding the account Claude
// would open in the main window: the current profile's, else the one signed in
// to the main folder last.
func (s *trayState) separateWindowForMain() (string, string, bool) {
	profiles, err := s.store.List()
	if err != nil {
		return "", "", false
	}
	current, _ := s.store.Current()
	account := profileAccount(profiles, current)
	if account == "" {
		appData, err := s.platform.AppDataPath()
		if err != nil {
			return "", "", false
		}
		account = profile.AccountUUIDHashAt(appData)
	}
	folder, ok := s.separateWindowHolding(profiles, account)
	if !ok {
		return "", "", false
	}
	return folder, s.cachedProfileDisplayName(filepath.Base(folder)), true
}

// Claude refuses to save its records through a link, so a separate window
// keeps a records folder of its own. Before it starts, whatever it saved last
// time goes to the main folder, then its account's sessions come back from
// there with the worktree reservations they hold.
func (s *trayState) prepareSeparateWindow(name, folder string) error {
	appData, err := s.platform.AppDataPath()
	if err != nil {
		return err
	}
	profiles, err := s.store.List()
	if err != nil {
		return err
	}
	account := profile.AccountUUIDHashAt(folder)
	if account == "" {
		account = profileAccount(profiles, name)
	}
	if err := sessions.UnlinkRecords(folder); err != nil {
		return err
	}
	if running, _ := s.claudeStateSnapshot(); !running || account != profile.AccountUUIDHashAt(appData) {
		if _, err := sessions.MirrorAccount(folder, appData, account); err != nil {
			return err
		}
	}
	if _, err := sessions.MirrorAccount(appData, folder, account); err != nil {
		return err
	}
	return sessions.CopyWorktreeLeases(appData, folder, account)
}

// What a separate window saves is copied to the main folder on every refresh,
// so Continue a session lists it and the main window has it after a switch.
// The account signed in to the main window is left to that window while it
// runs; while it is closed, its folder gets that account's work too, so it
// opens with everything.
func (s *trayState) syncSeparateWindows() {
	if _, ok := s.platform.(platform.SeparateWindowRunner); !ok || s.backgroundReadsPaused() {
		return
	}
	appData, err := s.platform.AppDataPath()
	if err != nil {
		return
	}
	profiles, err := s.store.List()
	if err != nil {
		return
	}
	mainAccount := profile.AccountUUIDHashAt(appData)
	mainRunning, _ := s.claudeStateSnapshot()
	message := ""
	for _, meta := range profiles {
		folder, err := separateWindowFolder(meta.Name)
		if err != nil {
			continue
		}
		account := profile.AccountUUIDHashAt(folder)
		if account == "" || (mainRunning && account == mainAccount) {
			continue
		}
		if _, err := sessions.MirrorAccount(folder, appData, account); err != nil && !errors.Is(err, sessions.ErrLinked) {
			message = "Could not copy the sessions of a separate window: " + err.Error()
		}
	}
	s.mu.Lock()
	changed := message != s.separateSyncErr
	s.separateSyncErr = message
	s.mu.Unlock()
	if changed && message != "" {
		s.setStatus(message)
	}
}

func profileAccount(profiles []profile.Meta, name string) string {
	for _, meta := range profiles {
		if meta.Name == name {
			return meta.AccountUUIDHash
		}
	}
	return ""
}

func (s *trayState) offerSeparateWindow(label, folder string) {
	title, message := separateWindowInUsePrompt(label)
	if confirmed, err := nativeTrayConfirm(title, message); err != nil || !confirmed {
		return
	}
	if err := s.platform.(platform.SeparateWindowRunner).OpenSeparateWindow(folder, ""); err != nil {
		s.setStatus("Could not bring the separate window forward: " + err.Error())
		trayWarning("Could not bring the separate window forward", err.Error())
	}
}

func separateWindowInUsePrompt(label string) (string, string) {
	return label + " is open in a separate window",
		label + " is already signed in to a separate Claude window. Opening it in the main window too would let both windows change the same sessions.\n\n" +
			"Bring that window to the front?"
}

func separateWindowPrompt(label string) (string, string) {
	return "Open " + label + " in a separate window?",
		"A second Claude Desktop window opens for " + label + ", next to the main one.\n\n" +
			"The first time it opens signed out: choose “Sign in with code” and sign in as " + label + ". “Continue with Google” cannot finish in a separate window.\n\n" +
			"It uses about 2 GB of memory while it is open."
}
