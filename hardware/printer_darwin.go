//go:build darwin

package hardware

import "os/exec"

// MacPrinter implements PrinterDriver for macOS using CUPS lp command.
type MacPrinter struct{}

func (m *MacPrinter) Print(filepath string) error {
	cmd := exec.Command("lp", filepath)
	return cmd.Run()
}
