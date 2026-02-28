//go:build windows

package hardware

import "os/exec"

// WinPrinter implements PrinterDriver for Windows using SumatraPDF silent print.
type WinPrinter struct{}

func (w *WinPrinter) Print(filepath string) error {
	cmd := exec.Command("SumatraPDF.exe", "-print-to-default", "-silent", filepath)
	return cmd.Run()
}
