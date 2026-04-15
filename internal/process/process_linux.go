package process

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// GetByPID retrieves detailed process information on Linux
func GetByPID(pid int) (*ProcessDetail, error) {
	pidStr := strconv.Itoa(pid)
	procDir := filepath.Join("/proc", pidStr)

	// Check if process exists
	if _, err := os.Stat(procDir); err != nil {
		return nil, fmt.Errorf("process %d not found", pid)
	}

	detail := &ProcessDetail{
		PID: pid,
	}

	// Process name from /proc/{pid}/comm
	if data, err := os.ReadFile(filepath.Join(procDir, "comm")); err == nil {
		detail.Name = strings.TrimSpace(string(data))
	}

	// Command line from /proc/{pid}/cmdline
	if data, err := os.ReadFile(filepath.Join(procDir, "cmdline")); err == nil {
		detail.CommandLine = strings.TrimSpace(strings.ReplaceAll(string(data), "\x00", " "))
	}

	// Parse /proc/{pid}/status for User and Memory
	if data, err := os.ReadFile(filepath.Join(procDir, "status")); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "Uid:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					if u, err := user.LookupId(fields[1]); err == nil {
						detail.User = u.Username
					} else {
						detail.User = fields[1]
					}
				}
			}
			if strings.HasPrefix(line, "VmRSS:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					if kb, err := strconv.ParseFloat(fields[1], 64); err == nil {
						detail.MemoryMB = kb / 1024.0
					}
				}
			}
		}
	}

	// Parse /proc/{pid}/stat for CPU and start time
	if data, err := os.ReadFile(filepath.Join(procDir, "stat")); err == nil {
		statContent := string(data)
		idx := strings.LastIndex(statContent, ")")
		if idx >= 0 && idx+2 < len(statContent) {
			rest := strings.Fields(statContent[idx+2:])
			if len(rest) >= 20 {
				// utime = field 14 overall = rest[11], stime = field 15 = rest[12]
				utime, _ := strconv.ParseUint(rest[11], 10, 64)
				stime, _ := strconv.ParseUint(rest[12], 10, 64)
				clkTck := uint64(100) // sysconf(_SC_CLK_TCK) usually 100
				totalSec := (utime + stime) / clkTck

				// start time = rest[19] (field 22 overall)
				startTicks, err := strconv.ParseUint(rest[19], 10, 64)
				if err == nil {
					bootTime := getBootTime()
					if !bootTime.IsZero() {
						startSec := startTicks / clkTck
						detail.StartTime = bootTime.Add(time.Duration(startSec) * time.Second)
						detail.Uptime = time.Since(detail.StartTime)
					}
				}

				// Rough CPU estimate: total CPU seconds / uptime
				if detail.Uptime > 0 {
					detail.CPUPercent = float64(totalSec) / detail.Uptime.Seconds() * 100.0
				}
			}
		}
	}

	// Find ports for this PID
	detail.Ports = findPortsForPID(pid)

	return detail, nil
}

// getBootTime reads /proc/stat for btime
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

// findPortsForPID finds all ports a PID is listening on
func findPortsForPID(pid int) []int {
	// Read socket inodes for this PID
	fdDir := filepath.Join("/proc", strconv.Itoa(pid), "fd")
	fds, err := os.ReadDir(fdDir)
	if err != nil {
		return nil
	}

	inodes := make(map[uint64]bool)
	for _, fd := range fds {
		link, err := os.Readlink(filepath.Join(fdDir, fd.Name()))
		if err != nil {
			continue
		}
		if strings.HasPrefix(link, "socket:[") && strings.HasSuffix(link, "]") {
			inodeStr := link[8 : len(link)-1]
			inode, err := strconv.ParseUint(inodeStr, 10, 64)
			if err == nil {
				inodes[inode] = true
			}
		}
	}

	if len(inodes) == 0 {
		return nil
	}

	// Find these inodes in /proc/net/tcp and /proc/net/udp
	var ports []int
	portSet := make(map[int]bool)
	for _, path := range []string{"/proc/net/tcp", "/proc/net/tcp6", "/proc/net/udp", "/proc/net/udp6"} {
		found := findPortsByInodes(path, inodes)
		for _, p := range found {
			if !portSet[p] {
				portSet[p] = true
				ports = append(ports, p)
			}
		}
	}
	return ports
}

// findPortsByInodes scans a /proc/net file for matching inodes
func findPortsByInodes(path string, inodes map[uint64]bool) []int {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var ports []int
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if i == 0 {
			continue // skip header
		}
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}
		inode, err := strconv.ParseUint(fields[9], 10, 64)
		if err != nil {
			continue
		}
		if !inodes[inode] {
			continue
		}
		// Parse port from local_address
		localAddr := fields[1]
		parts := strings.SplitN(localAddr, ":", 2)
		if len(parts) != 2 {
			continue
		}
		port, err := strconv.ParseInt(parts[1], 16, 32)
		if err != nil {
			continue
		}
		if port > 0 {
			ports = append(ports, int(port))
		}
	}
	return ports
}
