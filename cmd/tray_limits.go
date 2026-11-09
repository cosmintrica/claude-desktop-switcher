//go:build windows || darwin

package cmd

import (
	"math"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/TinoA/claude-desktop-switcher/internal/account"
	"github.com/getlantern/systray"
)

// Account limits are claude.ai's 5-hour and weekly usage of each saved
// account. They are read when the tray starts, after a switch for the account
// it left, and on Refresh limits; only the percentages and reset times are
// kept, in memory.

const limitsWaitForOperations = 2 * time.Minute

// refreshLimits reads the limits of the named accounts, or of all of them.
func (s *trayState) refreshLimits(names ...string) {
	if !s.limitsReading.CompareAndSwap(false, true) {
		return
	}
	defer s.limitsReading.Store(false)
	// Reading a profile's cookies while an operation renames its folder makes
	// Windows refuse the rename.
	for deadline := time.Now().Add(limitsWaitForOperations); s.backgroundReadsPaused(); time.Sleep(time.Second) {
		if time.Now().After(deadline) {
			return
		}
	}
	appData, err := s.platform.AppDataPath()
	if err != nil {
		s.setStatus("Could not read account limits: " + err.Error())
		return
	}
	profiles, err := s.store.List()
	if err != nil {
		s.setStatus("Could not read account limits: " + err.Error())
		return
	}
	paths := make(map[string]string, len(profiles))
	for _, meta := range profiles {
		if len(names) > 0 && !slices.Contains(names, meta.Name) {
			continue
		}
		if path := s.store.ProfileCookiesPath(meta.Name); path != "" {
			paths[meta.Name] = path
		}
	}
	if len(paths) == 0 {
		return
	}
	results := account.FetchUsage(filepath.Join(appData, "Local State"), paths)
	now := time.Now()
	s.mu.Lock()
	if s.limits == nil {
		s.limits = make(map[string]account.Usage)
	}
	for name, result := range results {
		if result.Err == nil && result.Known {
			s.limits[name] = result.Usage
		}
	}
	s.mu.Unlock()
	for name, result := range results {
		label := s.cachedProfileDisplayName(name)
		switch {
		case result.Err != nil:
			s.setStatus("Could not read the limits of " + label + ": " + result.Err.Error())
		case !result.Known:
			s.setStatus("claude.ai shows no limits for " + label)
		default:
			s.setStatus("Limits of " + label + ": " + limitsText(result.Usage, now))
		}
	}
	s.limitsItem.SetTitle("Refresh limits (read at " + now.Format("15:04") + ")")
	s.refreshAccounts()
}

func (s *trayState) handleRefreshLimits(item *systray.MenuItem) {
	for range item.ClickedCh {
		go s.refreshLimits()
	}
}

// switchProfile switches, then reads the limits of the account it left: its
// usage moved while it was in use.
func (s *trayState) switchProfile(name string, card *switchOverlay) error {
	previous, _ := s.store.Current()
	err := switchProfileFromTray(name, s.store, s.platform, card, s.cachedProfileDisplayName)
	if err == nil && previous != "" && previous != name {
		go s.refreshLimits(previous)
	}
	return err
}

type limitWindow struct {
	name  string
	limit account.Limit
}

// limitWindows lists the limits claude.ai reported for an account, as they
// stand now.
func limitWindows(usage account.Usage, now time.Time) []limitWindow {
	var rows []limitWindow
	if limit := usage.FiveHour.At(now); limit.Known {
		rows = append(rows, limitWindow{"5h", limit})
	}
	if limit := usage.Week.At(now); limit.Known {
		rows = append(rows, limitWindow{"week", limit})
	}
	return rows
}

// limitSummary fits both limits on one row of the account's submenu, like
// "5h 10% · week 17%", with the time a full window starts over:
// "week 100% (Tue 00:00)".
func limitSummary(usage account.Usage, now time.Time) string {
	var parts []string
	for _, row := range limitWindows(usage, now) {
		part := row.name + " " + percentText(row.limit.Percent)
		if row.limit.Percent >= 100 && !row.limit.Resets.IsZero() {
			part += " (" + resetTime(row.limit, now) + ")"
		}
		parts = append(parts, part)
	}
	return strings.Join(parts, " · ")
}

// limitsText reads like "5h 52% used  ·  week 11% used", with the time a
// window starts over once it is nearly used up, and "full" once it is.
func limitsText(usage account.Usage, now time.Time) string {
	var parts []string
	if limit := usage.FiveHour.At(now); limit.Known {
		parts = append(parts, "5h "+usedText(limit, now))
	}
	if limit := usage.Week.At(now); limit.Known {
		parts = append(parts, "week "+usedText(limit, now))
	}
	text := ""
	for i, part := range parts {
		if i > 0 {
			text += "  ·  "
		}
		text += part
	}
	return text
}

// usedText names the percent as the used part, so 0% reads as an account
// with room and 100% as one that waits for its reset.
func usedText(limit account.Limit, now time.Time) string {
	resets := !limit.Resets.IsZero()
	switch {
	case limit.Percent >= 100 && resets:
		return "full until " + resetTime(limit, now)
	case limit.Percent >= 100:
		return "full"
	case limit.Percent >= 80 && resets:
		return percentText(limit.Percent) + " used, resets " + resetTime(limit, now)
	}
	return percentText(limit.Percent) + " used"
}

func limitText(limit account.Limit, now time.Time) string {
	text := percentText(limit.Percent)
	if limit.Percent >= 80 && !limit.Resets.IsZero() {
		text += " until " + resetTime(limit, now)
	}
	return text
}

func resetTime(limit account.Limit, now time.Time) string {
	layout := "15:04"
	if limit.Resets.Sub(now) > 20*time.Hour {
		layout = "Mon 15:04"
	}
	return limit.Resets.Local().Format(layout)
}

func percentText(percent float64) string {
	return strconv.Itoa(int(math.Round(percent))) + "%"
}
