//go:build darwin

package main

import "photobox/hardware"

// NewPlatformApp creates an App with macOS-specific hardware drivers.
func NewPlatformApp() *App {
	return NewApp(&hardware.MacCamera{}, &hardware.MacPrinter{})
}
