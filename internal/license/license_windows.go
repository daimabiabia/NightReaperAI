//go:build windows

package license

import (
	"golang.org/x/sys/windows/registry"
)

// machineGUID reads the Windows MachineGuid from the registry — a stable
// per-install identifier that survives reboots and user changes.
func machineGUID() string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Cryptography`, registry.QUERY_VALUE|registry.WOW64_64KEY)
	if err != nil {
		return "unknown-guid"
	}
	defer k.Close()
	v, _, err := k.GetStringValue("MachineGuid")
	if err != nil || v == "" {
		return "unknown-guid"
	}
	return v
}
