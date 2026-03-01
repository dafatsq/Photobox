package admin

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// persistedState is used strictly for serializing to config.json
type persistedState struct {
	BypassPayment bool    `json:"bypassPayment"`
	Frames        []Frame `json:"frames"`
}

// PhotoLayout defines the exact coordinates and size for a single photo in the composite.
type PhotoLayout struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Frame represents a selectable frame/border option.
type Frame struct {
	ID       string        `json:"id"`
	Label    string        `json:"label"`
	FilePath string        `json:"filePath"` // absolute path on disk
	Template string        `json:"template"` // "strip_2x6" | "postcard_4x6" | "" (all)
	Layouts  []PhotoLayout `json:"layouts"`  // Custom coordinates for each photo
}

// AdminConfig holds all admin-configurable settings, safe for concurrent access.
type AdminConfig struct {
	mu            sync.RWMutex
	bypassPayment bool
	frames        []Frame
	framesDir     string // directory where PNG frame files are stored
}

// NewAdminConfig creates a config with sensible defaults and attempts to load existing data.
func NewAdminConfig(framesDir string) *AdminConfig {
	c := &AdminConfig{
		bypassPayment: false,
		framesDir:     framesDir,
		frames: []Frame{
			{ID: "none", Label: "No Frame", FilePath: "", Template: ""},
		},
	}

	// Attempt to load existing config on boot
	if err := c.Load(); err != nil {
		log.Printf("[Admin Config] No existing config found or failed to load, starting fresh: %v", err)
	}

	return c
}

// ConfigFilePath returns the absolute path to the config.json file.
func (c *AdminConfig) ConfigFilePath() string {
	return filepath.Join(c.framesDir, "config.json")
}

// Save writes the current config state todisk.
// This is called internally, assumes calling function already holds the c.mu.Lock()
func (c *AdminConfig) Save() error {
	state := persistedState{
		BypassPayment: c.bypassPayment,
		Frames:        make([]Frame, len(c.frames)),
	}
	copy(state.Frames, c.frames)

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(c.framesDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(c.ConfigFilePath(), data, 0644)
}

// Load reads the config from disk, replacing current state.
func (c *AdminConfig) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := os.ReadFile(c.ConfigFilePath())
	if err != nil {
		return err
	}

	var state persistedState
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	c.bypassPayment = state.BypassPayment

	// Ensure "none" frame always exists, otherwise load everything
	hasNone := false
	for _, f := range state.Frames {
		if f.ID == "none" {
			hasNone = true
			break
		}
	}

	if !hasNone {
		// prepended
		state.Frames = append([]Frame{{ID: "none", Label: "No Frame", FilePath: "", Template: ""}}, state.Frames...)
	}

	c.frames = state.Frames

	log.Printf("[Admin Config] Loaded %d frames from %s", len(c.frames), c.ConfigFilePath())
	return nil
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
	if err := c.Save(); err != nil {
		log.Printf("[Admin Config] Failed to save after setting bypass: %v", err)
	}
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
	if err := c.Save(); err != nil {
		log.Printf("[Admin Config] Failed to save after adding frame: %v", err)
	}
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
	if err := c.Save(); err != nil {
		log.Printf("[Admin Config] Failed to save after removing frame: %v", err)
	}
}

func (c *AdminConfig) UpdateFrameLayouts(id string, layouts []PhotoLayout) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.frames {
		if c.frames[i].ID == id {
			// Make a copy of layouts to prevent external modification
			newLayouts := make([]PhotoLayout, len(layouts))
			copy(newLayouts, layouts)
			c.frames[i].Layouts = newLayouts
			if err := c.Save(); err != nil {
				log.Printf("[Admin Config] Failed to save after updating frame layouts: %v", err)
			}
			break
		}
	}
}

func (c *AdminConfig) FramesDir() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.framesDir
}
