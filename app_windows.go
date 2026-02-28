//go:build windows

package main

import "photobox/hardware"

// NewPlatformApp creates an App with Windows-specific hardware drivers.
func NewPlatformApp() *App {
	return NewApp(&hardware.WinCamera{}, &hardware.WinPrinter{})
}
