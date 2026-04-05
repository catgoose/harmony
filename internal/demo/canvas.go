// setup:feature:demo

package demo

import (
	"fmt"
	"sync"
	"time"
)

// CanvasSize is the grid dimension (CanvasSize x CanvasSize).
const CanvasSize = 64

// CanvasPalette is the set of colors players can choose from.
var CanvasPalette = []string{
	"#ef4444", // red
	"#f97316", // orange
	"#eab308", // yellow
	"#22c55e", // green
	"#06b6d4", // cyan
	"#3b82f6", // blue
	"#8b5cf6", // violet
	"#ec4899", // pink
	"#000000", // black
	"#ffffff", // white
}

// CanvasClient tracks a connected client.
type CanvasClient struct {
	LastSeen time.Time `json:"-"`
	Color    string    `json:"color"`
}

// PixelCanvas is a thread-safe in-memory pixel grid with client tracking.
type PixelCanvas struct {
	clients map[string]*CanvasClient
	Cells   [CanvasSize * CanvasSize]string
	mu      sync.RWMutex
}

// NewPixelCanvas creates an empty canvas.
func NewPixelCanvas() *PixelCanvas {
	return &PixelCanvas{
		clients: make(map[string]*CanvasClient),
	}
}

// PlaceColor sets a cell's color.
func (pc *PixelCanvas) PlaceColor(x, y int, color string) error {
	if x < 0 || x >= CanvasSize || y < 0 || y >= CanvasSize {
		return fmt.Errorf("coordinates out of range")
	}
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.Cells[y*CanvasSize+x] = color
	return nil
}

// Snapshot returns a copy of the grid.
func (pc *PixelCanvas) Snapshot() [CanvasSize * CanvasSize]string {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.Cells
}

// Reset clears the entire canvas.
func (pc *PixelCanvas) Reset() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.Cells = [CanvasSize * CanvasSize]string{}
}

// TouchClient registers or refreshes a client with their current color.
func (pc *PixelCanvas) TouchClient(clientID, color string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	cl, ok := pc.clients[clientID]
	if !ok {
		cl = &CanvasClient{}
		pc.clients[clientID] = cl
	}
	cl.Color = color
	cl.LastSeen = time.Now()
}

// ActiveClients returns the count and colors of clients seen in the last 30 seconds.
func (pc *PixelCanvas) ActiveClients() []CanvasClient {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	cutoff := time.Now().Add(-30 * time.Second)
	out := make([]CanvasClient, 0)
	for _, c := range pc.clients {
		if c.LastSeen.After(cutoff) {
			out = append(out, *c)
		}
	}
	return out
}

// PruneStale removes clients not seen in the last 60 seconds.
func (pc *PixelCanvas) PruneStale() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	cutoff := time.Now().Add(-60 * time.Second)
	for id, c := range pc.clients {
		if c.LastSeen.Before(cutoff) {
			delete(pc.clients, id)
		}
	}
}
