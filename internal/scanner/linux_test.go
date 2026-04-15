package scanner

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func skipIfNotLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("skipping Linux-specific test on", runtime.GOOS)
	}
}

func TestParseProcNetLine_Valid(t *testing.T) {
	skipIfNotLinux(t)

	tests := []struct {
		name     string
		line     string
		protocol string
		wantPort int
		wantState string
	}{
		{
			name:      "LISTEN on port 8080",
			line:      "   0: 00000000:1F90 00000000:0000 0A 00000000:00000000 00:00000000 00000000   1000        0 12345 1 0000000000000000 100 0 0 10 0",
			protocol:  "tcp",
			wantPort:  8080,
			wantState: "LISTEN",
		},
		{
			name:      "ESTABLISHED on port 3000",
			line:      "   1: 0100007F:0BB8 0100007F:C5E0 01 00000000:00000000 00:00000000 00000000   1000        0 23456 1 0000000000000000 100 0 0 10 0",
			protocol:  "tcp",
			wantPort:  3000,
			wantState: "ESTABLISHED",
		},
		{
			name:      "UDP UNCONN on port 5353",
			line:      "   2: 00000000:14E9 00000000:0000 07 00000000:00000000 00:00000000 00000000     0        0 34567 2 0000000000000000 0",
			protocol:  "udp",
			wantPort:  5353,
			wantState: "UNCONN",
		},
		{
			name:      "port 22 SSH",
			line:      "   3: 00000000:0016 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 45678 1 0000000000000000 100 0 0 10 0",
			protocol:  "tcp",
			wantPort:  22,
			wantState: "LISTEN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := parseProcNetLine(tt.line, tt.protocol)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if entry.localPort != tt.wantPort {
				t.Errorf("port = %d, want %d", entry.localPort, tt.wantPort)
			}
			if entry.state != tt.wantState {
				t.Errorf("state = %q, want %q", entry.state, tt.wantState)
			}
		})
	}
}

func TestParseProcNetLine_Invalid(t *testing.T) {
	skipIfNotLinux(t)

	tests := []struct {
		name string
		line string
	}{
		{"empty line", ""},
		{"too few fields", "0: 00000000:1F90 00000000:0000"},
		{"invalid local address", "   0: badformat 00000000:0000 0A 00000000:00000000 00:00000000 00000000 0 0 12345 1"},
		{"invalid port hex", "   0: 00000000:ZZZZ 00000000:0000 0A 00000000:00000000 00:00000000 00000000 0 0 12345 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseProcNetLine(tt.line, "tcp")
			if err == nil {
				t.Error("expected error for invalid line")
			}
		})
	}
}

