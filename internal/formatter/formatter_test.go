package formatter

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ozayartunboran/portmap/internal/detector"
	"github.com/ozayartunboran/portmap/internal/process"
	"github.com/ozayartunboran/portmap/internal/scanner"
)

func TestNew_ValidFormats(t *testing.T) {
	tests := []struct {
		format string
		typ    string
	}{
		{"terminal", "*formatter.TerminalFormatter"},
		{"json", "*formatter.JSONFormatter"},
		{"markdown", "*formatter.MarkdownFormatter"},
		{"compact", "*formatter.CompactFormatter"},
	}
	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			f, err := New(tt.format, true)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if f == nil {
				t.Fatal("formatter is nil")
			}
		})
	}
}

func TestNew_InvalidFormat(t *testing.T) {
	_, err := New("xml", true)
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

// --- Sample data for formatter tests ---

func sampleScanResult() *scanner.ScanResult {
	return &scanner.ScanResult{
		Ports: []scanner.PortInfo{
			{Port: 3000, Protocol: "tcp", PID: 1234, ProcessName: "node", User: "dev", State: "LISTEN"},
			{Port: 8080, Protocol: "tcp", PID: 5678, ProcessName: "go", User: "dev", State: "LISTEN"},
		},
		ScannedRange: "1-65535",
		Duration:     42 * time.Millisecond,
		Platform:     "linux",
		Total:        2,
	}
}

func sampleCheckResult_NoConflicts() *detector.CheckResult {
	return &detector.CheckResult{
		Conflicts:    []detector.Conflict{},
		OKServices:   []string{"api", "frontend"},
		TotalChecked: 2,
	}
}

func sampleCheckResult_WithConflicts() *detector.CheckResult {
	return &detector.CheckResult{
		Conflicts: []detector.Conflict{
			{
				Port:          3000,
				Type:          detector.ConflictOccupied,
				ServiceName:   "api",
				ActualProcess: "nginx",
				ActualPID:     999,
				Suggestion:    3001,
				Message:       "port 3000 expected by api but occupied by nginx",
			},
		},
		OKServices:   []string{"frontend"},
		TotalChecked: 2,
	}
}

func sampleProcessDetail() *process.ProcessDetail {
	return &process.ProcessDetail{
		PID:         1234,
		Name:        "node",
		CommandLine: "node server.js",
		User:        "dev",
		Ports:       []int{3000, 3001},
		CPUPercent:  2.5,
		MemoryMB:    128.3,
		StartTime:   time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
		Uptime:      2 * time.Hour,
	}
}

// --- Terminal Formatter Tests ---

func TestTerminalFormatter_FormatScan(t *testing.T) {
	f := &TerminalFormatter{NoColor: true}
	output := f.FormatScan(sampleScanResult())

	if !strings.Contains(output, "PORT") {
		t.Error("missing PORT header")
	}
	if !strings.Contains(output, "3000") {
		t.Error("missing port 3000")
	}
	if !strings.Contains(output, "8080") {
		t.Error("missing port 8080")
	}
	if !strings.Contains(output, "node") {
		t.Error("missing process name node")
	}
	if !strings.Contains(output, "2 ports found") {
		t.Error("missing total count")
	}
}

func TestTerminalFormatter_FormatCheck_NoConflicts(t *testing.T) {
	f := &TerminalFormatter{NoColor: true}
	output := f.FormatCheck(sampleCheckResult_NoConflicts())

	if !strings.Contains(output, "No conflicts") {
		t.Error("missing no conflicts message")
	}
	if !strings.Contains(output, "api") {
		t.Error("missing OK service name")
	}
}

func TestTerminalFormatter_FormatCheck_WithConflicts(t *testing.T) {
	f := &TerminalFormatter{NoColor: true}
	output := f.FormatCheck(sampleCheckResult_WithConflicts())

	if !strings.Contains(output, "1 conflict") {
		t.Error("missing conflict count")
	}
	if !strings.Contains(output, "OCCUPIED") {
		t.Error("missing conflict type")
	}
	if !strings.Contains(output, "nginx") {
		t.Error("missing actual process")
	}
	if !strings.Contains(output, "3001") {
		t.Error("missing suggestion")
	}
}

func TestTerminalFormatter_FormatInfo(t *testing.T) {
	f := &TerminalFormatter{NoColor: true}
	output := f.FormatInfo(sampleProcessDetail())

	if !strings.Contains(output, "1234") {
		t.Error("missing PID")
	}
	if !strings.Contains(output, "node") {
		t.Error("missing process name")
	}
	if !strings.Contains(output, "node server.js") {
		t.Error("missing command line")
	}
	if !strings.Contains(output, "dev") {
		t.Error("missing user")
	}
}

func TestTerminalFormatter_FormatInfo_Nil(t *testing.T) {
	f := &TerminalFormatter{NoColor: true}
	output := f.FormatInfo(nil)
	if !strings.Contains(output, "No process found") {
		t.Error("expected 'No process found' for nil detail")
	}
}

// --- JSON Formatter Tests ---

func TestJSONFormatter_FormatScan(t *testing.T) {
	f := &JSONFormatter{}
	output := f.FormatScan(sampleScanResult())

	var result scanner.ScanResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("total = %d, want 2", result.Total)
	}
	if len(result.Ports) != 2 {
		t.Errorf("ports count = %d, want 2", len(result.Ports))
	}
}

