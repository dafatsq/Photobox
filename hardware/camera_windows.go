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

// Capture triggers a capture via digiCamControl's scripting HTTP API and downloads the result.
func (w *WinCamera) Capture(filename string) error {
	// Get the current file count so we can detect the new image after capture
	countBefore, err0 := w.getFileCount()
	println("[DCC] File count before capture:", countBefore, "err:", fmt.Sprint(err0))

	// Send capture command via dcc's scripting API (slc=capture)
	println("[DCC] Sending slc=capture request...")
	resp, err := w.client.Get(dccBaseURL + "/?slc=capture")
	if err != nil {
		println("[DCC] slc=capture request error:", err.Error())
		return fmt.Errorf("digiCamControl capture request failed: %w", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	println("[DCC] slc=capture status:", resp.StatusCode, "body:", string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("digiCamControl capture failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	// Poll until a new image appears in the file list (up to 15s)
	println("[DCC] Polling for new image...")
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		count, err := w.getFileCount()
		println("[DCC] File count:", count, "err:", fmt.Sprint(err))
		if err == nil && count > countBefore {
			println("[DCC] New image detected! count:", count)
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Download the last captured image from digiCamControl
	println("[DCC] Downloading last capture to:", filename)
	return w.downloadLastCapture(filename)
}

// getFileCount returns the number of files in dcc's current session.
func (w *WinCamera) getFileCount() (int, error) {
	resp, err := w.client.Get(dccBaseURL + "/filelist.json")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var files []dccFileItem
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return 0, err
	}
	return len(files), nil
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
	println("[DCC] Downloading:", dccBaseURL+lastFile.Original)

	// Download the original image via /image/ endpoint
	imgResp, err := w.client.Get(dccBaseURL + lastFile.Original)
	if err != nil {
		return fmt.Errorf("failed to download capture: %w", err)
	}
	defer imgResp.Body.Close()
	println("[DCC] Download status:", imgResp.StatusCode)

	out, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	n, err := io.Copy(out, imgResp.Body)
	println("[DCC] Bytes written:", n)
	if err != nil {
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
