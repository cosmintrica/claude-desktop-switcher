//go:build windows

package cmd

import (
	"path/filepath"
	"testing"
)

func TestSeparateMenuState(t *testing.T) {
	tests := []struct {
		name          string
		inMainWindow  bool
		state         separateWindowState
		enabled       bool
		wantTitle     string
		wantClickable bool
	}{
		{"in the main window", true, separateWindowState{}, true, "claude@accounts.test (main window)", false},
		{"closed", false, separateWindowState{}, true, "claude@accounts.test", true},
		{"open", false, separateWindowState{open: true}, true, "✓ claude@accounts.test (open)", true},
		{"busy", false, separateWindowState{open: true}, false, "✓ claude@accounts.test (open)", false},
		{"other account", false, separateWindowState{open: true, otherAccount: true}, true, "✓ claude@accounts.test (open) - another account is signed in there", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			title, clickable := separateMenuState("claude@accounts.test", test.inMainWindow, test.state, test.enabled)
			if title != test.wantTitle || clickable != test.wantClickable {
				t.Fatalf("separateMenuState() = %q, %v; want %q, %v", title, clickable, test.wantTitle, test.wantClickable)
			}
		})
	}
}

func TestSeparateWindowFolderSitsOutsideTheStorePackage(t *testing.T) {
	local := t.TempDir()
	t.Setenv("LOCALAPPDATA", local)
	folder, err := separateWindowFolder("Main-account")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(local, "Claude-Accounts", "Main-account"); folder != want {
		t.Fatalf("folder = %q, want %q", folder, want)
	}
}

func TestSeparateWindowPrompt(t *testing.T) {
	title, message := separateWindowPrompt("claude1@accounts.test")
	if title != "Open claude1@accounts.test in a separate window?" {
		t.Fatalf("title = %q", title)
	}
	want := "A second Claude Desktop window opens for claude1@accounts.test, next to the main one.\n\n" +
		"The first time it opens signed out: choose “Sign in with code” and sign in as claude1@accounts.test.\n\n" +
		"It uses about 2 GB of memory while it is open."
	if message != want {
		t.Fatalf("message = %q\nwant %q", message, want)
	}
}
