package formatter

import (
	"fmt"
	"strings"

	"github.com/ozayartunboran/portmap/internal/detector"
	"github.com/ozayartunboran/portmap/internal/process"
	"github.com/ozayartunboran/portmap/internal/scanner"
)

// MarkdownFormatter outputs markdown table format
type MarkdownFormatter struct{}

func (f *MarkdownFormatter) FormatScan(result *scanner.ScanResult) string {
	var b strings.Builder
	b.WriteString("| Port | Proto | PID | Process | User | State |\n")
	b.WriteString("|------|-------|-----|---------|------|-------|\n")
	for _, p := range result.Ports {
		b.WriteString(fmt.Sprintf("| %d | %s | %d | %s | %s | %s |\n",
			p.Port, p.Protocol, p.PID, p.ProcessName, p.User, p.State))
	}
	b.WriteString(fmt.Sprintf("\n*%d ports found (scanned %s in %s)*\n",
		result.Total, result.ScannedRange, result.Duration))
	return b.String()
}

func (f *MarkdownFormatter) FormatCheck(result *detector.CheckResult) string {
	var b strings.Builder
	if len(result.Conflicts) == 0 {
		b.WriteString("**No conflicts found.**\n")
		return b.String()
	}
	b.WriteString(fmt.Sprintf("**%d conflict(s) found:**\n\n", len(result.Conflicts)))
	b.WriteString("| Port | Type | Service | Actual Process | Suggestion |\n")
	b.WriteString("|------|------|---------|----------------|------------|\n")
	for _, c := range result.Conflicts {
		b.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %d |\n",
			c.Port, c.Type, c.ServiceName, c.ActualProcess, c.Suggestion))
	}
	return b.String()
}

func (f *MarkdownFormatter) FormatInfo(detail *process.ProcessDetail) string {
	if detail == nil {
		return "*No process found.*\n"
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("**PID:** %d\n\n", detail.PID))
	b.WriteString(fmt.Sprintf("**Name:** %s\n\n", detail.Name))
	b.WriteString(fmt.Sprintf("**Command:** `%s`\n\n", detail.CommandLine))
	b.WriteString(fmt.Sprintf("**Ports:** %v\n", detail.Ports))
	return b.String()
}
