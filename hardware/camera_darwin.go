//go:build darwin

package hardware

import "os/exec"

// MacCamera implements CameraDriver for macOS using gphoto2.
// Live view is not available via gphoto2 — the frontend will
// fall back to WebRTC (webcam/capture card) when LiveViewURL is empty.
type MacCamera struct{}

func (m *MacCamera) Capture(filename string) error {
	cmd := exec.Command("gphoto2", "--capture-image-and-download", "--filename", filename)
	return cmd.Run()
}

// LiveViewURL returns empty string on macOS (no digiCamControl).
// The frontend falls back to WebRTC for live preview.
func (m *MacCamera) LiveViewURL() string {
	return ""
}

// StartLiveView is a no-op on macOS (WebRTC handles preview).
func (m *MacCamera) StartLiveView() error { return nil }

// StopLiveView is a no-op on macOS.
func (m *MacCamera) StopLiveView() error { return nil }
