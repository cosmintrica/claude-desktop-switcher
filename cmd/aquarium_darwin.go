//go:build darwin

package cmd

import "github.com/TinoA/claude-desktop-switcher/internal/profile"

// The aquarium draws above the Windows taskbar clock; macOS has no such spot
// yet, so its menu item stays hidden and these keep the tray's calls valid.

type aquariumWindow struct{}

type aquariumAccount struct {
	key  string
	main bool
}

func startAquarium(*trayState, bool) *aquariumWindow { return nil }

func (*aquariumWindow) setAccounts([]aquariumAccount) {}

func aquariumWanted() bool { return false }

func (s *trayState) aquariumAccountsLocked([]profile.Meta, string, bool) []aquariumAccount {
	return nil
}

func (s *trayState) handleAquariumItem() {
	s.aquariumItem.Hide()
}
