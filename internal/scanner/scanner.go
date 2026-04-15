package scanner

import (
	"runtime"
	"time"
)

// PortInfo holds information about a single port
type PortInfo struct {
	Port        int       `json:"port"`
	Protocol    string    `json:"protocol"`     // "tcp" | "udp"
	PID         int       `json:"pid"`          // 0 if unknown
	ProcessName string    `json:"process_name"` // "unknown" if not found
	User        string    `json:"user"`
	State       string    `json:"state"` // LISTEN | ESTABLISHED | TIME_WAIT | ...
	CommandLine string    `json:"command_line"`
	StartTime   time.Time `json:"start_time,omitempty"`
}

// ScanOptions configures what to scan
type ScanOptions struct {
	Port       int    // Specific port (0 = all)
	RangeStart int    // Start of range (default 1)
	RangeEnd   int    // End of range (default 65535)
	TCPOnly    bool   // Only TCP
	UDPOnly    bool   // Only UDP
	ListenOnly bool   // Only LISTEN state
	Filter     string // Process name filter (substring)
}

// ScanResult holds the results of a port scan
type ScanResult struct {
	Ports        []PortInfo    `json:"ports"`
	ScannedRange string        `json:"scanned_range"`
	Duration     time.Duration `json:"duration"`
	Platform     string        `json:"platform"`
	Total        int           `json:"total"`
}

// Scanner defines the port scanning interface
type Scanner interface {
	Scan(opts ScanOptions) (*ScanResult, error)
}

// New returns a platform-appropriate scanner
func New() Scanner {
	switch runtime.GOOS {
	case "linux":
		return &LinuxScanner{}
	case "darwin":
		return &DarwinScanner{}
	default:
		// Fallback to basic net.Listen probe
		return &LinuxScanner{} // TODO: implement fallback scanner
	}
}

// DefaultOptions returns scan options for common dev ports
func DefaultOptions() ScanOptions {
	return ScanOptions{
		RangeStart: 1,
		RangeEnd:   65535,
	}
}
