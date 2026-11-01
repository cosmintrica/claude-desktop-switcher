//go:build windows

package sessions

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/TinoA/claude-desktop-switcher/internal/profile"
)

// An older switcher linked a separate window's records folder to the main one
// with a directory junction; Claude could not save through it.
func TestUnlinkRecordsReplacesALinkWithoutTouchingTheMainFolder(t *testing.T) {
	main, window := t.TempDir(), t.TempDir()
	writeRecordFiles(t, main, accountA, map[string]fileAt{"local_a.json": {"main", time.Unix(1792000000, 0)}})
	link := filepath.Join(window, recordsDirName)
	if output, err := exec.Command("cmd", "/c", "mklink", "/J", link, filepath.Join(main, recordsDirName)).CombinedOutput(); err != nil {
		t.Fatalf("mklink: %v %s", err, output)
	}
	if _, err := MirrorAccount(window, main, profile.AccountUUIDHash(accountA)); !errors.Is(err, ErrLinked) {
		t.Fatalf("mirroring through a link = %v, want ErrLinked", err)
	}

	if err := UnlinkRecords(window); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(link); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("the link is still there: %v", err)
	}
	if got := recordFiles(t, main); got[accountA+"/"+org+"/local_a.json"] != "main" {
		t.Fatalf("the main records changed: %v", got)
	}
	if copied, err := MirrorAccount(main, window, profile.AccountUUIDHash(accountA)); err != nil || copied != 1 {
		t.Fatalf("filling the window = %d, %v", copied, err)
	}
	if got := recordFiles(t, window); got[accountA+"/"+org+"/local_a.json"] != "main" {
		t.Fatalf("window records = %v", got)
	}
}

func TestReadingARecordDoesNotStopClaudeFromDeletingIt(t *testing.T) {
	dir := t.TempDir()
	write := func(name string) string {
		t.Helper()
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(`{"sessionId":"local_a1"}`), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}

	path := write("local_a1.json")
	shared, err := openShared(path)
	if err != nil {
		t.Fatal(err)
	}
	defer shared.Close()
	if err := os.Remove(path); err != nil {
		t.Fatalf("deleting a record while the switcher reads it: %v", err)
	}

	other := write("local_a2.json")
	plain, err := os.Open(other)
	if err != nil {
		t.Fatal(err)
	}
	defer plain.Close()
	if err := os.Remove(other); err == nil {
		t.Fatal("os.Open no longer blocks a delete, so this test proves nothing")
	}
}
