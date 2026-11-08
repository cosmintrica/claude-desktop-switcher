//go:build windows || darwin

package cmd

import (
	"os"
	"strings"
	"time"
)

// followThisCopy points Start at login at the copy running now, so a build
// with a new name or in another folder is the one that starts next time.
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
		// The system settings can change it while the tray runs.
		if on, _ := startsAtLogin(); on != s.loginItem.Checked() {
			if on {
				s.loginItem.Check()
			} else {
				s.loginItem.Uncheck()
			}
		}
	}
}
