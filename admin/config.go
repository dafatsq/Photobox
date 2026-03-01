package admin

import (
	"sync"
)

// Frame represents a selectable frame/border option.
type Frame struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	FilePath string `json:"filePath"` // absolute path on disk
	Template string `json:"template"` // "strip_2x6" | "postcard_4x6" | "" (all)
}

// AdminConfig holds all admin-configurable settings, safe for concurrent access.
type AdminConfig struct {
	mu            sync.RWMutex
	bypassPayment bool
	frames        []Frame
	framesDir     string // directory where PNG frame files are stored
}

// NewAdminConfig creates a config with sensible defaults.
func NewAdminConfig(framesDir string) *AdminConfig {
	return &AdminConfig{
		bypassPayment: false,
		framesDir:     framesDir,
		frames: []Frame{
			{ID: "none", Label: "No Frame", FilePath: "", Template: ""},
		},
	}
}

func (c *AdminConfig) GetBypassPayment() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.bypassPayment
}

func (c *AdminConfig) SetBypassPayment(val bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bypassPayment = val
}

func (c *AdminConfig) GetFrames() []Frame {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]Frame, len(c.frames))
	copy(result, c.frames)
	return result
}

func (c *AdminConfig) AddFrame(f Frame) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Prevent duplicates
	for _, existing := range c.frames {
		if existing.ID == f.ID {
			return
		}
	}
	c.frames = append(c.frames, f)
}

func (c *AdminConfig) RemoveFrame(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	filtered := make([]Frame, 0, len(c.frames))
	for _, f := range c.frames {
		if f.ID != id {
			filtered = append(filtered, f)
		}
	}
	c.frames = filtered
}

func (c *AdminConfig) FramesDir() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.framesDir
}
