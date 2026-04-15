package allocator

import (
	"fmt"
	"net"
	"testing"

	"github.com/ozayartunboran/portmap/internal/config"
)

// occupyPort starts a TCP listener and returns it (caller must Close)
func occupyPort(t *testing.T, port int) net.Listener {
	t.Helper()
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		t.Fatalf("failed to occupy port %d: %v", port, err)
	}
	return ln
}

// findFreeTestPort finds a free port for testing purposes
func findFreeTestPort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port
}

func TestFindFree_NearestStrategy_PreferredFree(t *testing.T) {
	// Find a free port to use as preferred
	preferred := findFreeTestPort(t)

	alloc := New()
	ports, err := alloc.FindFree(AllocRequest{
		PreferredPort: preferred,
		Range:         config.PortRange{Start: preferred - 10, End: preferred + 10},
		Strategy:      StrategyNearest,
		Count:         1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 1 {
		t.Fatalf("got %d ports, want 1", len(ports))
	}
	if ports[0] != preferred {
		t.Errorf("port = %d, want preferred %d", ports[0], preferred)
	}
}

func TestFindFree_NearestStrategy_Occupied(t *testing.T) {
	// Find a free port, then occupy it
	preferred := findFreeTestPort(t)
	ln := occupyPort(t, preferred)
	defer ln.Close()

	alloc := New()
	ports, err := alloc.FindFree(AllocRequest{
		PreferredPort: preferred,
		Range:         config.PortRange{Start: preferred - 10, End: preferred + 10},
		Strategy:      StrategyNearest,
		Count:         1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 1 {
		t.Fatalf("got %d ports, want 1", len(ports))
	}
	if ports[0] == preferred {
		t.Error("should not return the occupied port")
	}
	// Should be ±1 from preferred (nearest)
	diff := ports[0] - preferred
	if diff < 0 {
		diff = -diff
	}
	if diff > 10 {
		t.Errorf("returned port %d is too far from preferred %d", ports[0], preferred)
	}
}

func TestFindFree_SequentialStrategy(t *testing.T) {
	// Use a high port range that's likely free
	start := findFreeTestPort(t)

	alloc := New()
	ports, err := alloc.FindFree(AllocRequest{
		Range:    config.PortRange{Start: start, End: start + 100},
		Strategy: StrategySequential,
		Count:    1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 1 {
		t.Fatalf("got %d ports, want 1", len(ports))
	}
	if ports[0] < start || ports[0] > start+100 {
		t.Errorf("port %d out of range [%d, %d]", ports[0], start, start+100)
	}
}

func TestFindFree_RandomStrategy(t *testing.T) {
	start := findFreeTestPort(t)

	alloc := New()
	ports, err := alloc.FindFree(AllocRequest{
		Range:    config.PortRange{Start: start, End: start + 100},
		Strategy: StrategyRandom,
		Count:    1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 1 {
		t.Fatalf("got %d ports, want 1", len(ports))
	}
	if ports[0] < start || ports[0] > start+100 {
		t.Errorf("port %d out of range [%d, %d]", ports[0], start, start+100)
	}
}

func TestFindFree_MultipleCount(t *testing.T) {
	start := findFreeTestPort(t)

	alloc := New()
	ports, err := alloc.FindFree(AllocRequest{
		Range:    config.PortRange{Start: start, End: start + 200},
		Strategy: StrategySequential,
		Count:    3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 3 {
		t.Fatalf("got %d ports, want 3", len(ports))
	}

	// All should be unique
	seen := make(map[int]bool)
	for _, p := range ports {
		if seen[p] {
			t.Errorf("duplicate port %d", p)
		}
		seen[p] = true
	}
}

func TestFindFree_ExcludeList(t *testing.T) {
	start := findFreeTestPort(t)

	alloc := New()
	exclude := []int{start, start + 1, start + 2}
	ports, err := alloc.FindFree(AllocRequest{
		Range:    config.PortRange{Start: start, End: start + 100},
		Strategy: StrategySequential,
		Count:    1,
		Exclude:  exclude,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 1 {
		t.Fatalf("got %d ports, want 1", len(ports))
	}
	for _, ex := range exclude {
		if ports[0] == ex {
			t.Errorf("returned excluded port %d", ports[0])
		}
	}
}

func TestFindFree_NoFreePorts(t *testing.T) {
	// Use a tiny range and occupy all ports
	start := findFreeTestPort(t)
	var listeners []net.Listener
	defer func() {
		for _, ln := range listeners {
			ln.Close()
		}
	}()

	// Occupy a small range
	rangeSize := 3
	for i := 0; i < rangeSize; i++ {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", start+i))
		if err != nil {
			// Port might already be in use, try to find another start
			t.Skipf("could not occupy port %d: %v", start+i, err)
		}
		listeners = append(listeners, ln)
	}

	alloc := New()
	_, err := alloc.FindFree(AllocRequest{
		Range:    config.PortRange{Start: start, End: start + rangeSize - 1},
		Strategy: StrategySequential,
		Count:    1,
	})
	if err == nil {
		t.Error("expected error when no free ports available")
	}
}

func TestIsPortFree(t *testing.T) {
	// Find a free port
	freePort := findFreeTestPort(t)

	// Should be free
	if !IsPortFree(freePort) {
		t.Errorf("port %d should be free", freePort)
	}

	// Occupy it
	ln := occupyPort(t, freePort)
	defer ln.Close()

	// Should not be free
	if IsPortFree(freePort) {
		t.Errorf("port %d should not be free (occupied)", freePort)
	}
}

func TestFindFree_DefaultCount(t *testing.T) {
	start := findFreeTestPort(t)

	alloc := New()
	// Count=0 should default to 1
	ports, err := alloc.FindFree(AllocRequest{
		Range:    config.PortRange{Start: start, End: start + 100},
		Strategy: StrategySequential,
		Count:    0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 1 {
		t.Errorf("got %d ports, want 1 (default count)", len(ports))
	}
}

func TestFindFree_DefaultRange(t *testing.T) {
	alloc := New()
	// Range with zero values should use defaults
	ports, err := alloc.FindFree(AllocRequest{
		Range:    config.PortRange{Start: 0, End: 0},
		Strategy: StrategySequential,
		Count:    1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 1 {
		t.Fatalf("got %d ports, want 1", len(ports))
	}
	if ports[0] < 3000 || ports[0] > 9999 {
		t.Errorf("port %d not in default range 3000-9999", ports[0])
	}
}
