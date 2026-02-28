//go:build windows

package hardware

import "os/exec"

// WinCamera implements CameraDriver for Windows using digiCamControl CLI.
type WinCamera struct{}

func (w *WinCamera) Capture(filename string) error {
	cmd := exec.Command("CameraControlCmd.exe", "/capture", "/filename", filename)
	return cmd.Run()
}
