package hardware

// CameraDriver defines the interface for camera hardware control.
type CameraDriver interface {
	Capture(filename string) error
}

// PrinterDriver defines the interface for printer hardware control.
type PrinterDriver interface {
	Print(filepath string) error
}
