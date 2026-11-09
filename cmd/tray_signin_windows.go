//go:build windows

package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows"
)

// What Claude logs when a window sends its sign-in to the browser. The browser
// hands the answer to the main window, so a separate window cannot finish it.
var browserSignInMarker = []byte("[Auth] Using system browser for: /login/")

const browserSignInReadLimit = 1 << 20

func (s *trayState) watchBrowserSignIns() {
	root, err := separateWindowFolder("")
	if err != nil {
		return
	}
	offsets := map[string]int64{}
	browserSignIns(root, offsets, true)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if s.backgroundReadsPaused() {
			continue
		}
		for _, name := range browserSignIns(root, offsets, false) {
			s.warnBrowserSignIn(name)
		}
	}
}

// browserSignIns reads what each separate window logged since the last call
// and returns the windows that sent their sign-in to the browser. The first
// call only notes where each log ends.
func browserSignIns(root string, offsets map[string]int64, first bool) []string {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		path := filepath.Join(root, name, "logs", "main.log")
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		size := info.Size()
		offset, seen := offsets[name]
		offsets[name] = size
		if first {
			continue
		}
		// Claude starts a new log when the old one grows too big.
		if !seen || size < offset {
			offset = 0
		}
		if size == offset {
			continue
		}
		if size-offset > browserSignInReadLimit {
			offset = size - browserSignInReadLimit
		}
		if chunk, err := readLogRange(path, offset, size-offset); err == nil && bytes.Contains(chunk, browserSignInMarker) {
			names = append(names, name)
		}
	}
	return names
}

// Claude renames its log when it grows too big, which Windows refuses while a
// handle lacks FILE_SHARE_DELETE, and os.Open does not set it.
func readLogRange(path string, offset, length int64) ([]byte, error) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	handle, err := windows.CreateFile(name, windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(handle), path)
	defer file.Close()
	chunk := make([]byte, length)
	n, err := file.ReadAt(chunk, offset)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return chunk[:n], nil
}

func (s *trayState) warnBrowserSignIn(name string) {
	s.mu.Lock()
	if s.signInWarningOpen {
		s.mu.Unlock()
		return
	}
	s.signInWarningOpen = true
	s.mu.Unlock()
	label := s.cachedProfileDisplayName(name)
	s.setStatus("The separate window of " + label + " sent its sign-in to the browser")
	title, message := browserSignInWarning(label)
	go func() {
		trayWarning(title, message)
		s.mu.Lock()
		s.signInWarningOpen = false
		s.mu.Unlock()
	}()
}

func browserSignInWarning(label string) (string, string) {
	return "Use “Sign in with code” in this window",
		"The separate window of " + label + " cannot finish “Continue with Google” or another sign-in through the browser: the browser hands it to the main Claude window, so this window stays signed out.\n\n" +
			"In the window of " + label + ", choose “Sign in with code” and sign in as " + label + "."
}
