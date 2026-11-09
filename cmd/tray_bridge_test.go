//go:build windows || darwin

package cmd

import (
	"reflect"
	"strings"
	"testing"

	"github.com/TinoA/claude-desktop-switcher/internal/bridge"
)

func TestBridgeSummaryCountsTheWorkingSessions(t *testing.T) {
	tests := []struct {
		sessions []bridge.Session
		want     string
	}{
		{nil, "no sessions"},
		{[]bridge.Session{{Name: "One"}}, "1 session, all idle"},
		{[]bridge.Session{{Busy: true}, {}, {Busy: true}}, "3 sessions, 2 working"},
	}
	for _, test := range tests {
		if got := bridgeSummary(bridge.State{Sessions: test.sessions}); got != test.want {
			t.Fatalf("bridgeSummary() = %q, want %q", got, test.want)
		}
	}
}

func TestBridgeEntriesListTheSessionsThenWhatWaits(t *testing.T) {
	state := bridge.State{
		Sessions: []bridge.Session{{Name: "Site design & mockups", Account: "office@", Busy: true}, {Account: "private@"}},
		Unread:   1,
		Claims:   2,
		Handoffs: 1,
	}
	want := []string{
		"Site design && mockups (office@)\tworking",
		"Untitled session (private@)\tidle",
		"1 message waiting to be read",
		"2 files or tasks claimed",
		"1 handoff waiting to be taken",
	}
	if got, _ := entryTitlesAndIDs(bridgeEntries(state)); !reflect.DeepEqual(got, want) {
		t.Fatalf("entries = %q\nwant %q", got, want)
	}
	if got, _ := entryTitlesAndIDs(bridgeEntries(bridge.State{})); !reflect.DeepEqual(got, []string{"No session is on the bridge right now"}) {
		t.Fatalf("empty bridge entries = %q", got)
	}
}

func TestBridgeEntriesFitTheMenuSlots(t *testing.T) {
	state := bridge.State{Sessions: make([]bridge.Session, 25), Unread: 3, Claims: 1, Handoffs: 2}
	entries, ids := entryTitlesAndIDs(bridgeEntries(state))
	if len(entries) > sessionMenuSlots {
		t.Fatalf("%d entries for %d slots", len(entries), sessionMenuSlots)
	}
	if more := entries[sessionMenuSlots-4]; more != "and 9 more sessions" {
		t.Fatalf("row after the shown sessions = %q", more)
	}
	for _, id := range ids {
		if id != "" {
			t.Fatal("bridge rows are not clickable")
		}
	}
	if last := entries[len(entries)-1]; !strings.HasPrefix(last, "2 handoffs") {
		t.Fatalf("last row = %q", last)
	}
}
