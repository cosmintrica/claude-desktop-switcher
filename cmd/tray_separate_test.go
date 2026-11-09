//go:build windows || darwin

package cmd

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestSeparateWindowFolderSitsOutsideTheStorePackage(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("separate windows run on Windows only")
	}
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
		"The first time it opens signed out: choose “Sign in with code” and sign in as office2@accounts.test.\n\n" +
		"It uses about 2 GB of memory while it is open."
	if message != want {
		t.Fatalf("message = %q\nwant %q", message, want)
	}
}
