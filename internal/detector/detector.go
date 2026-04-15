package detector

import (
	"github.com/ozayartunboran/portmap/internal/config"
	"github.com/ozayartunboran/portmap/internal/scanner"
)

// ConflictType categorizes the kind of port conflict
type ConflictType string

const (
	ConflictOccupied     ConflictType = "OCCUPIED"      // Port is used by another process
	ConflictDuplicate    ConflictType = "DUPLICATE"      // Same port assigned to multiple services in config
	ConflictRangeOverlap ConflictType = "RANGE_OVERLAP"  // Port ranges overlap
)

// Conflict describes a single port conflict
type Conflict struct {
	Port            int          `json:"port"`
	Type            ConflictType `json:"type"`
	ServiceName     string       `json:"service_name"`      // Config service name
	ExpectedService string       `json:"expected_service"`   // Who should own this port
	ActualProcess   string       `json:"actual_process"`     // Who currently owns this port
	ActualPID       int          `json:"actual_pid"`
	Suggestion      int          `json:"suggestion"`         // Suggested alternative port
	Message         string       `json:"message"`            // Human-readable message
}

// CheckResult holds the full result of a conflict check
type CheckResult struct {
	Conflicts    []Conflict `json:"conflicts"`
	OKServices   []string   `json:"ok_services"`
	TotalChecked int        `json:"total_checked"`
}

// Detector checks for port conflicts between config and running ports
type Detector struct {
	config  *config.PortmapConfig
	scanner scanner.Scanner
}

// New creates a new Detector
func New(cfg *config.PortmapConfig, scn scanner.Scanner) *Detector {
	return &Detector{
		config:  cfg,
		scanner: scn,
	}
}

// Check performs the conflict detection
func (d *Detector) Check() (*CheckResult, error) {
	// TODO: Phase 2 — implement:
	// 1. Check config for duplicate ports (config-internal)
	// 2. Scan active ports
	// 3. For each configured service:
	//    - Is its port occupied by someone else? → OCCUPIED
	//    - Is its port assigned to multiple services? → DUPLICATE
	// 4. For conflicts, suggest alternatives using allocator
	// 5. Return CheckResult

	result := &CheckResult{
		Conflicts:    []Conflict{},
		OKServices:   []string{},
		TotalChecked: len(d.config.Services),
	}

	return result, nil
}

// HasConflicts returns true if any conflicts were found
func (r *CheckResult) HasConflicts() bool {
	return len(r.Conflicts) > 0
}
