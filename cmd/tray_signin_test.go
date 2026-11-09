//go:build windows

package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestBrowserSignInsFindTheWindowThatSentItsSignInToTheBrowser(t *testing.T) {
	root := t.TempDir()
	write := func(name, text string, appendText bool) {
		t.Helper()
		path := filepath.Join(root, name, "logs", "main.log")
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		flags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
		if appendText {
			flags = os.O_CREATE | os.O_WRONLY | os.O_APPEND
		}
		file, err := os.OpenFile(path, flags, 0o600)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		if _, err := file.WriteString(text); err != nil {
			t.Fatal(err)
		}
	}
	google := "2026-11-09 20:18:55 [info] [Auth] Using system browser for: /login/app-google-auth\n"
	write("Office", google+strings.Repeat("x", 400)+"\n", false)
	write("Private", "2026-11-09 20:18:50 [info] Starting app\n", false)
	offsets := map[string]int64{}
	if names := browserSignIns(root, offsets, true); names != nil {
		t.Fatalf("first read = %v, want nothing: lines already there are old", names)
	}
	write("Private", "2026-11-09 20:19:00 [info] [startup-sign-in] state\n", true)
	if names := browserSignIns(root, offsets, false); names != nil {
		t.Fatalf("names = %v, want nothing", names)
	}
	write("Private", google, true)
	write("New", google, false)
	if names := browserSignIns(root, offsets, false); !reflect.DeepEqual(names, []string{"New", "Private"}) {
		t.Fatalf("names = %v", names)
	}
	write("Office", google, false)
	if names := browserSignIns(root, offsets, false); !reflect.DeepEqual(names, []string{"Office"}) {
		t.Fatalf("names after a new log = %v", names)
	}
	if names := browserSignIns(root, offsets, false); names != nil {
		t.Fatalf("names = %v, want nothing new", names)
	}
}

func TestBrowserSignInWarningSaysTheWayIn(t *testing.T) {
	title, message := browserSignInWarning("office2@accounts.test")
	if title != "Use “Sign in with code” in this window" {
		t.Fatalf("title = %q", title)
	}
	if !strings.Contains(message, "choose “Sign in with code” and sign in as office2@accounts.test.") {
		t.Fatalf("message = %q", message)
	}
}
