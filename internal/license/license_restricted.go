//go:build !local

// Restricted build (default, used for public distribution):
//  - task-execution APIs return 403 unless the machine is licensed
//  - licenseSalt and allowedMachines are injected at build time via
//    -ldflags "-X cyberstrike-ai/internal/license.licenseSalt=... -X ...allowedMachines=..."
//  - source built WITHOUT these flags refuses every machine: a
//    self-compiled copy from this repository is permanently restricted
//
// License key formats accepted (set in config.yaml -> license_key):
//   HEX32            permanent key  = SHA256(salt|MC|)[32]
//   HEX32@YYYYMMDD   time-limited key = SHA256(salt|MC|YYYYMMDD)[32],
//                    valid through the END of the expiry day (local time).
//                    The expiry date is part of the signed material, so
//                    editing the date in config.yaml invalidates the key.
package license

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

// licenseSalt: build-time secret. Empty in the public source tree, which
// disables all license keys (nobody can derive a valid key without it).
var licenseSalt = ""

// allowedMachines: comma-separated machine codes (normalized, no dashes)
// permanently whitelisted at build time. Empty in the public source tree.
var allowedMachines = ""

// configKey: license key supplied at runtime through config.yaml.
var configKey = ""

// SetConfigKey wires the config.yaml license_key into the check.
func SetConfigKey(k string) { configKey = strings.TrimSpace(k) }

// Restricted reports whether this build currently runs in restricted mode.
func Restricted() bool {
	if licenseSalt == "" && allowedMachines == "" {
		return true // public source build: no license material at all
	}
	if authorized() {
		return false
	}
	if configKey != "" && VerifyKey(configKey) {
		return false
	}
	return true
}

// authorized checks the build-time machine whitelist.
func authorized() bool {
	if allowedMachines == "" {
		return false
	}
	normalized := strings.ToLower(strings.ReplaceAll(MachineCode(), "-", ""))
	for _, h := range strings.Split(allowedMachines, ",") {
		if strings.TrimSpace(strings.ToLower(h)) == normalized {
			return true
		}
	}
	return false
}

// VerifyKey validates a license key supplied through config.yaml
// (license_key). Returns false when no salt was compiled in.
func VerifyKey(key string) bool {
	if licenseSalt == "" {
		return false
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return false
	}
	hexPart, expiry := key, ""
	if i := strings.IndexByte(key, '@'); i >= 0 {
		hexPart, expiry = key[:i], key[i+1:]
		if !expiryValid(expiry) {
			return false // expired or malformed date
		}
	}
	if len(hexPart) != 32 || strings.TrimSpace(hexPart) != hexPart {
		return false
	}
	return strings.EqualFold(hexPart, keyFor(licenseSalt, expiry))
}

// expiryValid reports whether the YYYYMMDD expiry is well-formed and
// still in the future. A license is valid through the whole expiry day.
func expiryValid(e string) bool {
	if len(e) != 8 {
		return false
	}
	t, err := time.ParseInLocation("20060102", e, time.Local)
	if err != nil {
		return false
	}
	return time.Now().Before(t.AddDate(0, 0, 1)) // valid until end of expiry day
}

// keyFor derives the expected key part for this machine under the salt.
// expiry=="" is the permanent form; otherwise expiry is YYYYMMDD and is
// bound into the signature.
func keyFor(salt, expiry string) string {
	sum := sha256.Sum256([]byte(salt + "|" + MachineCode() + "|" + expiry))
	return hex.EncodeToString(sum[:])[:32]
}

// LicenseStatus returns the status payload shown on /api/license/status.
func LicenseStatus() map[string]interface{} {
	return map[string]interface{}{
		"restricted":   Restricted(),
		"build":        "public",
		"machine_code": MachineCode(),
	}
}
