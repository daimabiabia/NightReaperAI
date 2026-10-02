//go:build !local

// Restricted build (default, used for public distribution):
//  - task-execution APIs return 403 unless the machine is licensed
//  - licenseSalt and allowedMachines are injected at build time via
//    -ldflags "-X cyberstrike-ai/internal/license.licenseSalt=... -X ...allowedMachines=..."
//  - source built WITHOUT these flags refuses every machine: a
//    self-compiled copy from this repository is permanently restricted
package license

import "strings"

// licenseSalt: build-time secret. Empty in the public source tree, which
// disables all license keys (nobody can derive a valid key without it).
var licenseSalt = ""

// allowedMachines: comma-separated machine fingerprints hashed with the
// salt at build time. Empty in the public source tree.
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

// authorized checks the build-time machine list; the config license key
// path is handled by VerifyKey (called from app.go with the config value).
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
	if licenseSalt == "" || strings.TrimSpace(key) == "" {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(key), keyFor(licenseSalt))
}

// LicenseStatus returns the status payload shown on /api/license/status.
func LicenseStatus() map[string]interface{} {
	return map[string]interface{}{
		"restricted":   Restricted(),
		"build":        "public",
		"machine_code": MachineCode(),
	}
}
