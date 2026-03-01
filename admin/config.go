package admin

import (
	"sync"
)

// Frame represents a selectable frame/border option.
type Frame struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Color string `json:"color"`
}

// AdminConfig holds all admin-configurable settings, safe for concurrent access.
type AdminConfig struct {
	mu            sync.RWMutex
	bypassPayment bool
	frames        []Frame
}

// NewAdminConfig creates a config with sensible defaults.
func NewAdminConfig() *AdminConfig {
	return &AdminConfig{
		bypassPayment: false,
		frames: []Frame{
			{ID: "none", Label: "No Frame", Color: "transparent"},
			{ID: "classic_black", Label: "Classic Black", Color: "#1a1a1a"},
			{ID: "classic_white", Label: "Classic White", Color: "#ffffff"},
			{ID: "neon_pink", Label: "Neon Pink", Color: "#ff2a6d"},
			{ID: "neon_blue", Label: "Neon Blue", Color: "#05d9e8"},
			{ID: "vintage_gold", Label: "Vintage Gold", Color: "#d4af37"},
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
