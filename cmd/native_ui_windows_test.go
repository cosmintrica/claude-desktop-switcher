//go:build windows

package cmd

import (
	"strings"
	"testing"
	"time"
)

func TestNativeBackupFileFilterUsesDoubleNULTerminator(t *testing.T) {
	filter := nativeBackupFileFilter()
	if len(filter) < 2 || filter[len(filter)-1] != 0 || filter[len(filter)-2] != 0 {
		t.Fatal("native backup filter is not double-NUL terminated")
	}
}

func TestDefaultBackupFilenameIdentifiesOnlyPasswordBackups(t *testing.T) {
	now := time.Date(2026, 7, 23, 20, 55, 0, 0, time.Local)
	if got := defaultBackupFilename(now, false); got != "claude-desktop-switcher-2026-07-23_2055.csb" {
		t.Fatalf("local filename = %q", got)
	}
	if got := defaultBackupFilename(now, true); got != "claude-desktop-switcher-2026-07-23_2055-pw.csb" {
		t.Fatalf("password filename = %q", got)
	}
}

func TestCloseConfirmationAllowsOnlyOneDialogPerWorkflow(t *testing.T) {
	workflow := &addWorkflow{}
	state := &trayState{workflow: workflow}
	if !state.beginCloseConfirmation(workflow) {
		t.Fatal("first confirmation should start")
	}
	if state.beginCloseConfirmation(workflow) {
		t.Fatal("duplicate confirmation should be rejected")
	}
	state.endCloseConfirmation()
	if !state.beginCloseConfirmation(workflow) {
		t.Fatal("confirmation should be allowed again after closing")
	}
}

func TestAutomaticProfileNameUsesDistinctInternalUUIDs(t *testing.T) {
	first, err := automaticProfileName(func(string) bool { return false })
	if err != nil {
		t.Fatal(err)
	}
	second, err := automaticProfileName(func(string) bool { return false })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(first, "account-") || len(first) != len("account-00000000-0000-0000-0000-000000000000") {
		t.Fatalf("automatic name = %q", first)
	}
	if first == second {
		t.Fatal("automatic profile IDs must be distinct")
	}
}

type fakeHideableMenuItem struct {
	events []string
}

func (f *fakeHideableMenuItem) Disable()        { f.events = append(f.events, "disable") }
func (f *fakeHideableMenuItem) Hide()           { f.events = append(f.events, "hide") }
func (f *fakeHideableMenuItem) Enable()         { f.events = append(f.events, "enable") }
func (f *fakeHideableMenuItem) Show()           { f.events = append(f.events, "show") }
func (f *fakeHideableMenuItem) SetTitle(string) { f.events = append(f.events, "title") }

func TestRetireMenuItemHidesAfterItsFinalUpdate(t *testing.T) {
	item := &fakeHideableMenuItem{}
	retireMenuItem(item)
	if len(item.events) != 2 || item.events[0] != "disable" || item.events[1] != "hide" {
		t.Fatalf("events = %v", item.events)
	}
}

// A title or state change would put a hidden row back in the menu, which
// closes the menu if it is open.
func TestHiddenMenuRowIsOnlyHidden(t *testing.T) {
	item := &fakeHideableMenuItem{}
	updateMenuRow(item, false, "Close Claude Desktop", true)
	if strings.Join(item.events, ",") != "hide" {
		t.Fatalf("hidden row events = %v", item.events)
	}
	item.events = nil
	updateMenuRow(item, true, "Close Claude Desktop", false)
	if got := strings.Join(item.events, ","); got != "title,disable,show" {
		t.Fatalf("shown row events = %v", got)
	}
}

func TestStopMenuWatcherIsIdempotent(t *testing.T) {
	stop := make(chan struct{})
	stops := map[string]chan struct{}{"work": stop}
	stopMenuWatcher(stops, "work")
	stopMenuWatcher(stops, "work")
	select {
	case <-stop:
	default:
		t.Fatal("watcher was not stopped")
	}
}
