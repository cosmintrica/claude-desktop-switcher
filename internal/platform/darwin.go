//go:build darwin

package platform

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	appName              = "Claude"
	appDataDir           = "Application Support"
	killPollInterval     = 100 * time.Millisecond
	killMaxPolls         = 50 // 50 × 100ms = 5 second timeout before SIGKILL
	killForceDelay       = 300 * time.Millisecond
	claudeProcessPattern = `/Applications/Claude.app/Contents/MacOS/Claude|Claude Helper`
	claudeMainPattern    = `^/Applications/Claude\.app/Contents/MacOS/Claude`
	claudeBundleID       = "com.anthropic.claudefordesktop"
)

var errProcessAbsent = errors.New("process absent")

type darwinPlatform struct{}

func current() Platform { return &darwinPlatform{} }

func (d *darwinPlatform) AppDataPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", appDataDir, appName), nil
}

func (d *darwinPlatform) IsInstalled() bool {
	_, err := os.Stat("/Applications/Claude.app")
	return err == nil
}

func (d *darwinPlatform) WaitForLoginWindow(ctx context.Context) error { return nil }

// pgrep and pkill skip their own ancestors unless given -a, and a switcher
// started from a terminal inside Claude has Claude among them.
func (d *darwinPlatform) IsRunning() (bool, error) {
	err := runProcessCommand("pgrep", "-a", "-f", claudeProcessPattern)
	if errors.Is(err, errProcessAbsent) {
		return false, nil
	}
	return err == nil, err
}

func (d *darwinPlatform) KillApp() error {
	return stopClaudeProcesses(runProcessCommand, time.Sleep, killMaxPolls)
}

func (d *darwinPlatform) LaunchApp() error {
	// Claude Code ships its own claude.app, so the bundle ID picks the desktop app.
	return exec.Command("open", "-b", claudeBundleID).Run()
}

func runProcessCommand(name string, args ...string) error {
	err := exec.Command(name, args...).Run()
	if _, ok := err.(*exec.ExitError); ok {
		return errProcessAbsent
	}
	return err
}

func stopClaudeProcesses(run func(string, ...string) error, sleep func(time.Duration), polls int) error {
	if err := run("pgrep", "-a", "-f", claudeProcessPattern); errors.Is(err, errProcessAbsent) {
		return nil
	} else if err != nil {
		return err
	}
	// Only the main process gets SIGTERM: Electron then quits as with Cmd+Q, and
	// its network service writes the cookies out before the helpers stop.
	_ = run("pkill", "-a", "-TERM", "-f", claudeMainPattern)
	for range polls {
		sleep(killPollInterval)
		if err := run("pgrep", "-a", "-f", claudeProcessPattern); errors.Is(err, errProcessAbsent) {
			return nil
		}
	}
	_ = run("pkill", "-a", "-KILL", "-f", claudeProcessPattern)
	sleep(killForceDelay)
	if err := run("pgrep", "-a", "-f", claudeProcessPattern); errors.Is(err, errProcessAbsent) {
		return nil
	}
	return fmt.Errorf("processes for Claude Desktop remain after forced termination")
}
