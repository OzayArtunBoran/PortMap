package allocator

import (
	"fmt"
	"math/rand"
	"net"

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
	PreferredPort int              // Desired port (0 = any)
	Range         config.PortRange // Search range
	Strategy      AllocStrategy    // Search strategy
	Count         int              // How many ports to find (default 1)
	Exclude       []int            // Ports to skip
}

// Allocator finds available ports
type Allocator struct{}

// New creates a new Allocator
func New() *Allocator {
	return &Allocator{}
}

// FindFree finds available ports matching the request
func (a *Allocator) FindFree(req AllocRequest) ([]int, error) {
	if req.Count <= 0 {
		req.Count = 1
	}
	if req.Range.Start == 0 {
		req.Range.Start = 3000
	}
	if req.Range.End == 0 {
		req.Range.End = 9999
	}

	excludeSet := make(map[int]bool, len(req.Exclude))
	for _, p := range req.Exclude {
		excludeSet[p] = true
	}

	switch req.Strategy {
	case StrategyNearest:
		return a.findNearest(req, excludeSet)
	case StrategySequential:
		return a.findSequential(req, excludeSet)
	case StrategyRandom:
		return a.findRandom(req, excludeSet)
	default:
		return a.findNearest(req, excludeSet)
	}
}

func (a *Allocator) findNearest(req AllocRequest, excludeSet map[int]bool) ([]int, error) {
	preferred := req.PreferredPort
	if preferred == 0 {
		preferred = req.Range.Start
	}

	var found []int
	foundSet := make(map[int]bool)
	maxDelta := req.Range.End - req.Range.Start

	for delta := 0; delta <= maxDelta; delta++ {
		candidates := []int{preferred + delta}
		if delta > 0 {
			candidates = append(candidates, preferred-delta)
		}
		for _, port := range candidates {
			if port < req.Range.Start || port > req.Range.End {
				continue
			}
			if excludeSet[port] || foundSet[port] {
				continue
			}
			if IsPortFree(port) {
				found = append(found, port)
				foundSet[port] = true
				if len(found) >= req.Count {
					return found, nil
				}
			}
		}
	}

	if len(found) == 0 {
		return nil, fmt.Errorf("no free ports in range %d-%d", req.Range.Start, req.Range.End)
	}
	return found, nil
}

func (a *Allocator) findSequential(req AllocRequest, excludeSet map[int]bool) ([]int, error) {
	var found []int
	for port := req.Range.Start; port <= req.Range.End; port++ {
		if excludeSet[port] {
			continue
		}
		if IsPortFree(port) {
			found = append(found, port)
			if len(found) >= req.Count {
				return found, nil
			}
		}
	}
	if len(found) == 0 {
		return nil, fmt.Errorf("no free ports in range %d-%d", req.Range.Start, req.Range.End)
	}
	return found, nil
}

func (a *Allocator) findRandom(req AllocRequest, excludeSet map[int]bool) ([]int, error) {
	rangeSize := req.Range.End - req.Range.Start + 1
	if rangeSize <= 0 {
		return nil, fmt.Errorf("invalid port range %d-%d", req.Range.Start, req.Range.End)
	}

	var found []int
	foundSet := make(map[int]bool)
	maxAttempts := rangeSize * 2
	if maxAttempts > 10000 {
		maxAttempts = 10000
	}

	for attempt := 0; attempt < maxAttempts && len(found) < req.Count; attempt++ {
		port := req.Range.Start + rand.Intn(rangeSize)
		if excludeSet[port] || foundSet[port] {
			continue
		}
		if IsPortFree(port) {
			found = append(found, port)
			foundSet[port] = true
		}
	}

	if len(found) == 0 {
		return nil, fmt.Errorf("no free ports in range %d-%d", req.Range.Start, req.Range.End)
	}
	return found, nil
}

// IsPortFree checks if a port is available by attempting to bind
func IsPortFree(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}
