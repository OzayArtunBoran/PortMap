package detector

import (
	"fmt"
	"strings"

	"github.com/ozayartunboran/portmap/internal/allocator"
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
	result := &CheckResult{
		Conflicts:    []Conflict{},
		OKServices:   []string{},
		TotalChecked: len(d.config.Services),
	}

	// 1. Config-internal check: duplicate port assignments
	portOwners := make(map[int][]string)
	for name, svc := range d.config.Services {
		if svc.Port > 0 {
			portOwners[svc.Port] = append(portOwners[svc.Port], name)
		}
	}
	for port, owners := range portOwners {
		if len(owners) > 1 {
			for _, name := range owners {
				result.Conflicts = append(result.Conflicts, Conflict{
					Port:            port,
					Type:            ConflictDuplicate,
					ServiceName:     name,
					ExpectedService: name,
					Message:         fmt.Sprintf("port %d assigned to multiple services: %s", port, strings.Join(owners, ", ")),
				})
			}
		}
	}

	// 2. Scan active ports
	scanResult, err := d.scanner.Scan(scanner.ScanOptions{
		ListenOnly: true,
	})
	if err != nil {
		return result, fmt.Errorf("scan ports: %w", err)
	}

	// Build port → active process map
	activeMap := make(map[int]scanner.PortInfo)
	for _, p := range scanResult.Ports {
		activeMap[p.Port] = p
	}

	// Parse default range for allocator suggestions
	defaultRange, _ := config.ParseRange(d.config.Defaults.Range)
	if defaultRange.Start == 0 {
		defaultRange.Start = 3000
	}
	if defaultRange.End == 0 {
		defaultRange.End = 9999
	}
	alloc := allocator.New()

	// 3. Check each configured service
	for name, svc := range d.config.Services {
		if svc.Port <= 0 {
			continue
		}

		active, isActive := activeMap[svc.Port]
		if !isActive {
			// Port is free
			if svc.IsManaged() {
				result.OKServices = append(result.OKServices, name)
			} else {
				// Unmanaged service not running — just note it as OK
				result.OKServices = append(result.OKServices, name)
			}
			continue
		}

		// Port is active — check if it's the expected service
		if isExpectedProcess(name, svc, active) {
			result.OKServices = append(result.OKServices, name)
			continue
		}

		// Conflict: port occupied by different process
		suggestion := 0
		// Collect ports to exclude
		var exclude []int
		for _, s := range d.config.Services {
			if s.Port > 0 {
				exclude = append(exclude, s.Port)
			}
		}
		suggested, err := alloc.FindFree(allocator.AllocRequest{
			PreferredPort: svc.Port,
			Range:         defaultRange,
			Strategy:      allocator.StrategyNearest,
			Count:         1,
			Exclude:       exclude,
		})
		if err == nil && len(suggested) > 0 {
			suggestion = suggested[0]
		}

		result.Conflicts = append(result.Conflicts, Conflict{
			Port:            svc.Port,
			Type:            ConflictOccupied,
			ServiceName:     name,
			ExpectedService: name,
			ActualProcess:   active.ProcessName,
			ActualPID:       active.PID,
			Suggestion:      suggestion,
			Message: fmt.Sprintf("port %d expected by %q but occupied by %s (PID %d)",
				svc.Port, name, active.ProcessName, active.PID),
		})
	}

	return result, nil
}

// isExpectedProcess checks if the running process matches the expected service
func isExpectedProcess(name string, svc config.ServiceConfig, active scanner.PortInfo) bool {
	processLower := strings.ToLower(active.ProcessName)
	nameLower := strings.ToLower(name)

	// Direct name match
	if strings.Contains(processLower, nameLower) {
		return true
	}

	// Check if the command matches
	if svc.Command != "" {
		cmdParts := strings.Fields(svc.Command)
		if len(cmdParts) > 0 {
			cmdBase := strings.ToLower(cmdParts[0])
			if strings.Contains(processLower, cmdBase) {
				return true
			}
		}
	}

	return false
}

// HasConflicts returns true if any conflicts were found
func (r *CheckResult) HasConflicts() bool {
	return len(r.Conflicts) > 0
}
