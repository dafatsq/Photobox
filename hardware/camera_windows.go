//go:build windows

package hardware

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	// digiCamControl default web server address
	dccBaseURL = "http://localhost:5513"
	dccTimeout = 30 * time.Second
)

// WinCamera implements CameraDriver for Windows using digiCamControl's HTTP API.
// digiCamControl must be running with its web server enabled (default port 5513).
type WinCamera struct {
	client *http.Client
}

func NewWinCamera() *WinCamera {
	return &WinCamera{
		client: &http.Client{Timeout: dccTimeout},
	}
}

// Capture triggers a capture via digiCamControl's HTTP API and downloads the result.
func (w *WinCamera) Capture(filename string) error {
	// Send capture command
	resp, err := w.client.Get(dccBaseURL + "/?CMD=Capture&param1=" + filename)
	if err != nil {
		return fmt.Errorf("digiCamControl capture request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("digiCamControl capture failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	// Wait briefly for the camera to process and write the file
	time.Sleep(2 * time.Second)

	// Verify the file was created
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// Try alternative: download the last captured image
		return w.downloadLastCapture(filename)
	}

	return nil
}

// downloadLastCapture downloads the most recent capture from digiCamControl.
func (w *WinCamera) downloadLastCapture(filename string) error {
	resp, err := w.client.Get(dccBaseURL + "/last.jpg")
	if err != nil {
		return fmt.Errorf("failed to download last capture: %w", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to write capture data: %w", err)
	}

	return nil
}

// LiveViewURL returns the digiCamControl live view JPEG endpoint.
func (w *WinCamera) LiveViewURL() string {
	return dccBaseURL + "/liveview.jpg"
}
