//go:build !windows

package license

// machineGUID falls back to a constant on non-Windows builds; the
// fingerprint still includes the hostname.
func machineGUID() string { return "non-windows" }
