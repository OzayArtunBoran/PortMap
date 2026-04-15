package detector

import (
	"testing"

	"github.com/ozayartunboran/portmap/internal/config"
	"github.com/ozayartunboran/portmap/internal/scanner"
)

// mockScanner implements scanner.Scanner for testing
type mockScanner struct {
	ports []scanner.PortInfo
	err   error
}

func (m *mockScanner) Scan(opts scanner.ScanOptions) (*scanner.ScanResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &scanner.ScanResult{
		Ports: m.ports,
		Total: len(m.ports),
	}, nil
}

func TestCheck_NoConflicts(t *testing.T) {
	cfg := &config.PortmapConfig{
		Defaults: config.ConfigDefaults{Range: "3000-9999", Strategy: "nearest"},
		Services: map[string]config.ServiceConfig{
			"api":      {Port: 3000},
			"frontend": {Port: 4000},
		},
		Groups: map[string]config.GroupConfig{},
	}

	// No ports active — no conflicts
	scn := &mockScanner{ports: []scanner.PortInfo{}}
	det := New(cfg, scn)

	result, err := det.Check()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Conflicts) != 0 {
		t.Errorf("conflicts = %d, want 0", len(result.Conflicts))
	}
	if len(result.OKServices) != 2 {
		t.Errorf("ok services = %d, want 2", len(result.OKServices))
	}
	if result.TotalChecked != 2 {
		t.Errorf("total checked = %d, want 2", result.TotalChecked)
	}
}

func TestCheck_OccupiedPort(t *testing.T) {
	cfg := &config.PortmapConfig{
		Defaults: config.ConfigDefaults{Range: "3000-9999", Strategy: "nearest"},
		Services: map[string]config.ServiceConfig{
			"api": {Port: 3000},
		},
		Groups: map[string]config.GroupConfig{},
	}

	// Port 3000 occupied by a different process
	scn := &mockScanner{
		ports: []scanner.PortInfo{
			{Port: 3000, Protocol: "tcp", PID: 999, ProcessName: "nginx", State: "LISTEN"},
		},
	}
	det := New(cfg, scn)

	result, err := det.Check()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Conflicts) != 1 {
		t.Fatalf("conflicts = %d, want 1", len(result.Conflicts))
	}
	c := result.Conflicts[0]
	if c.Type != ConflictOccupied {
		t.Errorf("type = %q, want %q", c.Type, ConflictOccupied)
	}
	if c.Port != 3000 {
		t.Errorf("port = %d, want 3000", c.Port)
	}
	if c.ActualProcess != "nginx" {
		t.Errorf("actual process = %q, want %q", c.ActualProcess, "nginx")
	}
	if c.ActualPID != 999 {
		t.Errorf("actual PID = %d, want 999", c.ActualPID)
	}
}

func TestCheck_OccupiedByExpectedProcess(t *testing.T) {
	cfg := &config.PortmapConfig{
		Defaults: config.ConfigDefaults{Range: "3000-9999", Strategy: "nearest"},
		Services: map[string]config.ServiceConfig{
			"nginx": {Port: 8080},
		},
		Groups: map[string]config.GroupConfig{},
	}

	// Port occupied by nginx itself — should be OK
	scn := &mockScanner{
		ports: []scanner.PortInfo{
			{Port: 8080, Protocol: "tcp", PID: 100, ProcessName: "nginx", State: "LISTEN"},
		},
	}
	det := New(cfg, scn)

	result, err := det.Check()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Conflicts) != 0 {
		t.Errorf("conflicts = %d, want 0 (expected process match)", len(result.Conflicts))
	}
	if len(result.OKServices) != 1 {
		t.Errorf("ok services = %d, want 1", len(result.OKServices))
	}
}

func TestCheck_DuplicateInConfig(t *testing.T) {
	// Bypass Validate() by building config directly
	cfg := &config.PortmapConfig{
		Defaults: config.ConfigDefaults{Range: "3000-9999", Strategy: "nearest"},
		Services: map[string]config.ServiceConfig{
			"svc1": {Port: 3000},
			"svc2": {Port: 3000},
		},
		Groups: map[string]config.GroupConfig{},
	}

	scn := &mockScanner{ports: []scanner.PortInfo{}}
	det := New(cfg, scn)

	result, err := det.Check()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should detect duplicate
	duplicateCount := 0
	for _, c := range result.Conflicts {
		if c.Type == ConflictDuplicate {
			duplicateCount++
		}
	}
	if duplicateCount != 2 {
		t.Errorf("duplicate conflicts = %d, want 2 (one per service)", duplicateCount)
	}
}

