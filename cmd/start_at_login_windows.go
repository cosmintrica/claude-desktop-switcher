//go:build windows

package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/windows/registry"
)

// Start at login is a Run value under the user's key. The installer can also
// have put a shortcut in the Startup folder, and Settings > Apps > Startup can
// switch the Run value off without deleting it.
var (
	loginRunKey      = `Software\Microsoft\Windows\CurrentVersion\Run`
	loginApprovedKey = `Software\Microsoft\Windows\CurrentVersion\Explorer\StartupApproved\Run`
	loginShortcut    = func() string {
		return filepath.Join(os.Getenv("APPDATA"), `Microsoft\Windows\Start Menu\Programs\Startup`, ProductName+".lnk")
	}
)

func loginCommand(exe string) string { return `"` + exe + `" tray` }

// startsAtLogin tells whether Windows starts the switcher at sign-in, with
// the command it runs when that is the Run value.
func startsAtLogin() (bool, string) {
	if _, err := os.Stat(loginShortcut()); err == nil {
		return true, ""
	}
	run, err := registry.OpenKey(registry.CURRENT_USER, loginRunKey, registry.QUERY_VALUE)
	if err != nil {
		return false, ""
	}
	defer run.Close()
	command, _, err := run.GetStringValue(ProductName)
	if err != nil {
		return false, ""
	}
	approved, err := registry.OpenKey(registry.CURRENT_USER, loginApprovedKey, registry.QUERY_VALUE)
	if err != nil {
		return true, command
	}
	defer approved.Close()
	state, _, err := approved.GetBinaryValue(ProductName)
	// Settings marks an entry it switched off with an odd first byte.
	return err != nil || len(state) == 0 || state[0]&1 == 0, command
}

func setStartAtLogin(on bool, exe string) error {
	run, _, err := registry.CreateKey(registry.CURRENT_USER, loginRunKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer run.Close()
	if !on {
		if err := run.DeleteValue(ProductName); err != nil && !errors.Is(err, registry.ErrNotExist) {
			return err
		}
		if err := os.Remove(loginShortcut()); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	}
	if err := run.SetStringValue(ProductName, loginCommand(exe)); err != nil {
		return err
	}
	// A switch-off left in Settings would keep Windows from starting it.
	approved, err := registry.OpenKey(registry.CURRENT_USER, loginApprovedKey, registry.SET_VALUE)
	if err != nil {
		return nil
	}
	defer approved.Close()
	if err := approved.DeleteValue(ProductName); err != nil && !errors.Is(err, registry.ErrNotExist) {
		return err
	}
	return nil
}

// followThisCopy points the Run value at the copy running now, so a build
// with a new name or in another folder is the one Windows starts next time.
func followThisCopy(exe string) (bool, error) {
	on, command := startsAtLogin()
	if !on || command == "" || strings.EqualFold(command, loginCommand(exe)) {
		return false, nil
	}
	return true, setStartAtLogin(true, exe)
}

func (s *trayState) handleStartAtLogin() {
	exe, exeErr := os.Executable()
	if exeErr == nil {
		if moved, err := followThisCopy(exe); err != nil {
			s.setStatus("Could not point Start at login at this copy: " + err.Error())
		} else if moved {
			s.setStatus("Start at login now starts this copy")
		}
	}
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.loginItem.ClickedCh:
			on, _ := startsAtLogin()
			err := exeErr
			if err == nil {
				err = setStartAtLogin(!on, exe)
			}
			switch {
			case err != nil:
				s.setStatus("Could not change Start at login: " + err.Error())
				trayWarning("Could not change Start at login", err.Error())
			case on:
				s.setStatus("Start at login turned off")
			default:
				s.setStatus("Start at login turned on")
			}
		case <-ticker.C:
		}
		// Settings > Apps > Startup can change it while the tray runs.
		if on, _ := startsAtLogin(); on != s.loginItem.Checked() {
			if on {
				s.loginItem.Check()
			} else {
				s.loginItem.Uncheck()
			}
		}
	}
}