func TestJSONFormatter_FormatCheck(t *testing.T) {
	f := &JSONFormatter{}
	output := f.FormatCheck(sampleCheckResult_WithConflicts())

	var result detector.CheckResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(result.Conflicts) != 1 {
		t.Errorf("conflicts = %d, want 1", len(result.Conflicts))
	}
}

func TestJSONFormatter_FormatInfo(t *testing.T) {
	f := &JSONFormatter{}
	output := f.FormatInfo(sampleProcessDetail())

	var result process.ProcessDetail
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result.PID != 1234 {
		t.Errorf("PID = %d, want 1234", result.PID)
	}
}

func TestJSONFormatter_FormatInfo_Nil(t *testing.T) {
	f := &JSONFormatter{}
	output := f.FormatInfo(nil)
	if strings.TrimSpace(output) != "{}" {
		t.Errorf("expected empty JSON object, got %q", output)
	}
}

// --- Markdown Formatter Tests ---

func TestMarkdownFormatter_FormatScan(t *testing.T) {
	f := &MarkdownFormatter{}
	output := f.FormatScan(sampleScanResult())

	if !strings.Contains(output, "| Port |") {
		t.Error("missing markdown table header")
	}
	if !strings.Contains(output, "|---") {
		t.Error("missing markdown separator")
	}
	if !strings.Contains(output, "| 3000 |") {
		t.Error("missing port 3000 row")
	}
	if !strings.Contains(output, "2 ports found") {
		t.Error("missing total count")
	}
}

func TestMarkdownFormatter_FormatCheck_NoConflicts(t *testing.T) {
	f := &MarkdownFormatter{}
	output := f.FormatCheck(sampleCheckResult_NoConflicts())

	if !strings.Contains(output, "No conflicts") {
		t.Error("missing no conflicts message")
	}
}

func TestMarkdownFormatter_FormatCheck_WithConflicts(t *testing.T) {
	f := &MarkdownFormatter{}
	output := f.FormatCheck(sampleCheckResult_WithConflicts())

	if !strings.Contains(output, "1 conflict") {
		t.Error("missing conflict count")
	}
	if !strings.Contains(output, "| Port |") {
		t.Error("missing markdown table header")
	}
}

func TestMarkdownFormatter_FormatInfo(t *testing.T) {
	f := &MarkdownFormatter{}
	output := f.FormatInfo(sampleProcessDetail())

	if !strings.Contains(output, "**PID:**") {
		t.Error("missing PID label")
	}
	if !strings.Contains(output, "1234") {
		t.Error("missing PID value")
	}
}

func TestMarkdownFormatter_FormatInfo_Nil(t *testing.T) {
	f := &MarkdownFormatter{}
	output := f.FormatInfo(nil)
	if !strings.Contains(output, "No process found") {
		t.Error("expected 'No process found' for nil")
	}
}

// --- Compact Formatter Tests ---

func TestCompactFormatter_FormatScan(t *testing.T) {
	f := &CompactFormatter{}
	output := f.FormatScan(sampleScanResult())

	if !strings.Contains(output, "3000:node(1234)") {
		t.Errorf("missing compact port entry, got %q", output)
	}
	if !strings.Contains(output, "8080:go(5678)") {
		t.Errorf("missing compact port entry, got %q", output)
	}
}

func TestCompactFormatter_FormatScan_Empty(t *testing.T) {
	f := &CompactFormatter{}
	result := &scanner.ScanResult{Ports: []scanner.PortInfo{}, Total: 0}
	output := f.FormatScan(result)
	if strings.TrimSpace(output) != "" {
		t.Errorf("expected empty output for no ports, got %q", output)
	}
}

func TestCompactFormatter_FormatCheck_NoConflicts(t *testing.T) {
	f := &CompactFormatter{}
	output := f.FormatCheck(sampleCheckResult_NoConflicts())
	if strings.TrimSpace(output) != "ok" {
		t.Errorf("expected 'ok', got %q", output)
	}
}

func TestCompactFormatter_FormatCheck_WithConflicts(t *testing.T) {
	f := &CompactFormatter{}
	output := f.FormatCheck(sampleCheckResult_WithConflicts())

	if !strings.Contains(output, "3000:OCCUPIED:api") {
		t.Errorf("missing compact conflict entry, got %q", output)
	}
}

func TestCompactFormatter_FormatInfo(t *testing.T) {
	f := &CompactFormatter{}
	output := f.FormatInfo(sampleProcessDetail())

	if !strings.Contains(output, "1234:node:") {
		t.Errorf("missing compact info, got %q", output)
	}
}

func TestCompactFormatter_FormatInfo_Nil(t *testing.T) {
	f := &CompactFormatter{}
	output := f.FormatInfo(nil)
	if strings.TrimSpace(output) != "none" {
		t.Errorf("expected 'none', got %q", output)
	}
}
