//go:build darwin

package hardware

import "os/exec"

// MacCamera implements CameraDriver for macOS using gphoto2.
type MacCamera struct{}

func (m *MacCamera) Capture(filename string) error {
	cmd := exec.Command("gphoto2", "--capture-image-and-download", "--filename", filename)
	return cmd.Run()
}
