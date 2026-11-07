//go:build windows

package account

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	_ "modernc.org/sqlite"
)

const windowsCookiePrefix = "v10"

// windowsDecryptor opens Chromium's v10 cookies on Windows: AES-256-GCM under
// the key that Local State keeps wrapped with DPAPI for this Windows user.
// The key and the values it opens stay in memory.
type windowsDecryptor struct {
	aead cipher.AEAD
}

func newWindowsDecryptor(localState string) (*windowsDecryptor, error) {
	data, err := os.ReadFile(localState)
	if err != nil {
		return nil, err
	}
	var state struct {
		OSCrypt struct {
			EncryptedKey string `json:"encrypted_key"`
		} `json:"os_crypt"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, errors.New("Claude's Local State cannot be read")
	}
	wrapped, err := base64.StdEncoding.DecodeString(state.OSCrypt.EncryptedKey)
	if err != nil || !bytes.HasPrefix(wrapped, []byte("DPAPI")) {
		return nil, errors.New("Claude's Local State has no cookie key")
	}
	key, err := dpapiUnprotect(wrapped[len("DPAPI"):])
	if err != nil {
		return nil, err
	}
	defer clear(key)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &windowsDecryptor{aead: aead}, nil
}

func dpapiUnprotect(blob []byte) ([]byte, error) {
	if len(blob) == 0 {
		return nil, errors.New("the cookie key is empty")
	}
	in := windows.DataBlob{Size: uint32(len(blob)), Data: &blob[0]}
	var out windows.DataBlob
	if err := windows.CryptUnprotectData(&in, nil, nil, 0, nil, 0, &out); err != nil {
		return nil, err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	return bytes.Clone(unsafe.Slice(out.Data, out.Size)), nil
}

// cookie reads one claude.ai cookie from a Cookies database copy; an empty
// value means the copy has none.
func (d *windowsDecryptor) cookie(cookiesPath, name string) (string, error) {
	path := filepath.ToSlash(cookiesPath)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	dsn := &url.URL{Scheme: "file", Path: path}
	query := dsn.Query()
	query.Set("mode", "ro")
	query.Add("_pragma", "busy_timeout(2000)")
	dsn.RawQuery = query.Encode()
	db, err := sql.Open("sqlite", dsn.String())
	if err != nil {
		return "", err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var blob []byte
	err = db.QueryRowContext(ctx,
		`SELECT encrypted_value FROM cookies WHERE host_key LIKE '%claude.ai' AND name = ? ORDER BY last_access_utc DESC LIMIT 1`,
		name).Scan(&blob)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", errors.New("the saved cookies cannot be read")
	}
	return d.open(blob)
}

func (d *windowsDecryptor) open(blob []byte) (string, error) {
	if !bytes.HasPrefix(blob, []byte(windowsCookiePrefix)) {
		return "", errors.New("the saved sign-in is in an unknown format")
	}
	sealed := blob[len(windowsCookiePrefix):]
	size := d.aead.NonceSize()
	if len(sealed) < size+d.aead.Overhead() {
		return "", errors.New("the saved sign-in is cut short")
	}
	plain, err := d.aead.Open(nil, sealed[:size], sealed[size:], nil)
	if err != nil {
		return "", errors.New("the saved sign-in does not open with Claude's current key")
	}
	return string(stripDomainPrefix(plain)), nil
}
