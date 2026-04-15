package process

import (
	"fmt"
	"time"

	"github.com/ozayartunboran/portmap/internal/scanner"
)

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
	scn := scanner.New()
	result, err := scn.Scan(scanner.ScanOptions{
		Port: port,
	})
	if err != nil {
		return nil, fmt.Errorf("scan port %d: %w", port, err)
	}

	if len(result.Ports) == 0 {
		return nil, fmt.Errorf("nothing running on port %d", port)
	}

	pid := result.Ports[0].PID
	if pid == 0 {
		return &ProcessDetail{
			Name:  result.Ports[0].ProcessName,
			Ports: []int{port},
		}, nil
	}

	detail, err := GetByPID(pid)
	if err != nil {
		// Return partial info from scan
		return &ProcessDetail{
			PID:         pid,
			Name:        result.Ports[0].ProcessName,
			User:        result.Ports[0].User,
			CommandLine: result.Ports[0].CommandLine,
			Ports:       []int{port},
			StartTime:   result.Ports[0].StartTime,
		}, nil
	}

	// Ensure the queried port is in the list
	hasPort := false
	for _, p := range detail.Ports {
		if p == port {
			hasPort = true
			break
		}
	}
	if !hasPort {
		detail.Ports = append(detail.Ports, port)
	}

	return detail, nil
}
