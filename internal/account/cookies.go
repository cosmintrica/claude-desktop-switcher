//go:build darwin

package account

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const (
	keychainAccount  = "Claude Key"
	keychainService  = "Claude Safe Storage"
	pbkdf2Salt       = "saltysalt"
	pbkdf2Iterations = 1003
	pbkdf2KeyLen     = 16
	chromiumPrefix   = "v10"
)

// CookieManager decrypts Chromium session cookies using the macOS keychain key.
type CookieManager struct {
	key []byte
}

// NewCookieManager reads the AES key from the macOS keychain and returns a
// ready CookieManager. Returns an error if the keychain entry is missing.
func NewCookieManager() (*CookieManager, error) {
	out, err := exec.Command(
		"security", "find-generic-password",
		"-a", keychainAccount, "-s", keychainService, "-w",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("keychain: %w", err)
	}
	password := strings.TrimSpace(string(out))
	key := pbkdf2.Key([]byte(password), []byte(pbkdf2Salt), pbkdf2Iterations, pbkdf2KeyLen, sha1.New)
	return &CookieManager{key: key}, nil
}

// SessionKey reads and decrypts the sessionKey cookie from a Chromium Cookies
// SQLite file.
func (cm *CookieManager) SessionKey(cookiesPath string) (string, error) {
	value, err := cm.Cookie(cookiesPath, "sessionKey")
	if err == nil && value == "" {
		err = fmt.Errorf("sessionKey cookie not found in %s", cookiesPath)
	}
	return value, err
}

// Cookie reads and decrypts one claude.ai cookie, or "" when it is absent.
// Uses the system sqlite3 CLI to avoid adding a CGO dependency.
func (cm *CookieManager) Cookie(cookiesPath, name string) (string, error) {
	out, err := exec.Command(
		"sqlite3", cookiesPath,
		`SELECT hex(encrypted_value) FROM cookies `+
			`WHERE host_key LIKE '%.claude.ai' AND name='`+name+`' LIMIT 1`,
	).Output()
	if err != nil {
		return "", fmt.Errorf("sqlite3: %w", err)
	}
	hexVal := strings.TrimSpace(string(out))
	if hexVal == "" {
		return "", nil
	}
	blob, err := hex.DecodeString(hexVal)
	if err != nil {
		return "", fmt.Errorf("hex decode: %w", err)
	}
	return cm.decrypt(blob)
}

// decrypt decrypts a Chromium AES-128-CBC cookie blob.
// Chromium prepends a "v10" prefix; the IV is 16 space bytes.
func (cm *CookieManager) decrypt(blob []byte) (string, error) {
	if len(blob) < len(chromiumPrefix) || string(blob[:len(chromiumPrefix)]) != chromiumPrefix {
		return string(blob), nil // stored unencrypted
	}
	ct := blob[len(chromiumPrefix):]
	if len(ct) == 0 || len(ct)%aes.BlockSize != 0 {
		return "", fmt.Errorf("invalid ciphertext length: %d", len(ct))
	}
	block, err := aes.NewCipher(cm.key)
	if err != nil {
		return "", err
	}
	iv := bytes.Repeat([]byte{' '}, aes.BlockSize)
	pt := make([]byte, len(ct))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(pt, ct)

	// PKCS7 unpad
	pad := int(pt[len(pt)-1])
	if pad == 0 || pad > aes.BlockSize || pad > len(pt) {
		return "", fmt.Errorf("invalid PKCS7 padding: %d", pad)
	}
	return string(stripDomainPrefix(pt[:len(pt)-pad])), nil
}
