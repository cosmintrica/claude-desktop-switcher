//go:build darwin

package cmd

import (
	"bytes"
	"encoding/xml"
	"errors"
	"os"
	"path/filepath"
)

// Start at login is a LaunchAgent of the user's that runs the switcher's tray
// once at sign-in.
const loginAgentLabel = "io.github.cosmintrica.claude-desktop-switcher"

var loginAgentPath = func() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", loginAgentLabel+".plist")
}

func loginCommand(exe string) string { return exe }

func startsAtLogin() (bool, string) {
	data, err := os.ReadFile(loginAgentPath())
	if err != nil {
		return false, ""
	}
	var agent struct {
		Arguments []string `xml:"dict>array>string"`
	}
	if err := xml.Unmarshal(data, &agent); err != nil || len(agent.Arguments) == 0 {
		return true, ""
	}
	return true, agent.Arguments[0]
}

func setStartAtLogin(on bool, exe string) error {
	path := loginAgentPath()
	if !on {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var escaped bytes.Buffer
	if err := xml.EscapeText(&escaped, []byte(exe)); err != nil {
		return err
	}
	agent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>` + loginAgentLabel + `</string>
	<key>ProgramArguments</key>
	<array>
		<string>` + escaped.String() + `</string>
		<string>tray</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>ProcessType</key>
	<string>Interactive</string>
	<key>LimitLoadToSessionType</key>
	<string>Aqua</string>
</dict>
</plist>
`
	return os.WriteFile(path, []byte(agent), 0o644)
}
