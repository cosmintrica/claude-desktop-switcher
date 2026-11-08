//go:build windows

package cmd

import (
	"time"

	"github.com/TinoA/claude-desktop-switcher/internal/profile"
	"github.com/TinoA/claude-desktop-switcher/internal/sessions"
	"github.com/getlantern/systray"
)

// Every saved account is a row at the top of the menu that says where it is
// open and shows its limits; its submenu holds everything that can be done
// with it. systray keeps rows in the order they were made, so the rows are
// made before the rest of the menu, with spare ones for accounts added later.
const spareAccountRows = 3

type accountPlace int

const (
	placeClosed accountPlace = iota
	placeMain
	placeSeparate
)

type accountRow struct {
	item     *systray.MenuItem
	first    *systray.MenuItem
	second   *systray.MenuItem
	sessions *sessionMenu
	remove   *systray.MenuItem
	name     string
	place    accountPlace
	visible  bool
}

func (s *trayState) newAccountRow() {
	row := &accountRow{item: systray.AddMenuItem("", "")}
	row.first = row.item.AddSubMenuItem("", "")
	row.second = row.item.AddSubMenuItem("", "")
	row.sessions = newSessionMenu(row.item.AddSubMenuItem("Continue a session", "Open a recent Claude Code session of this account"))
	row.remove = row.item.AddSubMenuItem("Remove from Switcher...", "Remove the saved sign-in of this account; the Claude account stays")
	row.item.Hide()
	go s.watchAccountAction(row, row.first, true)
	go s.watchAccountAction(row, row.second, false)
	go s.watchRemoveAccount(row)
	for index := range row.sessions.slots {
		go s.watchSessionSlot(row, index)
	}
	s.accountRows = append(s.accountRows, row)
}

// updateAccountRowsLocked shows a row for each saved account and hides the
// rest. systray puts a hidden row back whenever it changes, so a spare row is
// only hidden, and a row that comes back gets its title first.
func (s *trayState) updateAccountRowsLocked(profiles []profile.Meta, active string, running, installed, busy bool, states map[string]separateWindowState, recent map[string][]sessions.Session, read bool) {
	for len(s.accountRows) < len(profiles) {
		s.newAccountRow()
	}
	now := time.Now()
	enabled := installed && !busy
	s.separateOpen = false
	for index, row := range s.accountRows {
		if index >= len(profiles) {
			if row.visible {
				row.item.Hide()
				row.visible = false
			}
			row.name = ""
			continue
		}
		meta := profiles[index]
		state, known := states[meta.Name]
		if !known {
			state = s.separateShown[meta.Name]
		}
		s.separateShown[meta.Name] = state
		if state.open {
			s.separateOpen = true
		}
		place := placeClosed
		switch {
		case running && active == meta.Name:
			place = placeMain
		case state.open:
			place = placeSeparate
		}
		row.item.SetTitle(accountRowTitle(s.profileLabels[meta.Name], place, state.otherAccount, limitNumbers(s.limits[meta.Name], now)))
		if place == placeMain || place == placeSeparate && !state.otherAccount {
			row.item.Check()
		} else {
			row.item.Uncheck()
		}
		first, second := accountActions(place)
		updateMenuRow(row.first, true, first, enabled)
		updateMenuRow(row.second, true, second, enabled)
		if busy || place == placeMain {
			row.remove.Disable()
		} else {
			row.remove.Enable()
		}
		entries := row.sessions.shown
		switch {
		case read:
			entries = sessionMenuEntries(recent[meta.AccountUUIDHash], now)
		case entries == nil || row.name != meta.Name:
			entries = []sessionEntry{{title: "Reading sessions..."}}
		}
		row.sessions.update(entries, enabled)
		row.name, row.place, row.visible = meta.Name, place, true
	}
	for name := range s.separateShown {
		if !profileListed(profiles, name) {
			delete(s.separateShown, name)
		}
	}
}

func profileListed(profiles []profile.Meta, name string) bool {
	for _, meta := range profiles {
		if meta.Name == name {
			return true
		}
	}
	return false
}

func accountRowTitle(label string, place accountPlace, otherAccount bool, limits string) string {
	title := menuLabel(label)
	switch {
	case place == placeMain:
		title += " (main window)"
	case place == placeSeparate && otherAccount:
		title += " (its window has another account)"
	case place == placeSeparate:
		title += " (separate window)"
	}
	if limits != "" {
		title += "\t" + limits
	}
	return title
}

func accountActions(place accountPlace) (string, string) {
	if place == placeClosed {
		return "Switch to this account", "Open in a separate window"
	}
	return "Show its window", "Close its window"
}

func (s *trayState) disableAccountRowsLocked() {
	for _, row := range s.accountRows {
		row.first.Disable()
		row.second.Disable()
		row.remove.Disable()
		row.sessions.disable()
	}
}

func (s *trayState) accountRowSnapshot(row *accountRow) (string, accountPlace) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return row.name, row.place
}

func (s *trayState) watchAccountAction(row *accountRow, item *systray.MenuItem, first bool) {
	for range item.ClickedCh {
		name, place := s.accountRowSnapshot(row)
		switch {
		case name == "":
		case place == placeClosed && first:
			s.switchTo(name, true)
		case place == placeClosed:
			s.openSeparateWindow(name)
		case first:
			s.showAccountWindow(aquariumAccount{key: name, main: place == placeMain})
		case place == placeSeparate:
			s.closeSeparateWindow(name)
		default:
			if running, _ := s.claudeStateSnapshot(); running {
				s.openOrCloseClaude()
			}
		}
	}
}

func (s *trayState) watchRemoveAccount(row *accountRow) {
	for range row.remove.ClickedCh {
		if name, _ := s.accountRowSnapshot(row); name != "" {
			s.deleteAccount(name, row.remove)
		}
	}
}
