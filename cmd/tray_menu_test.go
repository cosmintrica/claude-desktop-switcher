//go:build windows

package cmd

import (
	"testing"
	"time"
	"unicode/utf8"

	"github.com/TinoA/claude-desktop-switcher/internal/account"
)

func TestLimitsTextSaysHowMuchIsUsed(t *testing.T) {
	now := time.Date(2026, 11, 8, 1, 15, 0, 0, time.Local)
	soon := time.Date(2026, 11, 8, 4, 49, 0, 0, time.Local)
	monday := time.Date(2026, 11, 9, 23, 59, 0, 0, time.Local)
	tests := []struct {
		usage account.Usage
		want  string
	}{
		{account.Usage{FiveHour: account.Limit{Percent: 99, Resets: soon, Known: true}, Week: account.Limit{Percent: 51, Resets: monday, Known: true}}, "5h 99% used, resets 04:49  ·  week 51% used"},
		{account.Usage{FiveHour: account.Limit{Known: true}, Week: account.Limit{Known: true}}, "5h 0% used  ·  week 0% used"},
		{account.Usage{FiveHour: account.Limit{Known: true}, Week: account.Limit{Percent: 100, Resets: monday, Known: true}}, "5h 0% used  ·  week full until Mon 23:59"},
		{account.Usage{Week: account.Limit{Percent: 100, Known: true}}, "week full"},
	}
	for _, test := range tests {
		if got := limitsText(test.usage, now); got != test.want {
			t.Fatalf("limitsText() = %q, want %q", got, test.want)
		}
	}
}

func TestLimitNumbersSayHowMuchIsUsed(t *testing.T) {
	now := time.Date(2026, 11, 8, 1, 15, 0, 0, time.Local)
	soon := time.Date(2026, 11, 8, 4, 49, 0, 0, time.Local)
	monday := time.Date(2026, 11, 9, 23, 59, 0, 0, time.Local)
	usage := func(fiveHour, week float64) account.Usage {
		return account.Usage{FiveHour: account.Limit{Percent: fiveHour, Resets: soon, Known: true}, Week: account.Limit{Percent: week, Resets: monday, Known: true}}
	}
	tests := []struct {
		usage account.Usage
		want  string
	}{
		{usage(99, 51), "5h \u200799%     week \u200751%"},
		{usage(0, 0), "5h \u2007\u20070%     week \u2007\u20070%"},
		{usage(3, 100), "Mon 23:59     5h \u2007\u20073%     week 100%"},
		{usage(100, 40), "04:49     5h 100%     week \u200740%"},
		{usage(100, 100), "Mon 23:59     5h 100%     week 100%"},
		{account.Usage{Week: account.Limit{Percent: 20, Known: true}}, "week \u200720%"},
		{account.Usage{}, ""},
	}
	for _, test := range tests {
		if got := limitNumbers(test.usage, now); got != test.want {
			t.Fatalf("limitNumbers() = %q, want %q", got, test.want)
		}
	}
}

// Digits and figure spaces are equally wide in the menu font, so a padded
// percent is as wide as a full one and the percents of all accounts line up.
func TestPaddedPercentIsAsWideAsAFullOne(t *testing.T) {
	for _, percent := range []float64{0, 4, 50, 99.6, 100} {
		if got := utf8.RuneCountInString(paddedPercent(percent)); got != len("100%") {
			t.Fatalf("paddedPercent(%v) = %q, want %d characters", percent, paddedPercent(percent), len("100%"))
		}
	}
}

func TestAccountRowTitleSaysWhereTheAccountIsOpen(t *testing.T) {
	tests := []struct {
		name         string
		label        string
		place        accountPlace
		otherAccount bool
		limits       string
		want         string
	}{
		{"closed", "office@accounts.test", placeClosed, false, "", "office@accounts.test"},
		{"with limits", "office@accounts.test", placeClosed, false, "week \u200720%", "office@accounts.test\tweek \u200720%"},
		{"main window", "office@accounts.test", placeMain, false, "", "office@accounts.test (main window)"},
		{"separate window", "office@accounts.test", placeSeparate, false, "", "office@accounts.test (separate window)"},
		{"its window has another account", "office@accounts.test", placeSeparate, true, "", "office@accounts.test (its window has another account)"},
		{"ampersand", "R&D", placeClosed, false, "", "R&&D"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := accountRowTitle(test.label, test.place, test.otherAccount, test.limits); got != test.want {
				t.Fatalf("accountRowTitle() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestAccountActionsFollowWhereTheAccountIsOpen(t *testing.T) {
	tests := []struct {
		place         accountPlace
		first, second string
	}{
		{placeClosed, "Switch to this account", "Open in a separate window"},
		{placeMain, "Show its window", "Close its window"},
		{placeSeparate, "Show its window", "Close its window"},
	}
	for _, test := range tests {
		if first, second := accountActions(test.place); first != test.first || second != test.second {
			t.Fatalf("accountActions(%v) = %q, %q; want %q, %q", test.place, first, second, test.first, test.second)
		}
	}
}

func TestTrayLockRejectsDuplicateImmediately(t *testing.T) {
	t.Setenv("USERPROFILE", t.TempDir())
	first, err := acquireTrayLock()
	if err != nil {
		t.Fatal(err)
	}
	defer first.Release()

	started := time.Now()
	if _, err := acquireTrayLock(); err == nil {
		t.Fatal("second tray lock should be rejected")
	}
	if elapsed := time.Since(started); elapsed > 250*time.Millisecond {
		t.Fatalf("duplicate tray lock took %s; expected immediate rejection", elapsed)
	}
}

func TestBackgroundReadsPauseWhileProfilesChange(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	state := &trayState{}
	if state.backgroundReadsPaused() {
		t.Fatal("background reads paused without an operation")
	}
	state.switching = true
	if !state.backgroundReadsPaused() {
		t.Fatal("background reads continued during a switch")
	}
	state.switching = false
	state.workflow = &addWorkflow{}
	if !state.backgroundReadsPaused() {
		t.Fatal("background reads continued while an account was being added")
	}
	state.workflow = nil
	lock, err := acquireOperationLock("operation")
	if err != nil {
		t.Fatal(err)
	}
	if !state.backgroundReadsPaused() {
		t.Fatal("background reads continued while another operation held the lock")
	}
	lock.Release()
	if state.backgroundReadsPaused() {
		t.Fatal("background reads stayed paused after the operation finished")
	}
}

func TestClaudeMenuState(t *testing.T) {
	tests := []struct {
		name                                   string
		installed, running, busy, separateOpen bool
		wantTitle                              string
		wantEnabled                            bool
	}{
		{"not installed", false, false, false, false, "Claude Desktop: Not installed", false},
		{"closed", true, false, false, false, "Open Claude Desktop", true},
		{"open", true, true, false, false, "Close Claude Desktop", true},
		{"busy while closed", true, false, true, false, "Open Claude Desktop", false},
		{"busy while open", true, true, true, false, "Close Claude Desktop", false},
		{"open beside a separate window", true, true, false, true, "Close the main window", true},
		{"closed beside a separate window", true, false, false, true, "Open Claude Desktop", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			title, enabled := claudeMenuState(test.installed, test.running, test.busy, test.separateOpen)
			if title != test.wantTitle || enabled != test.wantEnabled {
				t.Fatalf("claudeMenuState() = %q, %v; want %q, %v", title, enabled, test.wantTitle, test.wantEnabled)
			}
		})
	}
}