func TestCheck_ManagedFalseNotRunning(t *testing.T) {
	managed := false
	cfg := &config.PortmapConfig{
		Defaults: config.ConfigDefaults{Range: "3000-9999", Strategy: "nearest"},
		Services: map[string]config.ServiceConfig{
			"external-db": {Port: 5432, Managed: &managed},
		},
		Groups: map[string]config.GroupConfig{},
	}

	// Port not active
	scn := &mockScanner{ports: []scanner.PortInfo{}}
	det := New(cfg, scn)

	result, err := det.Check()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be in OK (not a conflict for unmanaged not running)
	if len(result.Conflicts) != 0 {
		t.Errorf("conflicts = %d, want 0", len(result.Conflicts))
	}
	if len(result.OKServices) != 1 {
		t.Errorf("ok services = %d, want 1", len(result.OKServices))
	}
}

func TestCheck_CommandMatch(t *testing.T) {
	cfg := &config.PortmapConfig{
		Defaults: config.ConfigDefaults{Range: "3000-9999", Strategy: "nearest"},
		Services: map[string]config.ServiceConfig{
			"myapi": {Port: 3000, Command: "node server.js"},
		},
		Groups: map[string]config.GroupConfig{},
	}

	// Process name is "node" which matches the command
	scn := &mockScanner{
		ports: []scanner.PortInfo{
			{Port: 3000, Protocol: "tcp", PID: 100, ProcessName: "node", State: "LISTEN"},
		},
	}
	det := New(cfg, scn)

	result, err := det.Check()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Conflicts) != 0 {
		t.Errorf("conflicts = %d, want 0 (command match)", len(result.Conflicts))
	}
}

func TestCheck_SuggestionProvided(t *testing.T) {
	cfg := &config.PortmapConfig{
		Defaults: config.ConfigDefaults{Range: "3000-9999", Strategy: "nearest"},
		Services: map[string]config.ServiceConfig{
			"api": {Port: 3000},
		},
		Groups: map[string]config.GroupConfig{},
	}

	scn := &mockScanner{
		ports: []scanner.PortInfo{
			{Port: 3000, Protocol: "tcp", PID: 999, ProcessName: "other", State: "LISTEN"},
		},
	}
	det := New(cfg, scn)

	result, err := det.Check()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Conflicts) != 1 {
		t.Fatalf("conflicts = %d, want 1", len(result.Conflicts))
	}
	if result.Conflicts[0].Suggestion <= 0 {
		t.Error("expected a port suggestion > 0")
	}
}

func TestCheckResult_HasConflicts(t *testing.T) {
	empty := &CheckResult{Conflicts: []Conflict{}}
	if empty.HasConflicts() {
		t.Error("HasConflicts() = true, want false for empty")
	}

	withConflict := &CheckResult{Conflicts: []Conflict{{Port: 3000}}}
	if !withConflict.HasConflicts() {
		t.Error("HasConflicts() = false, want true")
	}
}

func TestIsExpectedProcess(t *testing.T) {
	tests := []struct {
		name    string
		svcName string
		svc     config.ServiceConfig
		active  scanner.PortInfo
		want    bool
	}{
		{
			name:    "name match",
			svcName: "nginx",
			svc:     config.ServiceConfig{Port: 80},
			active:  scanner.PortInfo{ProcessName: "nginx"},
			want:    true,
		},
		{
			name:    "name match case insensitive",
			svcName: "Redis",
			svc:     config.ServiceConfig{Port: 6379},
			active:  scanner.PortInfo{ProcessName: "redis-server"},
			want:    true,
		},
		{
			name:    "command match",
			svcName: "myapp",
			svc:     config.ServiceConfig{Port: 3000, Command: "node server.js"},
			active:  scanner.PortInfo{ProcessName: "node"},
			want:    true,
		},
		{
			name:    "no match",
			svcName: "api",
			svc:     config.ServiceConfig{Port: 3000},
			active:  scanner.PortInfo{ProcessName: "nginx"},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isExpectedProcess(tt.svcName, tt.svc, tt.active)
			if got != tt.want {
				t.Errorf("isExpectedProcess() = %v, want %v", got, tt.want)
			}
		})
	}
}
