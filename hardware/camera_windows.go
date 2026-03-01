//go:build windows

package hardware

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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
	// Get the DCC session folder and count existing .jpg files before capture
	sessionFolder, err := w.getDCCSessionFolder()
	if err != nil {
		println("[DCC] Could not get session folder:", err.Error(), "— trying filelist fallback")
		sessionFolder = ""
	}
	println("[DCC] Session folder:", sessionFolder)

	// Count existing files before capture (for comparison)
	filesBefore := w.listJPEGs(sessionFolder)
	println("[DCC] Files before capture:", len(filesBefore))

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

	// Poll until a new .jpg file appears in the session folder (up to 20s)
	println("[DCC] Polling for new image in folder:", sessionFolder)
	deadline := time.Now().Add(20 * time.Second)
	var newestFile string
	for time.Now().Before(deadline) {
		filesAfter := w.listJPEGs(sessionFolder)
		println("[DCC] Files after capture:", len(filesAfter))
		// Look for a file not present before
		for _, f := range filesAfter {
			found := false
			for _, bf := range filesBefore {
				if f == bf {
					found = true
					break
				}
			}
			if !found {
				newestFile = f
				println("[DCC] New file detected:", newestFile)
				break
			}
		}
		if newestFile != "" {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if newestFile == "" {
		// Fallback: try filelist.json as last resort
		println("[DCC] No new filesystem file found, trying filelist.json fallback...")
		return w.downloadLastCapture(filename)
	}

	println("[DCC] Copying", newestFile, "to", filename)
	return w.copyFile(newestFile, filename)
}

// getDCCSessionFolder queries DCC for the current session's save folder.
func (w *WinCamera) getDCCSessionFolder() (string, error) {
	resp, err := w.client.Get(dccBaseURL + "/?slc=get&param1=session.folder")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	folder := strings.TrimSpace(string(b))
	if folder == "" || folder == "null" {
		return "", fmt.Errorf("DCC returned empty session folder")
	}
	return folder, nil
}

// listJPEGs returns all .jpg files in the given directory (non-recursive).
func (w *WinCamera) listJPEGs(dir string) []string {
	if dir == "" {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".jpg") {
			files = append(files, dir+"\\"+e.Name())
		}
	}
	return files
}

// copyFile copies src to dst, retrying if the source is still locked by DCC.
func (w *WinCamera) copyFile(src, dst string) error {
	var in *os.File
	var err error
	// Retry up to 10 times (5 seconds) in case DCC still has the file open
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

// getFileCount returns the number of files in dcc's current session.
// An empty body (EOF) is treated as 0 files rather than an error.
func (w *WinCamera) getFileCount() (int, error) {
	resp, err := w.client.Get(dccBaseURL + "/filelist.json")
	if err != nil {
		return 0, err
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return 0, err
	}
	// DCC returns an empty body when the session has no images yet — treat as 0
	body = []byte(strings.TrimSpace(string(body)))
	if len(body) == 0 || string(body) == "null" {
		return 0, nil
	}
	var files []dccFileItem
	if err := json.Unmarshal(body, &files); err != nil {
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
	return dccBaseURL + "/liveviewwebcam.jpg?width=480"
}

// StartLiveView tells digiCamControl to open its Live View window,
// then waits briefly for the JPEG stream to become available.
func (w *WinCamera) StartLiveView() error {
	resp, err := w.client.Get(dccBaseURL + "/?CMD=LiveViewWnd_Show")
	if err != nil {
		return fmt.Errorf("digiCamControl StartLiveView request failed: %w", err)
	}
	io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("digiCamControl StartLiveView failed (HTTP %d)", resp.StatusCode)
	}
	// Give DCC ~1.2 s to spin up the live view feed before the frontend starts polling
	time.Sleep(1200 * time.Millisecond)
	return nil
}

// StopLiveView tells digiCamControl to close its Live View window,
// returning the camera to normal standby (prevents sensor overheating).
func (w *WinCamera) StopLiveView() error {
	resp, err := w.client.Get(dccBaseURL + "/?CMD=LiveViewWnd_Hide")
	if err != nil {
		return fmt.Errorf("digiCamControl StopLiveView request failed: %w", err)
	}
	io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("digiCamControl StopLiveView failed (HTTP %d)", resp.StatusCode)
	}
	return nil
}
