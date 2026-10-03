package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateStrongPassword returns a URL-safe random password of the given length.
func GenerateStrongPassword(length int) (string, error) {
	if length <= 0 {
		length = 24
	}

	randomBytes := make([]byte, length)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	password := base64.RawURLEncoding.EncodeToString(randomBytes)
	if len(password) > length {
		password = password[:length]
	}
	return password, nil
}

// setupCodeAlphabet excludes visually ambiguous characters (I/O/0/1) so the
// code can be safely copied from a console by hand.
const setupCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// GenerateSetupCode returns a short, unambiguous one-time setup code used to
// gate first-run admin initialization from the web UI.
func GenerateSetupCode(n int) (string, error) {
	if n <= 0 {
		n = 8
	}
	out := make([]byte, n)
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("生成初始化码失败: %w", err)
	}
	max := 256 - (256 % len(setupCodeAlphabet))
	for i := 0; i < n; i++ {
		// Rejection sampling keeps the alphabet uniformly distributed.
		for int(buf[i]) >= max {
			if _, err := rand.Read(buf[i : i+1]); err != nil {
				return "", err
			}
		}
		out[i] = setupCodeAlphabet[int(buf[i])%len(setupCodeAlphabet)]
	}
	return string(out), nil
}
