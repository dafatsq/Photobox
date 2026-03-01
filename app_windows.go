//go:build windows

package main

import (
	"photobox/admin"
	"photobox/hardware"
)

// NewPlatformApp creates an App with Windows-specific hardware drivers.
func NewPlatformApp(adminCfg *admin.AdminConfig) *App {
	return NewApp(hardware.NewWinCamera(), &hardware.WinPrinter{}, adminCfg)
}
