package scanner

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// DarwinScanner reads port info using lsof on macOS
type DarwinScanner struct{}

// Scan scans active ports on macOS using lsof -iTCP -iUDP -nP
func (s *DarwinScanner) Scan(opts ScanOptions) (*ScanResult, error) {
	start := time.Now()

	lsofPath, err := exec.LookPath("lsof")
	if err != nil {
		return nil, fmt.Errorf("lsof not found, install with: brew install lsof")
	}

	args := []string{"-nP"}
	if opts.TCPOnly {
		args = append(args, "-iTCP")
	} else if opts.UDPOnly {
		args = append(args, "-iUDP")
	} else {
		args = append(args, "-iTCP", "-iUDP")
	}

	cmd := exec.Command(lsofPath, args...)
	output, err := cmd.Output()
	if err != nil {
		// lsof returns exit code 1 when no results found, that's ok
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return &ScanResult{
				Ports:        []PortInfo{},
				ScannedRange: formatRange(opts),
				Duration:     time.Since(start),
				Platform:     "darwin",
				Total:        0,
			}, nil
		}
		return nil, fmt.Errorf("lsof failed: %w", err)
	}

	ports := parseLsofOutput(string(output), opts)

	result := &ScanResult{
		Ports:        ports,
		ScannedRange: formatRange(opts),
		Duration:     time.Since(start),
		Platform:     "darwin",
		Total:        len(ports),
	}
	return result, nil
}

// parseLsofOutput parses the output of lsof -iTCP -iUDP -nP
func parseLsofOutput(output string, opts ScanOptions) []PortInfo {
	var ports []PortInfo
	sc := bufio.NewScanner(strings.NewReader(output))
	first := true
	for sc.Scan() {
		if first {
			first = false
			continue // skip header
		}
		line := sc.Text()
		info, ok := parseLsofLine(line, opts)
		if ok {
			ports = append(ports, info)
		}
	}
	if ports == nil {
		return []PortInfo{}
	}
	return deduplicatePorts(ports)
}

// parseLsofLine parses a single lsof output line
// Format: COMMAND PID USER FD TYPE DEVICE SIZE/OFF NODE NAME
func parseLsofLine(line string, opts ScanOptions) (PortInfo, bool) {
	fields := strings.Fields(line)
	if len(fields) < 9 {
		return PortInfo{}, false
	}

	command := fields[0]
	pid, err := strconv.Atoi(fields[1])
	if err != nil {
		return PortInfo{}, false
	}
	userName := fields[2]
	nodeType := strings.ToLower(fields[7]) // TCP or UDP
	name := fields[8]
	// Sometimes there's a state in parens after name
	state := ""
	if len(fields) > 9 {
		s := fields[len(fields)-1]
		s = strings.Trim(s, "()")
		state = s
	}

	// Determine protocol
	protocol := ""
	if strings.Contains(nodeType, "tcp") || strings.EqualFold(fields[4], "IPv4") || strings.EqualFold(fields[4], "IPv6") {
		if strings.Contains(strings.ToLower(fields[7]), "tcp") {
			protocol = "tcp"
		} else if strings.Contains(strings.ToLower(fields[7]), "udp") {
			protocol = "udp"
		}
	}
	if protocol == "" {
		// Try from TYPE column
		if strings.EqualFold(nodeType, "tcp") {
			protocol = "tcp"
		} else if strings.EqualFold(nodeType, "udp") {
			protocol = "udp"
		} else {
			return PortInfo{}, false
		}
	}

	// Parse port from NAME field
	// Formats: *:port, host:port, host:port->host:port
	port := extractPortFromName(name)
	if port <= 0 {
		return PortInfo{}, false
	}

	// Apply filters
	if opts.Port > 0 && port != opts.Port {
		return PortInfo{}, false
	}
	rangeStart := opts.RangeStart
	if rangeStart == 0 {
		rangeStart = 1
	}
	rangeEnd := opts.RangeEnd
	if rangeEnd == 0 {
		rangeEnd = 65535
	}
	if port < rangeStart || port > rangeEnd {
		return PortInfo{}, false
	}
	if opts.ListenOnly && !strings.EqualFold(state, "LISTEN") {
		return PortInfo{}, false
	}
	if opts.Filter != "" && !strings.Contains(strings.ToLower(command), strings.ToLower(opts.Filter)) {
		return PortInfo{}, false
	}

	if state == "" {
		if protocol == "udp" {
			state = "UNCONN"
		} else {
			state = "ESTABLISHED"
		}
	}

	info := PortInfo{
		Port:        port,
		Protocol:    protocol,
		PID:         pid,
		ProcessName: command,
		User:        userName,
		State:       state,
	}

	return info, true
}

// extractPortFromName extracts the local port from lsof NAME field
func extractPortFromName(name string) int {
	// Handle connection format: host:port->host:port
	if idx := strings.Index(name, "->"); idx >= 0 {
		name = name[:idx]
	}
	// Find last colon
	idx := strings.LastIndex(name, ":")
	if idx < 0 {
		return 0
	}
	portStr := name[idx+1:]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0
	}
	return port
}
