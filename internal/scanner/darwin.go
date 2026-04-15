package scanner

// DarwinScanner reads port info using lsof on macOS
type DarwinScanner struct{}

// Scan scans active ports on macOS using lsof -iTCP -iUDP -nP
func (s *DarwinScanner) Scan(opts ScanOptions) (*ScanResult, error) {
	// TODO: Phase 2 — implement:
	// 1. Run `lsof -iTCP -iUDP -nP` and parse output
	// 2. Each line: COMMAND, PID, USER, FD, TYPE, DEVICE, SIZE, NODE, NAME
	// 3. Apply filters from opts
	// 4. Return ScanResult

	result := &ScanResult{
		Ports:        []PortInfo{},
		ScannedRange: formatRange(opts),
		Platform:     "darwin",
	}
	return result, nil
}
