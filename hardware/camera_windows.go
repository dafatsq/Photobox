//go:build windows

package hardware

import (
	"encoding/json"
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
	// Send capture command via dcc web server API
	resp, err := w.client.Get(dccBaseURL + "/?CMD=Capture")
	if err != nil {
		return fmt.Errorf("digiCamControl capture request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("digiCamControl capture failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	// Wait for the camera to process and write the file
	time.Sleep(3 * time.Second)

	// Download the last captured image from digiCamControl
	return w.downloadLastCapture(filename)
}

// dccFileItem represents an entry in dcc's /filelist.json response.
type dccFileItem struct {
	FileName   string `json:"FileName"`
	Original   string `json:"Original"`
	LargeThumb string `json:"LargeThumb"`
	Name       string `json:"Name"`
}

// downloadLastCapture fetches the file list from dcc and downloads the most recent image.
func (w *WinCamera) downloadLastCapture(filename string) error {
	// Get the file list from dcc
	listResp, err := w.client.Get(dccBaseURL + "/filelist.json")
	if err != nil {
		return fmt.Errorf("failed to get file list from digiCamControl: %w", err)
	}
	defer listResp.Body.Close()

	var files []dccFileItem
	if err := json.NewDecoder(listResp.Body).Decode(&files); err != nil {
		return fmt.Errorf("failed to parse file list: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no images found in digiCamControl session")
	}

	// Get the last (most recent) image
	lastFile := files[len(files)-1]

	// Download the original image via /image/ endpoint
	imgResp, err := w.client.Get(dccBaseURL + lastFile.Original)
	if err != nil {
		return fmt.Errorf("failed to download capture: %w", err)
	}
	defer imgResp.Body.Close()

	out, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, imgResp.Body); err != nil {
		return fmt.Errorf("failed to write capture data: %w", err)
	}

	return nil
}

// LiveViewURL returns the digiCamControl live view JPEG endpoint.
// It first pings the server to ensure it is running.
func (w *WinCamera) LiveViewURL() string {
	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get(dccBaseURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return "" // Return empty to trigger WebRTC fallback in frontend
	}
	resp.Body.Close()
	return dccBaseURL + "/liveviewwebcam.jpg"
}
