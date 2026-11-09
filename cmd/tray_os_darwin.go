//go:build darwin

package cmd

import (
	_ "embed"
	"os/exec"

	"github.com/getlantern/systray"
)

const (
	// The menu bar has room for the icon only; the title shows progress.
	trayTitle     = ""
	osAccountName = "macOS"
	// Backups without a password rely on Windows account protection.
	localBackupSupported = false
	// The releases carry Windows installers only, so no update is offered.
	releaseDownloads = false
)

//go:embed assets/macos-menubar.png
var trayIcon []byte

func setTrayIcon() { systray.SetIcon(trayIcon) }

func openURL(url string) error {
	return exec.Command("open", url).Run()
}

func openFolder(path string) error {
	return exec.Command("open", path).Run()
}

func setCurrentProcessAppUserModelID() error { return nil }

func (s *trayState) showFirstRunGuideIfPending() {}
