// Package license — machine fingerprinting shared by all build variants.
// The fingerprint is a non-reversible SHA-256 of the Windows MachineGuid
// and hostname, so displaying it to end users is safe.
package license

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

// fingerprint returns raw fingerprint material for the current machine.
func fingerprint() string {
	guid := machineGUID()
	host, _ := os.Hostname()
	return strings.ToLower(guid) + "|" + strings.ToLower(strings.TrimSpace(host))
}

// MachineCode returns the short, human-shareable license code of this
// machine (16 hex chars, formatted NRXX-XXXX-XXXX-XXXX).
func MachineCode() string {
	sum := sha256.Sum256([]byte(fingerprint()))
	raw := hex.EncodeToString(sum[:])[:16]
	raw = strings.ToUpper(raw)
	return fmt.Sprintf("NR%s-%s-%s-%s", raw[0:2], raw[2:6], raw[6:10], raw[10:14])
}

// (The license key derivation lives in the build variants: restricted
// builds support both permanent and time-limited keys.)
