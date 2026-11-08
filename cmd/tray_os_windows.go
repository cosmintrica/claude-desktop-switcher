//go:build windows

package cmd

import (
	_ "embed"
	"os/exec"

	"github.com/getlantern/systray"
	"golang.org/x/sys/windows"
)

const (
	trayTitle            = ProductName
	osAccountName        = "Windows"
	localBackupSupported = true
	releaseDownloads     = true
)

//go:embed assets/windows-claude-swap-tray.ico
var trayIcon []byte

func setTrayIcon() { systray.SetIcon(trayIcon) }

func openURL(url string) error {
	verb, err := windows.UTF16PtrFromString("open")
	if err != nil {
		return err
	}
	target, err := windows.UTF16PtrFromString(url)
	if err != nil {
		return err
	}
	return windows.ShellExecute(0, verb, target, nil, nil, windows.SW_SHOWNORMAL)
}

func openFolder(path string) error {
	return exec.Command("explorer.exe", path).Start()
}
