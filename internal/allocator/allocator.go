package allocator

import (
	"github.com/ozayartunboran/portmap/internal/config"
)

// AllocStrategy defines how to search for free ports
type AllocStrategy string

const (
	StrategyNearest    AllocStrategy = "nearest"    // Find port closest to preferred
	StrategySequential AllocStrategy = "sequential" // Find first available port
	StrategyRandom     AllocStrategy = "random"     // Find random available port
)

// AllocRequest specifies parameters for finding free ports
type AllocRequest struct {
	PreferredPort int           // Desired port (0 = any)
	Range         config.PortRange // Search range
	Strategy      AllocStrategy // Search strategy
	Count         int           // How many ports to find (default 1)
	Exclude       []int         // Ports to skip
}

// Allocator finds available ports
type Allocator struct{}

// New creates a new Allocator
func New() *Allocator {
	return &Allocator{}
}

// FindFree finds available ports matching the request
func (a *Allocator) FindFree(req AllocRequest) ([]int, error) {
	// TODO: Phase 2 — implement:
	// 1. Based on strategy:
	//    - nearest: start from preferred, expand ±1, ±2, ...
	//    - sequential: iterate from range.Start
	//    - random: pick random port in range, verify free
	// 2. For each candidate:
	//    - Skip if in exclude list
	//    - Try net.Listen(":{port}") to verify it's free
	//    - If free, add to results
	// 3. Return when Count ports found or range exhausted

	return []int{}, nil
}

// IsPortFree checks if a port is available by attempting to bind
func IsPortFree(port int) bool {
	// TODO: Phase 2 — implement:
	// net.Listen("tcp", fmt.Sprintf(":%d", port))
	// If successful, close and return true
	// If error, return false
	return false
}
