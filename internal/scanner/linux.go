package scanner

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// LinuxScanner reads port info from /proc/net/tcp and /proc/net/udp
type LinuxScanner struct{}

// tcpState maps kernel state codes to human-readable strings
var tcpState = map[string]string{
	"01": "ESTABLISHED",
	"02": "SYN_SENT",
	"03": "SYN_RECV",
	"04": "FIN_WAIT1",
	"05": "FIN_WAIT2",
	"06": "TIME_WAIT",
	"07": "CLOSE",
	"08": "CLOSE_WAIT",
	"09": "LAST_ACK",
	"0A": "LISTEN",
	"0B": "CLOSING",
}

// procNetEntry represents a parsed line from /proc/net/tcp or /proc/net/udp
type procNetEntry struct {
	localPort int
	state     string
	inode     uint64
}

// Scan scans active ports on Linux by parsing /proc/net/tcp and /proc/net/udp
func (s *LinuxScanner) Scan(opts ScanOptions) (*ScanResult, error) {
	start := time.Now()

	// Build inode→PID map
	inodePID := buildInodePIDMap()

	var ports []PortInfo

	if !opts.UDPOnly {
		tcpPorts, err := parseProcNet("/proc/net/tcp", "tcp", inodePID, opts)
		if err == nil {
			ports = append(ports, tcpPorts...)
		}
		tcpPorts6, err := parseProcNet("/proc/net/tcp6", "tcp", inodePID, opts)
		if err == nil {
			ports = append(ports, tcpPorts6...)
		}
	}

	if !opts.TCPOnly {
		udpPorts, err := parseProcNet("/proc/net/udp", "udp", inodePID, opts)
		if err == nil {
			ports = append(ports, udpPorts...)
		}
		udpPorts6, err := parseProcNet("/proc/net/udp6", "udp", inodePID, opts)
		if err == nil {
			ports = append(ports, udpPorts6...)
		}
	}

	// Deduplicate by port+protocol+pid (IPv4 and IPv6 can show same port)
	ports = deduplicatePorts(ports)

	result := &ScanResult{
		Ports:        ports,
		ScannedRange: formatRange(opts),
		Duration:     time.Since(start),
		Platform:     "linux",
		Total:        len(ports),
	}
	return result, nil
}

// parseProcNet parses a /proc/net/tcp or /proc/net/udp file
func parseProcNet(path, protocol string, inodePID map[uint64]int, opts ScanOptions) ([]PortInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var ports []PortInfo
	sc := bufio.NewScanner(f)
	first := true
	for sc.Scan() {
		if first {
			first = false
			continue // skip header
		}
		line := strings.TrimSpace(sc.Text())
		entry, err := parseProcNetLine(line, protocol)
		if err != nil {
			continue
		}

		// Apply filters
		if !matchesFilters(entry, opts) {
			continue
		}

		pid := inodePID[entry.inode]
		info := PortInfo{
			Port:     entry.localPort,
			Protocol: protocol,
			PID:      pid,
			State:    entry.state,
		}

		if pid > 0 {
			fillProcessInfo(&info, pid)
		} else {
			info.ProcessName = "unknown (requires root)"
			info.User = "-"
		}

		// Apply process name filter
		if opts.Filter != "" && !strings.Contains(strings.ToLower(info.ProcessName), strings.ToLower(opts.Filter)) {
			continue
		}

		ports = append(ports, info)
	}

	return ports, nil
}

// parseProcNetLine parses a single line from /proc/net/tcp or /proc/net/udp
func parseProcNetLine(line, protocol string) (procNetEntry, error) {
	fields := strings.Fields(line)
	if len(fields) < 10 {
		return procNetEntry{}, fmt.Errorf("too few fields")
	}

	// local_address is field[1], format: hex_ip:hex_port
	localAddr := fields[1]
	parts := strings.SplitN(localAddr, ":", 2)
	if len(parts) != 2 {
		return procNetEntry{}, fmt.Errorf("invalid local_address")
	}
	port, err := strconv.ParseInt(parts[1], 16, 32)
	if err != nil {
		return procNetEntry{}, fmt.Errorf("parse port: %w", err)
	}

	// State is field[3]
	stateHex := strings.ToUpper(fields[3])
	state := tcpState[stateHex]
	if state == "" {
		state = "UNKNOWN"
	}
	// For UDP, remap states
	if protocol == "udp" {
		switch stateHex {
		case "07":
			state = "UNCONN"
		case "01":
			state = "ESTABLISHED"
		default:
			state = "UNCONN"
		}
	}

	// Inode is field[9]
	inode, err := strconv.ParseUint(fields[9], 10, 64)
	if err != nil {
		return procNetEntry{}, fmt.Errorf("parse inode: %w", err)
	}

	return procNetEntry{
		localPort: int(port),
		state:     state,
		inode:     inode,
	}, nil
}

