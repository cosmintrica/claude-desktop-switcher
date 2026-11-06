//go:build windows

package cmd

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"golang.org/x/sys/windows/registry"
)

// useTestLoginKeys points Start at login at a throwaway key and folder, so
// the tests never touch the real Run key or Startup folder.
func useTestLoginKeys(t *testing.T) {
	t.Helper()
	const parent = `Software\ClaudeDesktopSwitcherTest`
	base := parent + `\` + strconv.FormatInt(time.Now().UnixNano(), 36)
	oldRun, oldApproved, oldShortcut := loginRunKey, loginApprovedKey, loginShortcut
	loginRunKey, loginApprovedKey = base+`\Run`, base+`\StartupApproved`
	shortcut := filepath.Join(t.TempDir(), ProductName+".lnk")
	loginShortcut = func() string { return shortcut }
	t.Cleanup(func() {
		loginRunKey, loginApprovedKey, loginShortcut = oldRun, oldApproved, oldShortcut
		for _, key := range []string{base + `\Run`, base + `\StartupApproved`, base, parent} {
			_ = registry.DeleteKey(registry.CURRENT_USER, key)
		}
	})
}

func TestStartAtLoginWritesAndRemovesTheRunValue(t *testing.T) {
	useTestLoginKeys(t)
	exe := `C:\Tools\claude-desktop-swap.exe`
	if on, _ := startsAtLogin(); on {
		t.Fatal("on before any change")
	}
	if err := setStartAtLogin(true, exe); err != nil {
		t.Fatal(err)
	}
	if on, command := startsAtLogin(); !on || command != `"C:\Tools\claude-desktop-swap.exe" tray` {
		t.Fatalf("on=%v command=%q", on, command)
	}
	for range 2 {
		if err := setStartAtLogin(false, exe); err != nil {
			t.Fatal(err)
		}
		if on, _ := startsAtLogin(); on {
			t.Fatal("still on after turning it off")
		}
	}
}

func TestStartAtLoginHonoursTheSwitchInSettings(t *testing.T) {
	useTestLoginKeys(t)
	exe := `C:\Tools\claude-desktop-swap.exe`
	if err := setStartAtLogin(true, exe); err != nil {
		t.Fatal(err)
	}
	approved, _, err := registry.CreateKey(registry.CURRENT_USER, loginApprovedKey, registry.SET_VALUE)
	if err != nil {
		t.Fatal(err)
	}
	err = approved.SetBinaryValue(ProductName, []byte{3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	approved.Close()
	if err != nil {
		t.Fatal(err)
	}
	if on, _ := startsAtLogin(); on {
		t.Fatal("Settings switched it off")
	}
	if moved, err := followThisCopy(`D:\new\claude-desktop-swap.exe`); moved || err != nil {
		t.Fatalf("a switched-off entry moved: %v %v", moved, err)
	}
	if err := setStartAtLogin(true, exe); err != nil {
		t.Fatal(err)
	}
	if on, _ := startsAtLogin(); !on {
		t.Fatal("turning it on in the menu must undo the switch-off in Settings")
	}
}

func TestStartAtLoginFollowsTheCopyThatRuns(t *testing.T) {
	useTestLoginKeys(t)
	if err := setStartAtLogin(true, `C:\old\claude-desktop-swap-1.exe`); err != nil {
		t.Fatal(err)
	}
	moved, err := followThisCopy(`C:\new\claude-desktop-swap-2.exe`)
	if !moved || err != nil {
		t.Fatalf("moved=%v err=%v", moved, err)
	}
	if _, command := startsAtLogin(); command != `"C:\new\claude-desktop-swap-2.exe" tray` {
		t.Fatalf("command=%q", command)
	}
	if moved, _ := followThisCopy(`C:\NEW\claude-desktop-swap-2.exe`); moved {
		t.Fatal("the same copy moved again")
	}
}

func TestStartAtLoginSeesTheInstallerShortcut(t *testing.T) {
	useTestLoginKeys(t)
	if err := os.WriteFile(loginShortcut(), []byte("shortcut"), 0o600); err != nil {
		t.Fatal(err)
	}
	if on, command := startsAtLogin(); !on || command != "" {
		t.Fatalf("on=%v command=%q", on, command)
	}
	if moved, _ := followThisCopy(`C:\Tools\claude-desktop-swap.exe`); moved {
		t.Fatal("the installer's shortcut is not the menu's to move")
	}
	if err := setStartAtLogin(false, `C:\Tools\claude-desktop-swap.exe`); err != nil {
		t.Fatal(err)
	}
	if on, _ := startsAtLogin(); on {
		t.Fatal("turning it off must remove the installer's shortcut too")
	}
}
