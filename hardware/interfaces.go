package hardware

// CameraDriver defines the interface for camera hardware control.
// Implementations provide both image capture and optional live view streaming.
type CameraDriver interface {
	// Capture takes a photo and saves it to the given filename.
	Capture(filename string) error

	// LiveViewURL returns the URL for the live view JPEG stream.
	// Returns empty string if live view is not available (fallback to WebRTC).
	LiveViewURL() string

	// StartLiveView activates the camera's live view mode.
	// Should be called when the capture screen becomes visible.
	StartLiveView() error

	// StopLiveView deactivates the camera's live view mode.
	// Should be called when leaving the capture screen to prevent sensor overheating.
	StopLiveView() error
}

// PrinterDriver defines the interface for printer hardware control.
type PrinterDriver interface {
	Print(filepath string) error
}
