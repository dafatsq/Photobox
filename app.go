package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"photobox/hardware"
)

// App struct holds the application state and hardware drivers.
type App struct {
	ctx     context.Context
	camera  hardware.CameraDriver
	printer hardware.PrinterDriver
	dataDir string // directory to store session outputs
}

// NewApp creates a new App application struct.
func NewApp(camera hardware.CameraDriver, printer hardware.PrinterDriver) *App {
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, "PhotoboxData")
	os.MkdirAll(dataDir, 0755)

	return &App{
		camera:  camera,
		printer: printer,
		dataDir: dataDir,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// CheckPaymentStatus checks the QRIS webhook for a given transaction ID.
// In production, this would call an external payment gateway API.
// For now, it simulates a successful payment after a short delay.
func (a *App) CheckPaymentStatus(trxID string) (bool, error) {
	if trxID == "" {
		return false, fmt.Errorf("transaction ID cannot be empty")
	}
	// TODO: Replace with actual QRIS webhook integration
	// Simulate payment check — in dev mode, always return true
	return true, nil
}

// TriggerCapture commands the camera to take a photo and returns
// the path to the captured image file.
func (a *App) TriggerCapture(sessionID string, sequence int) (string, error) {
	if sessionID == "" {
		return "", fmt.Errorf("session ID cannot be empty")
	}

	// Create session directory
	sessionDir := filepath.Join(a.dataDir, "sessions", sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create session directory: %w", err)
	}

	filename := filepath.Join(sessionDir, fmt.Sprintf("capture_%d_%d.jpg", sequence, time.Now().Unix()))

	if err := a.camera.Capture(filename); err != nil {
		return "", fmt.Errorf("camera capture failed: %w", err)
	}

	return filename, nil
}

// SaveWebRTCImage saves a base64 encoded image from the frontend's WebRTC capture.
func (a *App) SaveWebRTCImage(sessionID string, sequence int, base64Data string) (string, error) {
	if sessionID == "" {
		return "", fmt.Errorf("session ID cannot be empty")
	}

	// Remove data URI prefix if present
	if idx := strings.Index(base64Data, ","); idx != -1 {
		base64Data = base64Data[idx+1:]
	}

	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 image: %w", err)
	}

	// Create session directory
	sessionDir := filepath.Join(a.dataDir, "sessions", sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create session directory: %w", err)
	}

	filename := filepath.Join(sessionDir, fmt.Sprintf("capture_%d_%d.jpg", sequence, time.Now().Unix()))

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return "", fmt.Errorf("failed to save WebRTC capture: %w", err)
	}

	return filename, nil
}

// ProcessComposite combines captured images into a composite frame based on templateID.
// Supported templates: "strip_2x6" (4 photos in a vertical strip), "postcard_4x6" (4 photos in a grid).
func (a *App) ProcessComposite(images []string, templateID string) (string, error) {
	if len(images) == 0 {
		return "", fmt.Errorf("no images provided")
	}

	sessionDir := filepath.Dir(images[0])
	outputPath := filepath.Join(sessionDir, fmt.Sprintf("composite_%s_%d.png", templateID, time.Now().Unix()))

	// Load all source images
	srcImages := make([]image.Image, 0, len(images))
	for _, imgPath := range images {
		f, err := os.Open(imgPath)
		if err != nil {
			return "", fmt.Errorf("failed to open image %s: %w", imgPath, err)
		}

		var img image.Image
		if strings.HasSuffix(strings.ToLower(imgPath), ".png") {
			img, err = png.Decode(f)
		} else {
			img, err = jpeg.Decode(f)
		}
		f.Close()
		if err != nil {
			return "", fmt.Errorf("failed to decode image %s: %w", imgPath, err)
		}
		srcImages = append(srcImages, img)
	}

	var composite *image.RGBA

	switch templateID {
	case "strip_2x6":
		// Vertical strip: 2 inches wide, 6 inches tall at 300DPI = 600x1800px
		// Stack photos vertically
		photoW := 600
		photoH := 1800 / len(srcImages)
		composite = image.NewRGBA(image.Rect(0, 0, photoW, 1800))

		for i, src := range srcImages {
			destRect := image.Rect(0, i*photoH, photoW, (i+1)*photoH)
			draw.Draw(composite, destRect, src, src.Bounds().Min, draw.Src)
		}

	case "postcard_4x6":
		// Postcard grid: 4x6 inches at 300DPI = 1200x1800px
		// 2x2 grid of photos
		photoW := 600
		photoH := 900
		composite = image.NewRGBA(image.Rect(0, 0, 1200, 1800))

		for i, src := range srcImages {
			col := i % 2
			row := i / 2
			destRect := image.Rect(col*photoW, row*photoH, (col+1)*photoW, (row+1)*photoH)
			draw.Draw(composite, destRect, src, src.Bounds().Min, draw.Src)
		}

	default:
		return "", fmt.Errorf("unknown template: %s", templateID)
	}

	// Save composite
	outFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, composite); err != nil {
		return "", fmt.Errorf("failed to encode composite: %w", err)
	}

	return outputPath, nil
}

// PrintPhoto triggers a silent print of the final composite image.
func (a *App) PrintPhoto(finalImagePath string) error {
	if finalImagePath == "" {
		return fmt.Errorf("image path cannot be empty")
	}

	if _, err := os.Stat(finalImagePath); os.IsNotExist(err) {
		return fmt.Errorf("image file not found: %s", finalImagePath)
	}

	if err := a.printer.Print(finalImagePath); err != nil {
		return fmt.Errorf("print failed: %w", err)
	}

	return nil
}

// GetDataDir returns the data directory path (for frontend to know where files are).
func (a *App) GetDataDir() string {
	return a.dataDir
}

// GetLiveViewURL returns the live view URL from the camera driver.
// If empty, the frontend should fall back to WebRTC for preview.
func (a *App) GetLiveViewURL() string {
	return a.camera.LiveViewURL()
}
