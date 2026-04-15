package process

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// GetByPID retrieves detailed process information on macOS
func GetByPID(pid int) (*ProcessDetail, error) {
	pidStr := strconv.Itoa(pid)

	// ps -p {pid} -o pid=,pcpu=,pmem=,lstart=,command=
	cmd := exec.Command("ps", "-p", pidStr, "-o", "pid=,pcpu=,pmem=,lstart=,command=")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("process %d not found", pid)
	}

	line := strings.TrimSpace(string(output))
	if line == "" {
		return nil, fmt.Errorf("process %d not found", pid)
	}

	detail := &ProcessDetail{PID: pid}

	// Parse the output — fields are space-separated but lstart has spaces
	// Format: PID %CPU %MEM LSTART COMMAND
	// LSTART is like "Wed Apr 15 10:30:00 2026" (5 tokens)
	fields := strings.Fields(line)
	if len(fields) < 8 {
		return detail, nil
	}

	detail.CPUPercent, _ = strconv.ParseFloat(fields[1], 64)
	detail.MemoryMB, _ = strconv.ParseFloat(fields[2], 64) // This is % on macOS

	// lstart is fields[3:8] (day month date time year)
	if len(fields) >= 8 {
		lstartStr := strings.Join(fields[3:8], " ")
		if t, err := time.Parse("Mon Jan 2 15:04:05 2006", lstartStr); err == nil {
			detail.StartTime = t
			detail.Uptime = time.Since(t)
		}
	}

	// Command is fields[8:]
	if len(fields) > 8 {
		detail.CommandLine = strings.Join(fields[8:], " ")
		// Name is the basename of the command
		cmdParts := strings.Fields(detail.CommandLine)
		if len(cmdParts) > 0 {
			parts := strings.Split(cmdParts[0], "/")
			detail.Name = parts[len(parts)-1]
		}
	}

	// Get user
	userCmd := exec.Command("ps", "-p", pidStr, "-o", "user=")
	if userOut, err := userCmd.Output(); err == nil {
		detail.User = strings.TrimSpace(string(userOut))
	}

	return detail, nil
}
