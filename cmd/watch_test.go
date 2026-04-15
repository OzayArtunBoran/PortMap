package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/ozayartunboran/portmap/internal/scanner"
)

func TestRenderWatchTable(t *testing.T) {
	result := &scanner.ScanResult{
		Ports: []scanner.PortInfo{
			{Port: 8080, Protocol: "tcp", PID: 1234, ProcessName: "node", User: "dev", State: "LISTEN"},
			{Port: 3000, Protocol: "tcp", PID: 5678, ProcessName: "rails", User: "dev", State: "LISTEN"},
		},
		ScannedRange: "1-65535",
		Duration:     50 * time.Millisecond,
		Platform:     "linux",
		Total:        2,
	}

	output := renderWatchTable(result, nil, "1-65535", true)

	// Should contain header
	if !strings.Contains(output, "PORT") {
		t.Error("missing PORT header")
	}
	if !strings.Contains(output, "PROCESS") {
		t.Error("missing PROCESS header")
	}

	// Should contain port entries
	if !strings.Contains(output, "8080") {
		t.Error("missing port 8080")
	}
	if !strings.Contains(output, "3000") {
		t.Error("missing port 3000")
	}
	if !strings.Contains(output, "node") {
		t.Error("missing process name node")
	}

	// Should contain status bar
	if !strings.Contains(output, "q to quit") {
		t.Error("missing status bar")
	}
	if !strings.Contains(output, "2 ports") {
		t.Error("missing port count")
	}
}

func TestRenderWatchTable_WithHighlights(t *testing.T) {
	result := &scanner.ScanResult{
		Ports: []scanner.PortInfo{
			{Port: 8080, Protocol: "tcp", PID: 1234, ProcessName: "node", User: "dev", State: "LISTEN"},
		},
		Total: 1,
	}

	highlights := map[portKey]*highlightEntry{
		{Port: 8080, Protocol: "tcp", PID: 1234}: {
			info:      result.Ports[0],
			remaining: 3,
			closed:    false,
		},
		{Port: 3000, Protocol: "tcp", PID: 5678}: {
			info:      scanner.PortInfo{Port: 3000, Protocol: "tcp", PID: 5678, ProcessName: "rails", User: "dev", State: "LISTEN"},
			remaining: 1,
			closed:    true,
		},
	}

	// With color enabled, should contain ANSI codes
	output := renderWatchTable(result, highlights, "1-65535", false)
	if !strings.Contains(output, "\033[") {
		t.Error("expected ANSI codes when color enabled")
	}

	// With color disabled, should NOT contain ANSI codes
	output = renderWatchTable(result, highlights, "1-65535", true)
	if strings.Contains(output, "\033[") {
		t.Error("unexpected ANSI codes when color disabled")
	}

	// Closed port should still appear in output
	if !strings.Contains(output, "3000") {
		t.Error("closed port 3000 should still appear during highlight")
	}
}

func TestRenderWatchTable_Empty(t *testing.T) {
	result := &scanner.ScanResult{
		Ports: []scanner.PortInfo{},
		Total: 0,
	}

	output := renderWatchTable(result, nil, "8000-9000", true)
	if !strings.Contains(output, "0 ports") {
		t.Error("should show 0 ports")
	}
	if !strings.Contains(output, "8000-9000") {
		t.Error("should show watched range")
	}
}

func TestRenderWatchTable_SortedByPort(t *testing.T) {
	result := &scanner.ScanResult{
		Ports: []scanner.PortInfo{
			{Port: 9000, Protocol: "tcp", PID: 1, ProcessName: "c", User: "u", State: "LISTEN"},
			{Port: 3000, Protocol: "tcp", PID: 2, ProcessName: "a", User: "u", State: "LISTEN"},
			{Port: 5000, Protocol: "tcp", PID: 3, ProcessName: "b", User: "u", State: "LISTEN"},
		},
		Total: 3,
	}

	output := renderWatchTable(result, nil, "1-65535", true)

	idx3000 := strings.Index(output, "3000")
	idx5000 := strings.Index(output, "5000")
	idx9000 := strings.Index(output, "9000")

	if idx3000 > idx5000 || idx5000 > idx9000 {
		t.Error("ports should be sorted in ascending order")
	}
}

func TestParseWatchRange(t *testing.T) {
	tests := []struct {
		input     string
		wantStart int
		wantEnd   int
		wantPort  int
		wantErr   bool
	}{
		{"1-65535", 1, 65535, 0, false},
		{"8000-9000", 8000, 9000, 0, false},
		{"3000", 0, 0, 3000, false},
		{"abc", 0, 0, 0, true},
		{"9000-8000", 0, 0, 0, true},
		{"-1-100", 0, 0, 0, true},
	}

	for _, tt := range tests {
		opts, err := parseWatchRange(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseWatchRange(%q): err=%v, wantErr=%v", tt.input, err, tt.wantErr)
			continue
		}
		if err != nil {
			continue
		}
		if tt.wantPort > 0 {
			if opts.Port != tt.wantPort {
				t.Errorf("parseWatchRange(%q): port=%d, want=%d", tt.input, opts.Port, tt.wantPort)
			}
		} else {
			if opts.RangeStart != tt.wantStart || opts.RangeEnd != tt.wantEnd {
				t.Errorf("parseWatchRange(%q): range=%d-%d, want=%d-%d", tt.input, opts.RangeStart, opts.RangeEnd, tt.wantStart, tt.wantEnd)
			}
		}
	}
}
