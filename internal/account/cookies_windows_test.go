//go:build windows

package account

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

// testLocalState writes a Local State whose cookie key is wrapped with DPAPI
// for the user running the test, the way Chromium stores it.
func testLocalState(t *testing.T, dir string, key []byte) string {
	t.Helper()
	in := windows.DataBlob{Size: uint32(len(key)), Data: &key[0]}
	var out windows.DataBlob
	if err := windows.CryptProtectData(&in, nil, nil, 0, nil, 0, &out); err != nil {
		t.Fatal(err)
	}
	wrapped := append([]byte("DPAPI"), unsafe.Slice(out.Data, out.Size)...)
	windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	state := map[string]any{"os_crypt": map[string]string{"encrypted_key": base64.StdEncoding.EncodeToString(wrapped)}}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "Local State")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// seal encrypts a cookie value as Chromium v130+ does on Windows: the host's
// SHA-256 before the value, AES-256-GCM, a "v10" prefix and the nonce.
func seal(t *testing.T, key []byte, host, value string) []byte {
	t.Helper()
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256([]byte(host))
	plain := append(hash[:], value...)
	return append(append([]byte(windowsCookiePrefix), nonce...), aead.Seal(nil, nonce, plain, nil)...)
}

func testCookies(t *testing.T, dir string, key []byte, rows [][3]any) string {
	t.Helper()
	path := filepath.Join(dir, "Cookies")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE cookies (host_key TEXT, name TEXT, value TEXT DEFAULT '', encrypted_value BLOB, last_access_utc INTEGER)`); err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		host, name, value := row[0].(string), row[1].(string), row[2].(string)
		access := len(value)
		if _, err := db.Exec(`INSERT INTO cookies (host_key, name, encrypted_value, last_access_utc) VALUES (?, ?, ?, ?)`, host, name, seal(t, key, host, value), access); err != nil {
			t.Fatal(err)
		}
	}
	return path
}

func TestWindowsDecryptorOpensTheSavedSignIn(t *testing.T) {
	dir := t.TempDir()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	newest := "session-" + strings.Repeat("n", 24)
	cookies := testCookies(t, dir, key, [][3]any{
		{".claude.ai", "sessionKey", "session-old"},
		{".claude.ai", "sessionKey", newest},
		{".claude.ai", "lastActiveOrg", "org-1"},
		{".example.test", "sessionKey", "session-of-another-site-entirely"},
	})
	d, err := newWindowsDecryptor(testLocalState(t, dir, key))
	if err != nil {
		t.Fatal(err)
	}
	if got, err := d.cookie(cookies, "sessionKey"); err != nil || got != newest {
		t.Fatalf("sessionKey = %q, %v", got, err)
	}
	if got, err := d.cookie(cookies, "lastActiveOrg"); err != nil || got != "org-1" {
		t.Fatalf("lastActiveOrg = %q, %v", got, err)
	}
	if got, err := d.cookie(cookies, "missing"); err != nil || got != "" {
		t.Fatalf("missing = %q, %v", got, err)
	}
}

func TestWindowsDecryptorRefusesAnotherKey(t *testing.T) {
	dir := t.TempDir()
	key, other := bytes.Repeat([]byte{1}, 32), bytes.Repeat([]byte{2}, 32)
	cookies := testCookies(t, dir, key, [][3]any{{".claude.ai", "sessionKey", "session-value"}})
	d, err := newWindowsDecryptor(testLocalState(t, t.TempDir(), other))
	if err != nil {
		t.Fatal(err)
	}
	if got, err := d.cookie(cookies, "sessionKey"); err == nil || got != "" {
		t.Fatalf("opened with the wrong key: %q", got)
	}
}

func TestFetchUsageReportsAMissingLocalState(t *testing.T) {
	results := FetchUsage(filepath.Join(t.TempDir(), "Local State"), map[string]string{"one": "Cookies"})
	if results["one"].Err == nil {
		t.Fatal("no error without Local State")
	}
}
