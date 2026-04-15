package scanner

import "fmt"

// LinuxScanner reads port info from /proc/net/tcp and /proc/net/udp
type LinuxScanner struct{}

// Scan scans active ports on Linux by parsing /proc/net/tcp and /proc/net/udp
func (s *LinuxScanner) Scan(opts ScanOptions) (*ScanResult, error) {
	// TODO: Phase 2 — implement:
	// 1. Parse /proc/net/tcp for TCP ports
	// 2. Parse /proc/net/udp for UDP ports
	// 3. Map inode → PID via /proc/{pid}/fd/
	// 4. Get process name from /proc/{pid}/comm
	// 5. Get command line from /proc/{pid}/cmdline
	// 6. Apply filters from opts
	// 7. Return ScanResult

	result := &ScanResult{
		Ports:        []PortInfo{},
		ScannedRange: formatRange(opts),
		Platform:     "linux",
	}
	return result, nil
}

func formatRange(opts ScanOptions) string {
	if opts.Port > 0 {
		return fmt.Sprintf("%d", opts.Port)
	}
	start := opts.RangeStart
	if start == 0 {
		start = 1
	}
	end := opts.RangeEnd
	if end == 0 {
		end = 65535
	}
	return fmt.Sprintf("%d-%d", start, end)
}
