package formatter

import (
	"fmt"
	"strings"

	"github.com/ozayartunboran/portmap/internal/detector"
	"github.com/ozayartunboran/portmap/internal/process"
	"github.com/ozayartunboran/portmap/internal/scanner"
)

// TerminalFormatter outputs colored table format to the terminal
type TerminalFormatter struct {
	NoColor bool
}

func (f *TerminalFormatter) FormatScan(result *scanner.ScanResult) string {
	// TODO: Phase 2 — implement colored tabwriter output
	var b strings.Builder
	b.WriteString(fmt.Sprintf("PORT\tPROTO\tPID\tPROCESS\tUSER\tSTATE\n"))
	for _, p := range result.Ports {
		b.WriteString(fmt.Sprintf("%d\t%s\t%d\t%s\t%s\t%s\n",
			p.Port, p.Protocol, p.PID, p.ProcessName, p.User, p.State))
	}
	b.WriteString(fmt.Sprintf("\n%d ports found (scanned %s in %s)\n",
		result.Total, result.ScannedRange, result.Duration))
	return b.String()
}

func (f *TerminalFormatter) FormatCheck(result *detector.CheckResult) string {
	// TODO: Phase 2 — implement colored conflict output
	var b strings.Builder
	if len(result.Conflicts) == 0 {
		b.WriteString("No conflicts found.\n")
	} else {
		b.WriteString(fmt.Sprintf("%d conflict(s) found:\n", len(result.Conflicts)))
		for _, c := range result.Conflicts {
			b.WriteString(fmt.Sprintf("  [%s] port %d: %s\n", c.Type, c.Port, c.Message))
		}
	}
	return b.String()
}

func (f *TerminalFormatter) FormatInfo(detail *process.ProcessDetail) string {
	// TODO: Phase 2 — implement
	if detail == nil {
		return "No process found.\n"
	}
	return fmt.Sprintf("PID: %d\nName: %s\nCommand: %s\nPorts: %v\n",
		detail.PID, detail.Name, detail.CommandLine, detail.Ports)
}
