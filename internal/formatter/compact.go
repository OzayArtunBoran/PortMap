package formatter

import (
	"fmt"
	"strings"

	"github.com/ozayartunboran/portmap/internal/detector"
	"github.com/ozayartunboran/portmap/internal/process"
	"github.com/ozayartunboran/portmap/internal/scanner"
)

// CompactFormatter outputs one-line summaries for scripting
type CompactFormatter struct{}

func (f *CompactFormatter) FormatScan(result *scanner.ScanResult) string {
	// Format: 3000:node(1234) 8080:go(5678) 5432:postgres(910)
	parts := make([]string, 0, len(result.Ports))
	for _, p := range result.Ports {
		parts = append(parts, fmt.Sprintf("%d:%s(%d)", p.Port, p.ProcessName, p.PID))
	}
	return strings.Join(parts, " ") + "\n"
}

func (f *CompactFormatter) FormatCheck(result *detector.CheckResult) string {
	if len(result.Conflicts) == 0 {
		return "ok\n"
	}
	parts := make([]string, 0, len(result.Conflicts))
	for _, c := range result.Conflicts {
		parts = append(parts, fmt.Sprintf("%d:%s:%s", c.Port, c.Type, c.ServiceName))
	}
	return strings.Join(parts, " ") + "\n"
}

func (f *CompactFormatter) FormatInfo(detail *process.ProcessDetail) string {
	if detail == nil {
		return "none\n"
	}
	return fmt.Sprintf("%d:%s:%v\n", detail.PID, detail.Name, detail.Ports)
}
