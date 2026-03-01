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

	"photobox/admin"
	"photobox/hardware"

	xdraw "golang.org/x/image/draw"
)

// App struct holds the application state and hardware drivers.
type App struct {
	ctx      context.Context
	camera   hardware.CameraDriver
	printer  hardware.PrinterDriver
	dataDir  string // directory to store session outputs
	adminCfg *admin.AdminConfig
}

// NewApp creates a new App application struct.
func NewApp(camera hardware.CameraDriver, printer hardware.PrinterDriver, adminCfg *admin.AdminConfig) *App {
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, "PhotoboxData")
	os.MkdirAll(dataDir, 0755)

	return &App{
		camera:   camera,
		printer:  printer,
		dataDir:  dataDir,
		adminCfg: adminCfg,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// IsPaymentBypassed returns true if the admin has enabled payment bypass.
func (a *App) IsPaymentBypassed() bool {
	return a.adminCfg.GetBypassPayment()
}

// GetFrames returns the current list of available frames from admin config.
func (a *App) GetFrames() []admin.Frame {
	return a.adminCfg.GetFrames()
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
	fmt.Println("[APP] TriggerCapture called. Session:", sessionID, "Seq:", sequence)
	if sessionID == "" {
		return "", fmt.Errorf("session ID cannot be empty")
	}

	// Create session directory
	sessionDir := filepath.Join(a.dataDir, "sessions", sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		fmt.Println("[APP] TriggerCapture err create dir:", err)
		return "", fmt.Errorf("failed to create session directory: %w", err)
	}

	filename := filepath.Join(sessionDir, fmt.Sprintf("capture_%d_%d.jpg", sequence, time.Now().Unix()))
	fmt.Println("[APP] Target filename:", filename)

	if err := a.camera.Capture(filename); err != nil {
		fmt.Println("[APP] TriggerCapture camera.Capture err:", err)
		return "", fmt.Errorf("camera capture failed: %w", err)
	}

	fmt.Println("[APP] TriggerCapture success. Returning:", filename)
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

// ProcessComposite combines captured images into a composite frame based on templateID and frameID.
// Supported templates: "strip_2x6" (4 photos in a vertical strip), "postcard_4x6" (4 photos in a grid).
func (a *App) ProcessComposite(images []string, templateID string, frameID string) (string, error) {
	if len(images) == 0 {
		return "", fmt.Errorf("no images provided")
	}

	sessionDir := filepath.Dir(images[0])
	outputPath := filepath.Join(sessionDir, fmt.Sprintf("composite_%s_%s_%d.png", templateID, frameID, time.Now().Unix()))

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

	// Helper for center cropping an image to match a target aspect ratio
	centerCrop := func(bounds image.Rectangle, targetW, targetH int) image.Rectangle {
		srcW := bounds.Dx()
		srcH := bounds.Dy()

		srcRatio := float64(srcW) / float64(srcH)
		targetRatio := float64(targetW) / float64(targetH)

		var cropW, cropH int
		if srcRatio > targetRatio {
			// Source is relatively wider. Crop left/right.
			cropH = srcH
			cropW = int(float64(srcH) * targetRatio)
		} else {
			// Source is relatively taller. Crop top/bottom.
			cropW = srcW
			cropH = int(float64(srcW) / targetRatio)
		}

		xOffset := (srcW - cropW) / 2
		yOffset := (srcH - cropH) / 2

		return image.Rect(
			bounds.Min.X+xOffset,
			bounds.Min.Y+yOffset,
			bounds.Min.X+xOffset+cropW,
			bounds.Min.Y+yOffset+cropH,
		)
	}

	// Find selected frame to get layouts and overlay
	var selectedFrame *admin.Frame
	if frameID != "none" {
		for _, f := range a.adminCfg.GetFrames() {
			if f.ID == frameID {
				selectedFrame = &f
				break
			}
		}
	}

	// Determine composite resolution based on template
	var compW, compH int
	switch templateID {
	case "strip_2x6":
		compW, compH = 600, 1800
	case "postcard_4x6":
		compW, compH = 1200, 1800
	default:
		return "", fmt.Errorf("unknown template: %s", templateID)
	}

	composite = image.NewRGBA(image.Rect(0, 0, compW, compH))

	// Check if we have custom layouts for these photos
	hasValidLayouts := selectedFrame != nil && len(selectedFrame.Layouts) == len(srcImages)

	if hasValidLayouts {
		// Use custom coordinates from Admin Layout Editor
		for i, src := range srcImages {
			lo := selectedFrame.Layouts[i]
			destRect := image.Rect(lo.X, lo.Y, lo.X+lo.Width, lo.Y+lo.Height)
			cropRect := centerCrop(src.Bounds(), lo.Width, lo.Height)
			xdraw.CatmullRom.Scale(composite, destRect, src, cropRect, draw.Src, nil)
		}
	} else {
		// Fallback to rigid mathematical grid if no custom layouts exist
		switch templateID {
		case "strip_2x6":
			photoW := 600
			photoH := 1800 / len(srcImages)
			for i, src := range srcImages {
				destRect := image.Rect(0, i*photoH, photoW, (i+1)*photoH)
				cropRect := centerCrop(src.Bounds(), photoW, photoH)
				xdraw.CatmullRom.Scale(composite, destRect, src, cropRect, draw.Src, nil)
			}

		case "postcard_4x6":
			photoW := 600
			photoH := 900
			for i, src := range srcImages {
				col := i % 2
				row := i / 2
				destRect := image.Rect(col*photoW, row*photoH, (col+1)*photoW, (row+1)*photoH)
				cropRect := centerCrop(src.Bounds(), photoW, photoH)
				xdraw.CatmullRom.Scale(composite, destRect, src, cropRect, draw.Src, nil)
			}
		}
	}

	// Apply Frame PNG Overlay (Alpha Blending)
	if selectedFrame != nil && selectedFrame.FilePath != "" {
		f, err := os.Open(selectedFrame.FilePath)
		if err == nil {
			defer f.Close()
			overlayImg, err := png.Decode(f)
			if err == nil {
				// Scale overlay to match composite exactly just in case, though it should be exact
				scaledOverlay := image.NewRGBA(image.Rect(0, 0, compW, compH))
				xdraw.CatmullRom.Scale(scaledOverlay, scaledOverlay.Bounds(), overlayImg, overlayImg.Bounds(), draw.Over, nil)

				// Alpha blend over the photo grid
				draw.Draw(composite, composite.Bounds(), scaledOverlay, image.Point{}, draw.Over)
			}
		}
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

// SendEmail mocks sending the final image via email.
func (a *App) SendEmail(imagePath string, emailAddress string) error {
	fmt.Printf("[APP] Mock Send Email -> To: %s | Attachment: %s\n", emailAddress, imagePath)
	// In production, instantiate an SMTP client or AWS SES call here.
	time.Sleep(1500 * time.Millisecond) // simulate network delay
	return nil
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

// StartLiveView activates the camera's live view mode via digiCamControl.
// Call this when the CaptureScreen mounts so the camera is only in live view when needed.
func (a *App) StartLiveView() error {
	fmt.Println("[APP] StartLiveView called")
	return a.camera.StartLiveView()
}

// StopLiveView deactivates the camera's live view mode via digiCamControl.
// Call this when leaving the CaptureScreen to let the camera cool down.
func (a *App) StopLiveView() error {
	fmt.Println("[APP] StopLiveView called")
	return a.camera.StopLiveView()
}

// GetLiveViewURL returns the live view URL from the camera driver.
// If empty, the frontend should fall back to WebRTC for preview.
func (a *App) GetLiveViewURL() string {
	return a.camera.LiveViewURL()
}
