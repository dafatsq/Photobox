//go:build windows

package hardware

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	bridgePort    = 5513
	bridgeBaseURL = "http://localhost:5513"
	bridgeTimeout = 30 * time.Second
)

// WinCamera implements CameraDriver for Windows using the DSLRBridge process.
// DSLRBridge is a lightweight C# process that wraps digiCamControl's camera
// device libraries, providing direct DSLR control without needing the full
// digiCamControl application to be running.
type WinCamera struct {
	client    *http.Client
	bridgeCmd *exec.Cmd
	bridgeMu  sync.Mutex
	outputDir string
}

func NewWinCamera() *WinCamera {
	cam := &WinCamera{
		client: &http.Client{
			Timeout:   bridgeTimeout,
			Transport: &http.Transport{DisableKeepAlives: true},
		},
	}
	return cam
}

// EnsureBridge starts the DSLRBridge process if it's not already running.
// It looks for DSLRBridge.exe relative to the current executable.
func (w *WinCamera) EnsureBridge() error {
	w.bridgeMu.Lock()
	defer w.bridgeMu.Unlock()

	// Check if bridge is already running and responsive
	if w.bridgeCmd != nil && w.bridgeCmd.Process != nil {
		if w.pingBridge() {
			return nil
		}
	}

	// Find DSLRBridge.exe — look in several places
	bridgePath := w.findBridgeExe()
	if bridgePath == "" {
		return fmt.Errorf("DSLRBridge.exe not found — please build the DSLRBridge project")
	}
	println("[DSLRBridge] Found at:", bridgePath)

	// Set up output directory
	w.outputDir = filepath.Join(os.TempDir(), "DSLRBridge_Captures")
	os.MkdirAll(w.outputDir, 0755)

	// Launch bridge process
	cmd := exec.Command(bridgePath,
		"--port", fmt.Sprintf("%d", bridgePort),
		"--output", w.outputDir,
	)
	cmd.Dir = filepath.Dir(bridgePath) // Run from the bridge's directory (for EDSDK.dll)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start DSLRBridge: %w", err)
	}
	w.bridgeCmd = cmd
	println("[DSLRBridge] Started with PID:", cmd.Process.Pid)

	// Wait for bridge to become responsive (up to 10 seconds)
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if w.pingBridge() {
			println("[DSLRBridge] Bridge is responsive")
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("DSLRBridge started but not responding on port %d", bridgePort)
}

// StopBridge gracefully shuts down the DSLRBridge process.
func (w *WinCamera) StopBridge() {
	w.bridgeMu.Lock()
	defer w.bridgeMu.Unlock()

	if w.bridgeCmd == nil || w.bridgeCmd.Process == nil {
		return
	}

	println("[DSLRBridge] Shutting down bridge...")
	// Send shutdown command via HTTP
	client := &http.Client{
		Timeout:   2 * time.Second,
		Transport: &http.Transport{DisableKeepAlives: true},
	}
	resp, err := client.Get(bridgeBaseURL + "/shutdown")
	if err == nil {
		io.ReadAll(resp.Body)
		resp.Body.Close()
	}

	// Wait for process to exit (up to 5 seconds)
	done := make(chan error, 1)
	go func() { done <- w.bridgeCmd.Wait() }()
	select {
	case <-done:
		println("[DSLRBridge] Bridge process exited cleanly")
	case <-time.After(5 * time.Second):
		println("[DSLRBridge] Bridge process did not exit, killing...")
		w.bridgeCmd.Process.Kill()
	}
	w.bridgeCmd = nil
}

// findBridgeExe searches for DSLRBridge.exe in common locations.
func (w *WinCamera) findBridgeExe() string {
	candidates := []string{}

	// 1. Next to the current executable
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates, filepath.Join(exeDir, "DSLRBridge.exe"))
		candidates = append(candidates, filepath.Join(exeDir, "DSLRBridge", "DSLRBridge.exe"))
	}

	// 2. Relative to working directory (for development)
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(cwd, "DSLRBridge", "bin", "x86", "Debug", "DSLRBridge.exe"),
			filepath.Join(cwd, "DSLRBridge", "bin", "x86", "Release", "DSLRBridge.exe"),
			filepath.Join(cwd, "DSLRBridge", "DSLRBridge.exe"),
		)
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

