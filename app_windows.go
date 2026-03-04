//go:build windows

package main

import (
	"photobox/admin"
	"photobox/hardware"
)

// NewPlatformApp creates an App with Windows-specific hardware drivers.
// Camera driver is chosen based on the admin config's DSLR mode setting:
//   - "integrated": uses DSLRBridge.exe (our custom bridge, no DCC app needed)
//   - "legacy":     talks directly to a running DigiCamControl app
func NewPlatformApp(adminCfg *admin.AdminConfig) *App {
	var cam hardware.CameraDriver
	if adminCfg.GetDSLRMode() == "legacy" {
		println("[App] DSLR mode: legacy (DigiCamControl app)")
		cam = hardware.NewLegacyDCCCamera()
	} else {
		println("[App] DSLR mode: integrated (DSLRBridge)")
		cam = hardware.NewWinCamera()
	}
	return NewApp(cam, &hardware.WinPrinter{}, adminCfg)
}
