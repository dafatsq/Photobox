//go:build windows

package hardware

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// dccBaseURL is the address of the DigiCamControl application's built-in HTTP server.
// DCC must be running and its webserver enabled for this driver to work.
const dccBaseURL = "http://localhost:5513"
const dccCaptureFolder = "session_folder"

// LegacyDCCCamera implements CameraDriver by talking to a running DigiCamControl app.
// This requires DigiCamControl to be installed and running on the system.
type LegacyDCCCamera struct {
	client *http.Client
}

func NewLegacyDCCCamera() *LegacyDCCCamera {
	return &LegacyDCCCamera{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// dccGet is a convenience helper for issuing GET requests to the DCC HTTP API.
func (c *LegacyDCCCamera) dccGet(path string) (string, int, error) {
	resp, err := c.client.Get(dccBaseURL + path)
	if err != nil {
		return "", 0, fmt.Errorf("DCC request failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return strings.TrimSpace(string(body)), resp.StatusCode, nil
}

// Capture triggers a capture via DCC's ?slc=capture API and copies the file to filename.
func (c *LegacyDCCCamera) Capture(filename string) error {
	// 1. Trigger capture
	body, status, err := c.dccGet("/?slc=capture")
	if err != nil {
		return fmt.Errorf("DCC capture trigger failed: %w", err)
	}
	if status != 200 {
		return fmt.Errorf("DCC capture HTTP %d: %s", status, body)
	}

	// 2. Wait briefly for the file to be written
	time.Sleep(2 * time.Second)

	// 3. Get the session folder
	folderBody, _, err := c.dccGet("/?slc=get&param1=session.folder")
	if err != nil {
		return fmt.Errorf("DCC get session folder failed: %w", err)
	}
	sessionFolder := strings.TrimSpace(folderBody)
	if sessionFolder == "" {
		return fmt.Errorf("DCC returned empty session folder")
	}

	// 4. Find the most recently modified image file in the session folder
	srcPath, err := findNewestFile(sessionFolder)
	if err != nil {
		return fmt.Errorf("could not find captured file in %s: %w", sessionFolder, err)
	}

	println("[DCC] Captured file:", srcPath)
	return copyFileLegacy(srcPath, filename)
}

// LiveViewURL returns the DCC live view MJPEG/JPEG stream URL.
func (c *LegacyDCCCamera) LiveViewURL() string {
	// DCC exposes a live view endpoint when enabled
	return dccBaseURL + "/liveviewwebcam.jpg"
}

// StartLiveView tells DCC to show its live view window (activating the stream).
func (c *LegacyDCCCamera) StartLiveView() error {
	_, _, err := c.dccGet("/?CMD=LiveViewWnd_Show")
	if err != nil {
		println("[DCC] StartLiveView error (non-fatal):", err.Error())
	}
	time.Sleep(500 * time.Millisecond)
	return nil
}

// StopLiveView tells DCC to hide its live view window.
func (c *LegacyDCCCamera) StopLiveView() error {
	_, _, err := c.dccGet("/?CMD=LiveViewWnd_Hide")
	if err != nil {
		println("[DCC] StopLiveView error (non-fatal):", err.Error())
	}
	return nil
}

// IsCameraConnected checks if DCC is running and has a camera attached.
func (c *LegacyDCCCamera) IsCameraConnected() bool {
	// Retry for up to 5 seconds to give the camera time to be detected
	// This also ensures the frontend "Checking camera..." screen is visible
	for i := 0; i < 10; i++ {
		if c.checkConnectionSingle() {
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}

func (c *LegacyDCCCamera) checkConnectionSingle() bool {
	resp, err := (&http.Client{Timeout: 2 * time.Second}).Get(dccBaseURL + "/ping")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Connected bool `json:"connected"`
	}
	// Try JSON (our bridge format)
	if err := json.Unmarshal(body, &result); err == nil {
		return result.Connected
	}

	// DCC legacy mode: Check if a camera is actually recognized.
	// When no camera is connected, getting camera-specific properties like "iso"
	// returns an error string like "Object reference not set to an instance of an object."
	// or "Action not allowed" instead of a valid value.
	val, status, err := c.dccGet("/?slc=get&param1=iso")
	if err != nil || status != 200 {
		return false
	}

	valStr := strings.TrimSpace(val)
	if valStr == "" || strings.Contains(valStr, "Object reference") || strings.Contains(valStr, "error") || valStr == "?" {
		return false
	}

	return true
}

// findNewestFile returns the path of the newest file in dir.
func findNewestFile(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	var newest os.FileInfo
	var newestPath string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".raw" && ext != ".cr2" && ext != ".nef" {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if newest == nil || info.ModTime().After(newest.ModTime()) {
			newest = info
			newestPath = filepath.Join(dir, entry.Name())
		}
	}

	if newestPath == "" {
		return "", fmt.Errorf("no image files found in %s", dir)
	}
	return newestPath, nil
}

// copyFileLegacy copies src to dst with retries (in case DCC hasn't finished writing).
func copyFileLegacy(src, dst string) error {
	var in *os.File
	var err error
	for i := 0; i < 10; i++ {
		in, err = os.Open(src)
		if err == nil {
			break
		}
		println("[DCC] File still locked, retrying in 500ms...", err.Error())
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		return fmt.Errorf("failed to open source file after retries: %w", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer out.Close()

	n, err := io.Copy(out, in)
	println("[DCC] Bytes copied:", n)
	return err
}

// suppress unused import warning
var _ = url.QueryEscape
