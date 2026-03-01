//go:build darwin

package main

import (
	"photobox/admin"
	"photobox/hardware"
)

// NewPlatformApp creates an App with macOS-specific hardware drivers.
func NewPlatformApp(adminCfg *admin.AdminConfig) *App {
	return NewApp(&hardware.MacCamera{}, &hardware.MacPrinter{}, adminCfg)
}