func TestParseProcNet_WithTempFile(t *testing.T) {
	skipIfNotLinux(t)

	tmpDir := t.TempDir()
	procFile := filepath.Join(tmpDir, "tcp")

	content := `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 00000000:1F90 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 12345 1 0000000000000000 100 0 0 10 0
   1: 0100007F:0BB8 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 23456 1 0000000000000000 100 0 0 10 0
`
	os.WriteFile(procFile, []byte(content), 0644)

	inodePID := map[uint64]int{
		12345: 100,
		23456: 200,
	}

	ports, err := parseProcNet(procFile, "tcp", inodePID, ScanOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 2 {
		t.Fatalf("got %d ports, want 2", len(ports))
	}
	if ports[0].Port != 8080 {
		t.Errorf("first port = %d, want 8080", ports[0].Port)
	}
	if ports[1].Port != 3000 {
		t.Errorf("second port = %d, want 3000", ports[1].Port)
	}
}

func TestParseProcNet_EmptyFile(t *testing.T) {
	skipIfNotLinux(t)

	tmpDir := t.TempDir()
	procFile := filepath.Join(tmpDir, "tcp")
	os.WriteFile(procFile, []byte(""), 0644)

	ports, err := parseProcNet(procFile, "tcp", map[uint64]int{}, ScanOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 0 {
		t.Errorf("got %d ports, want 0", len(ports))
	}
}

func TestParseProcNet_HeaderOnly(t *testing.T) {
	skipIfNotLinux(t)

	tmpDir := t.TempDir()
	procFile := filepath.Join(tmpDir, "tcp")
	content := "  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode\n"
	os.WriteFile(procFile, []byte(content), 0644)

	ports, err := parseProcNet(procFile, "tcp", map[uint64]int{}, ScanOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 0 {
		t.Errorf("got %d ports, want 0", len(ports))
	}
}

func TestParseProcNet_FileNotFound(t *testing.T) {
	skipIfNotLinux(t)

	_, err := parseProcNet("/nonexistent/tcp", "tcp", map[uint64]int{}, ScanOptions{})
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestMatchesFilters(t *testing.T) {
	skipIfNotLinux(t)

	tests := []struct {
		name  string
		entry procNetEntry
		opts  ScanOptions
		want  bool
	}{
		{
			name:  "no filters match all",
			entry: procNetEntry{localPort: 8080, state: "LISTEN"},
			opts:  ScanOptions{},
			want:  true,
		},
		{
			name:  "specific port match",
			entry: procNetEntry{localPort: 8080, state: "LISTEN"},
			opts:  ScanOptions{Port: 8080},
			want:  true,
		},
		{
			name:  "specific port no match",
			entry: procNetEntry{localPort: 8080, state: "LISTEN"},
			opts:  ScanOptions{Port: 3000},
			want:  false,
		},
		{
			name:  "in range",
			entry: procNetEntry{localPort: 5000, state: "LISTEN"},
			opts:  ScanOptions{RangeStart: 3000, RangeEnd: 9999},
			want:  true,
		},
		{
			name:  "out of range",
			entry: procNetEntry{localPort: 80, state: "LISTEN"},
			opts:  ScanOptions{RangeStart: 3000, RangeEnd: 9999},
			want:  false,
		},
		{
			name:  "listen only, state LISTEN",
			entry: procNetEntry{localPort: 8080, state: "LISTEN"},
			opts:  ScanOptions{ListenOnly: true},
			want:  true,
		},
		{
			name:  "listen only, state ESTABLISHED",
			entry: procNetEntry{localPort: 8080, state: "ESTABLISHED"},
			opts:  ScanOptions{ListenOnly: true},
			want:  false,
		},
		{
			name:  "listen only, state UNCONN (UDP)",
			entry: procNetEntry{localPort: 5353, state: "UNCONN"},
			opts:  ScanOptions{ListenOnly: true},
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesFilters(tt.entry, tt.opts)
			if got != tt.want {
				t.Errorf("matchesFilters() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeduplicatePorts(t *testing.T) {
	skipIfNotLinux(t)

	ports := []PortInfo{
		{Port: 8080, Protocol: "tcp", PID: 100, State: "LISTEN"},
		{Port: 8080, Protocol: "tcp", PID: 100, State: "LISTEN"}, // duplicate
		{Port: 3000, Protocol: "tcp", PID: 200, State: "LISTEN"},
		{Port: 8080, Protocol: "tcp", PID: 100, State: "ESTABLISHED"}, // different state
	}

	result := deduplicatePorts(ports)
	if len(result) != 3 {
		t.Errorf("got %d ports, want 3", len(result))
	}
}

func TestDeduplicatePorts_Empty(t *testing.T) {
	skipIfNotLinux(t)

	result := deduplicatePorts(nil)
	if result == nil {
		t.Error("expected non-nil empty slice")
	}
	if len(result) != 0 {
		t.Errorf("got %d ports, want 0", len(result))
	}
}

func TestFormatRange(t *testing.T) {
	skipIfNotLinux(t)

	tests := []struct {
		name string
		opts ScanOptions
		want string
	}{
		{"specific port", ScanOptions{Port: 8080}, "8080"},
		{"default range", ScanOptions{}, "1-65535"},
		{"custom range", ScanOptions{RangeStart: 3000, RangeEnd: 9999}, "3000-9999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatRange(tt.opts)
			if got != tt.want {
				t.Errorf("formatRange() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTcpStateMap(t *testing.T) {
	skipIfNotLinux(t)

	expected := map[string]string{
		"01": "ESTABLISHED",
		"0A": "LISTEN",
		"06": "TIME_WAIT",
		"08": "CLOSE_WAIT",
	}

	for hex, want := range expected {
		got := tcpState[hex]
		if got != want {
			t.Errorf("tcpState[%q] = %q, want %q", hex, got, want)
		}
	}
}

func TestParseProcNet_WithRangeFilter(t *testing.T) {
	skipIfNotLinux(t)

	tmpDir := t.TempDir()
	procFile := filepath.Join(tmpDir, "tcp")

	// Port 22 (0x0016), port 8080 (0x1F90), port 3000 (0x0BB8)
	content := `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 00000000:0016 00000000:0000 0A 00000000:00000000 00:00000000 00000000  0        0 11111 1 0000000000000000 100 0 0 10 0
   1: 00000000:1F90 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 22222 1 0000000000000000 100 0 0 10 0
   2: 00000000:0BB8 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 33333 1 0000000000000000 100 0 0 10 0
`
	os.WriteFile(procFile, []byte(content), 0644)

	// Filter: only ports 3000-9999
	ports, err := parseProcNet(procFile, "tcp", map[uint64]int{}, ScanOptions{RangeStart: 3000, RangeEnd: 9999})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 2 {
		t.Fatalf("got %d ports, want 2 (port 22 should be filtered)", len(ports))
	}
}
