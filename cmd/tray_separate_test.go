//go:build windows

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TinoA/claude-desktop-switcher/internal/profile"
)

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

func TestFolderHoldingGoesByTheWindowsOwnSignIn(t *testing.T) {
	closed, office, private := t.TempDir(), t.TempDir(), t.TempDir()
	for folder, uuid := range map[string]string{closed: "uuid-private", office: "uuid-office", private: "uuid-private"} {
		if err := os.WriteFile(filepath.Join(folder, "config.json"), []byte(`{"lastKnownAccountUuid":"`+uuid+`"}`), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	folders := []string{closed, office, private}
	open := map[string]bool{normalizedFolder(office): true, normalizedFolder(private): true}
	account := profile.AccountUUIDHash("uuid-private")
	if folder, ok := folderHolding(folders, open, account); !ok || folder != private {
		t.Fatalf("folder = %q, %v, want the open window signed in to the account", folder, ok)
	}
	if folder, ok := folderHolding(folders, open, profile.AccountUUIDHash("uuid-other")); ok {
		t.Fatalf("folder = %q for an account no open window holds", folder)
	}
	if folder, ok := folderHolding(folders, open, ""); ok {
		t.Fatalf("folder = %q for an unknown account", folder)
	}
}

func TestSeparateWindowInUsePrompt(t *testing.T) {
	title, message := separateWindowInUsePrompt("private@accounts.test")
	if title != "private@accounts.test is open in a separate window" {
		t.Fatalf("title = %q", title)
	}
	want := "private@accounts.test is already signed in to a separate Claude window. Opening it in the main window too would let both windows change the same sessions.\n\n" +
		"Bring that window to the front?"
	if message != want {
		t.Fatalf("message = %q\nwant %q", message, want)
	}
}

func TestSeparateWindowPrompt(t *testing.T) {
	title, message := separateWindowPrompt("office2@accounts.test")
	if title != "Open office2@accounts.test in a separate window?" {
		t.Fatalf("title = %q", title)
	}
	want := "A second Claude Desktop window opens for office2@accounts.test, next to the main one.\n\n" +
		"The first time it opens signed out: choose “Sign in with code” and sign in as office2@accounts.test. “Continue with Google” cannot finish in a separate window.\n\n" +
		"It uses about 2 GB of memory while it is open."
	if message != want {
		t.Fatalf("message = %q\nwant %q", message, want)
	}
}
