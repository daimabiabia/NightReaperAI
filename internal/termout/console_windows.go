//go:build windows

package termout

import (
	"os"

	"golang.org/x/sys/windows"
)

// EnableVTProcessing enables ANSI escape-sequence parsing on the process
// console. A double-click launch attaches a real conhost that has
// virtual-terminal processing OFF by default, so color codes print as
// literal garbage like "[1m[36m". Call once at startup, before the banner.
// Piped/redirected handles have no console mode: they are skipped, and the
// escape codes simply flow into the pipe as before.
func EnableVTProcessing() {
	for _, f := range []*os.File{os.Stdout, os.Stderr} {
		h := windows.Handle(f.Fd())
		var mode uint32
		if err := windows.GetConsoleMode(h, &mode); err != nil {
			continue // not a console (piped/redirected): nothing to enable
		}
		_ = windows.SetConsoleMode(h, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	}
}
