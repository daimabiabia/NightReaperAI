//go:build local

// Full-unlock build (-tags local): operator's own machine, no restrictions.
package license

// Restricted always reports false — every feature is available.
func Restricted() bool { return false }

// VerifyKey is unused in the local build but kept for API compatibility.
func VerifyKey(_ string) bool { return true }

// SetConfigKey is a no-op in the local build (everything is unlocked).
func SetConfigKey(_ string) {}

// LicenseStatus returns the status payload shown on /api/license/status.
func LicenseStatus() map[string]interface{} {
	return map[string]interface{}{
		"restricted":   false,
		"build":        "local",
		"machine_code": MachineCode(),
	}
}
