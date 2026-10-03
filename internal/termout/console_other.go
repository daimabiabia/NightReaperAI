//go:build !windows

package termout

// EnableVTProcessing is a no-op on non-Windows platforms: POSIX terminals
// parse ANSI escape sequences natively.
func EnableVTProcessing() {}