// pingBridge checks if the bridge HTTP server is responding.
func (w *WinCamera) pingBridge() bool {
	client := &http.Client{
		Timeout:   1 * time.Second,
		Transport: &http.Transport{DisableKeepAlives: true},
	}
	resp, err := client.Get(bridgeBaseURL + "/ping")
	if err != nil {
		return false
	}
	io.ReadAll(resp.Body)
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// bridgeResponse is the generic JSON response from the bridge.
type bridgeResponse struct {
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
	File    string `json:"file,omitempty"`
	Camera  string `json:"camera,omitempty"`
	Battery int    `json:"battery,omitempty"`
}

// Capture triggers a capture via the DSLRBridge and copies the result to filename.
func (w *WinCamera) Capture(filename string) error {
	if err := w.EnsureBridge(); err != nil {
		return fmt.Errorf("DSLRBridge not available: %w", err)
	}

	println("[DSLRBridge] Sending capture request...")
	resp, err := w.client.Get(bridgeBaseURL + "/capture")
	if err != nil {
		return fmt.Errorf("capture request failed: %w", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	println("[DSLRBridge] Capture response:", string(body))

	var result bridgeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse capture response: %w", err)
	}

	if result.Status != "ok" {
		return fmt.Errorf("capture failed: %s", result.Error)
	}

	if result.File == "" {
		return fmt.Errorf("capture succeeded but no file path returned")
	}

	// The bridge returns forward slashes; convert back to native path
	srcPath := strings.ReplaceAll(result.File, "/", string(os.PathSeparator))
	println("[DSLRBridge] Copying", srcPath, "to", filename)
	return w.copyFile(srcPath, filename)
}

// LiveViewURL returns the DSLRBridge live view JPEG endpoint.
// Always returns the URL if the bridge is running (the endpoint self-handles errors).
func (w *WinCamera) LiveViewURL() string {
	if err := w.EnsureBridge(); err != nil {
		return ""
	}
	return bridgeBaseURL + "/liveview.jpg"
}

// StartLiveView tells the DSLRBridge to start the camera's live view.
// Errors are logged but not returned — live view is optional for DSLR capture.
func (w *WinCamera) StartLiveView() error {
	if err := w.EnsureBridge(); err != nil {
		println("[DSLRBridge] StartLiveView: bridge not available:", err.Error())
		return nil // Don't fail hard — camera can still capture without live view
	}

	// Use a short timeout — StartLiveView can block on older cameras
	lvClient := &http.Client{
		Timeout:   5 * time.Second,
		Transport: &http.Transport{DisableKeepAlives: true},
	}
	resp, err := lvClient.Get(bridgeBaseURL + "/liveview/start")
	if err != nil {
		println("[DSLRBridge] StartLiveView: request failed:", err.Error())
		return nil // Best-effort
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	var result bridgeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		println("[DSLRBridge] StartLiveView: parse error:", err.Error())
		return nil
	}
	if result.Status == "error" {
		println("[DSLRBridge] StartLiveView: bridge returned error:", result.Error)
		// Don't return error — live view is optional
		return nil
	}

	// Give the camera a moment to initialize the live view stream
	time.Sleep(800 * time.Millisecond)
	return nil
}

// StopLiveView tells the DSLRBridge to stop the camera's live view.
func (w *WinCamera) StopLiveView() error {
	resp, err := w.client.Get(bridgeBaseURL + "/liveview/stop")
	if err != nil {
		return fmt.Errorf("stop live view request failed: %w", err)
	}
	io.ReadAll(resp.Body)
	resp.Body.Close()
	return nil
}

// ConnectCamera tells the DSLRBridge to scan and connect to cameras.
func (w *WinCamera) ConnectCamera() error {
	if err := w.EnsureBridge(); err != nil {
		return err
	}

	resp, err := w.client.Get(bridgeBaseURL + "/connect")
	if err != nil {
		return fmt.Errorf("connect request failed: %w", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	var result bridgeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse connect response: %w", err)
	}
	if result.Status == "error" || result.Status == "no_camera" {
		return fmt.Errorf("connect failed: %s", result.Error)
	}

	println("[DSLRBridge] Connected to camera:", result.Camera)
	return nil
}

// DisconnectCamera tells the DSLRBridge to disconnect from the camera.
func (w *WinCamera) DisconnectCamera() error {
	resp, err := w.client.Get(bridgeBaseURL + "/disconnect")
	if err != nil {
		return fmt.Errorf("disconnect request failed: %w", err)
	}
	io.ReadAll(resp.Body)
	resp.Body.Close()
	return nil
}

// IsCameraConnected ensures the bridge is running, then checks if a camera
// is detected. Retries for up to ~8 seconds to allow the bridge's camera
// scan to complete before reporting failure.
func (w *WinCamera) IsCameraConnected() bool {
	// Start the bridge if not already running — must happen before checking camera status
	if err := w.EnsureBridge(); err != nil {
		println("[DSLRBridge] IsCameraConnected: bridge not available:", err.Error())
		return false
	}

	client := &http.Client{
		Timeout:   6 * time.Second,
		Transport: &http.Transport{DisableKeepAlives: true},
	} // /connect takes at least 1.5s
	maxAttempts := 4 // We wait in chunks, so 4 attempts is enough

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// 1. Initial ping check
		resp, err := client.Get(bridgeBaseURL + "/ping")
		var isConnected bool

		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			var result struct {
				Connected bool `json:"connected"`
			}
			if json.Unmarshal(body, &result) == nil {
				isConnected = result.Connected
			}
		}

		if isConnected {
			println("[DSLRBridge] Camera connected (attempt", attempt, ")")
			return true
		}

		// 2. If we reach here, bridge is running but camera is logically missing
		println("[DSLRBridge] Camera missing, forcing USB rescan (attempt", attempt, "/", maxAttempts, ")...")

		// 3. Trigger /connect rescan in the C# bridge
		cResp, cerr := client.Get(bridgeBaseURL + "/connect")
		if cerr == nil {
			io.ReadAll(cResp.Body)
			cResp.Body.Close()
		}

		if attempt < maxAttempts {
			time.Sleep(1 * time.Second)
		}
	}
	println("[DSLRBridge] Camera not detected after all retries")
	return false
}

// copyFile copies src to dst, retrying if the source is still locked.
func (w *WinCamera) copyFile(src, dst string) error {
	var in *os.File
	var err error
	// Retry up to 10 times (5 seconds) in case the file is still being written
	for i := 0; i < 10; i++ {
		in, err = os.Open(src)
		if err == nil {
			break
		}
		println("[DSLRBridge] File still locked, retrying in 500ms...", err.Error())
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
	println("[DSLRBridge] Bytes copied:", n)
	return err
}
