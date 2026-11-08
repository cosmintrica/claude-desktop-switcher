//go:build windows

package cmd

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/TinoA/claude-desktop-switcher/internal/bridge"
	"github.com/getlantern/systray"
)

// The tray shows what claude-bridge knows: the Claude Code sessions on the
// bridge, which of them are working, and what waits for them. The row and the
// icon's tooltip appear only while the bridge is installed.

const defaultTooltip = ProductName + " - Claude Desktop account switcher"

func (s *trayState) readBridge() (bridge.State, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return bridge.State{}, err
	}
	return bridge.Read(filepath.Join(home, ".claude-bridge"), filepath.Join(home, ".claude"), time.Now())
}

func (s *trayState) updateBridgeLocked(state bridge.State, err error) {
	if s.bridgeItem == nil {
		return
	}
	message := ""
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		message = "Could not read the session bridge: " + err.Error()
	}
	if message != s.bridgeReadErr {
		s.bridgeReadErr = message
		if message != "" {
			s.setStatus(message)
		}
	}
	if err != nil {
		updateMenuRow(s.bridgeItem, false, "", false)
		s.setTooltipLocked(defaultTooltip)
		return
	}
	summary := "Session bridge: " + bridgeSummary(state)
	updateMenuRow(s.bridgeItem, true, summary, true)
	s.bridgeMenu.update(bridgeEntries(state), false)
	s.setTooltipLocked(ProductName + "\n" + summary)
}

func (s *trayState) setTooltipLocked(text string) {
	if text == s.tooltip {
		return
	}
	s.tooltip = text
	systray.SetTooltip(text)
}

func bridgeSummary(state bridge.State) string {
	if len(state.Sessions) == 0 {
		return "no sessions"
	}
	working := 0
	for _, session := range state.Sessions {
		if session.Busy {
			working++
		}
	}
	text := countText(len(state.Sessions), "session", "sessions")
	if working == 0 {
		return text + ", all idle"
	}
	return text + ", " + strconv.Itoa(working) + " working"
}

// bridgeEntries lists one row per active session, then what waits for them;
// the rows are only read, never clicked.
func bridgeEntries(state bridge.State) []sessionEntry {
	var entries []sessionEntry
	shown := sessionMenuSlots - 4
	for index, session := range state.Sessions {
		if index == shown {
			entries = append(entries, sessionEntry{title: "and " + countText(len(state.Sessions)-shown, "more session", "more sessions")})
			break
		}
		name := session.Name
		if name == "" {
			name = "Untitled session"
		}
		title := menuLabel(shortTitle(name))
		if session.Account != "" {
			title += " (" + menuLabel(plainText(session.Account)) + ")"
		}
		place := "idle"
		if session.Busy {
			place = "working"
		}
		entries = append(entries, sessionEntry{title: title + "\t" + place})
	}
	if len(state.Sessions) == 0 {
		entries = append(entries, sessionEntry{title: "No session is on the bridge right now"})
	}
	if state.Unread > 0 {
		entries = append(entries, sessionEntry{title: countText(state.Unread, "message", "messages") + " waiting to be read"})
	}
	if state.Claims > 0 {
		entries = append(entries, sessionEntry{title: countText(state.Claims, "file or task", "files or tasks") + " claimed"})
	}
	if state.Handoffs > 0 {
		entries = append(entries, sessionEntry{title: countText(state.Handoffs, "handoff", "handoffs") + " waiting to be taken"})
	}
	return entries
}

func countText(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return strconv.Itoa(n) + " " + many
}
