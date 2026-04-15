package process

import "time"

// ProcessDetail holds detailed information about a running process
type ProcessDetail struct {
	PID         int           `json:"pid"`
	Name        string        `json:"name"`
	CommandLine string        `json:"command_line"`
	User        string        `json:"user"`
	Ports       []int         `json:"ports"`        // All ports this process listens on
	CPUPercent  float64       `json:"cpu_percent"`
	MemoryMB    float64       `json:"memory_mb"`
	StartTime   time.Time     `json:"start_time"`
	Uptime      time.Duration `json:"uptime"`
}

// GetByPort finds process details for a given port
func GetByPort(port int) (*ProcessDetail, error) {
	// TODO: Phase 2 — implement:
	// 1. Use scanner to find PID for port
	// 2. Call GetByPID with found PID
	return nil, nil
}

// GetByPID retrieves detailed process information
func GetByPID(pid int) (*ProcessDetail, error) {
	// TODO: Phase 2 — implement:
	// Linux: parse /proc/{pid}/stat, /proc/{pid}/status, /proc/{pid}/cmdline
	// macOS: `ps -p {pid} -o pid,pcpu,pmem,lstart,command`
	return nil, nil
}