// matchesFilters checks if a proc net entry matches the scan options
func matchesFilters(entry procNetEntry, opts ScanOptions) bool {
	if opts.Port > 0 && entry.localPort != opts.Port {
		return false
	}
	rangeStart := opts.RangeStart
	if rangeStart == 0 {
		rangeStart = 1
	}
	rangeEnd := opts.RangeEnd
	if rangeEnd == 0 {
		rangeEnd = 65535
	}
	if entry.localPort < rangeStart || entry.localPort > rangeEnd {
		return false
	}
	if opts.ListenOnly && entry.state != "LISTEN" && entry.state != "UNCONN" {
		return false
	}
	return true
}

// buildInodePIDMap scans /proc/*/fd/* to build a socket inode → PID map
func buildInodePIDMap() map[uint64]int {
	result := make(map[uint64]int)

	entries, err := os.ReadDir("/proc")
	if err != nil {
		return result
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		fdDir := filepath.Join("/proc", entry.Name(), "fd")
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue // permission denied or process gone
		}

		for _, fd := range fds {
			link, err := os.Readlink(filepath.Join(fdDir, fd.Name()))
			if err != nil {
				continue
			}
			// Format: socket:[inode]
			if strings.HasPrefix(link, "socket:[") && strings.HasSuffix(link, "]") {
				inodeStr := link[8 : len(link)-1]
				inode, err := strconv.ParseUint(inodeStr, 10, 64)
				if err != nil {
					continue
				}
				result[inode] = pid
			}
		}
	}

	return result
}

// fillProcessInfo populates PortInfo fields from /proc/{pid}/*
func fillProcessInfo(info *PortInfo, pid int) {
	pidStr := strconv.Itoa(pid)
	procDir := filepath.Join("/proc", pidStr)

	// Process name from /proc/{pid}/comm
	if data, err := os.ReadFile(filepath.Join(procDir, "comm")); err == nil {
		info.ProcessName = strings.TrimSpace(string(data))
	} else {
		info.ProcessName = "unknown"
	}

	// Command line from /proc/{pid}/cmdline (null-separated)
	if data, err := os.ReadFile(filepath.Join(procDir, "cmdline")); err == nil {
		cmdline := strings.ReplaceAll(string(data), "\x00", " ")
		info.CommandLine = strings.TrimSpace(cmdline)
	}

	// User from /proc/{pid}/status Uid line
	if data, err := os.ReadFile(filepath.Join(procDir, "status")); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "Uid:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					if u, err := user.LookupId(fields[1]); err == nil {
						info.User = u.Username
					} else {
						info.User = fields[1]
					}
				}
				break
			}
		}
	}
	if info.User == "" {
		info.User = "-"
	}

	// Start time from /proc/{pid}/stat field 22
	if data, err := os.ReadFile(filepath.Join(procDir, "stat")); err == nil {
		info.StartTime = parseStartTime(string(data))
	}
}

// parseStartTime extracts start time from /proc/{pid}/stat
func parseStartTime(statContent string) time.Time {
	// The comm field can contain spaces and parens, so find the last ')' first
	idx := strings.LastIndex(statContent, ")")
	if idx < 0 || idx+2 >= len(statContent) {
		return time.Time{}
	}
	// Fields after ')' start at index 2 (0-based from the full stat), field 22 is starttime
	// After ')', fields are space-separated. Field 0 after ')' is state (field 3 overall).
	// starttime is field 22 overall = field 19 after ')'
	rest := strings.Fields(statContent[idx+2:])
	if len(rest) < 20 {
		return time.Time{}
	}
	startTicks, err := strconv.ParseUint(rest[19], 10, 64)
	if err != nil {
		return time.Time{}
	}

	// Get system boot time
	bootTime := getBootTime()
	if bootTime.IsZero() {
		return time.Time{}
	}

	// Clock ticks per second (usually 100 on Linux)
	clkTck := uint64(100)
	startSec := startTicks / clkTck
	return bootTime.Add(time.Duration(startSec) * time.Second)
}

// getBootTime reads /proc/stat for boot time
func getBootTime() time.Time {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return time.Time{}
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "btime ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				btime, err := strconv.ParseInt(fields[1], 10, 64)
				if err == nil {
					return time.Unix(btime, 0)
				}
			}
		}
	}
	return time.Time{}
}

// deduplicatePorts removes duplicate entries (same port+protocol+pid)
func deduplicatePorts(ports []PortInfo) []PortInfo {
	type key struct {
		port     int
		protocol string
		pid      int
		state    string
	}
	seen := make(map[key]bool)
	var result []PortInfo
	for _, p := range ports {
		k := key{p.Port, p.Protocol, p.PID, p.State}
		if seen[k] {
			continue
		}
		seen[k] = true
		result = append(result, p)
	}
	// Return empty slice instead of nil
	if result == nil {
		return []PortInfo{}
	}
	return result
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

