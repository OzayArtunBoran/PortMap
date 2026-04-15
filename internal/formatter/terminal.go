package formatter

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/ozayartunboran/portmap/internal/detector"
	"github.com/ozayartunboran/portmap/internal/process"
	"github.com/ozayartunboran/portmap/internal/scanner"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorBold   = "\033[1m"
)

// TerminalFormatter outputs colored table format to the terminal
type TerminalFormatter struct {
	NoColor bool
}

func (f *TerminalFormatter) color(code string) string {
	if f.NoColor || !isTerminal() {
		return ""
	}
	return code
}

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func (f *TerminalFormatter) FormatScan(result *scanner.ScanResult) string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)

	// Header
	fmt.Fprintf(w, "%sPORT\tPROTO\tPID\tPROCESS\tUSER\tSTATE%s\n",
		f.color(colorBold), f.color(colorReset))

	for _, p := range result.Ports {
		stateColor := f.stateColor(p.State)
		fmt.Fprintf(w, "%s%d\t%s\t%d\t%s\t%s\t%s%s\n",
			stateColor, p.Port, p.Protocol, p.PID, p.ProcessName, p.User, p.State, f.color(colorReset))
	}
	w.Flush()

	b.WriteString(fmt.Sprintf("\n%d ports found (scanned %s in %s)\n",
		result.Total, result.ScannedRange, result.Duration.Round(1_000_000)))
	return b.String()
}

func (f *TerminalFormatter) stateColor(state string) string {
	switch strings.ToUpper(state) {
	case "LISTEN", "UNCONN":
		return f.color(colorGreen)
	case "ESTABLISHED":
		return f.color(colorYellow)
	case "TIME_WAIT", "CLOSE_WAIT", "FIN_WAIT1", "FIN_WAIT2":
		return f.color(colorBlue)
	default:
		return ""
	}
}

func (f *TerminalFormatter) FormatCheck(result *detector.CheckResult) string {
	var b strings.Builder

	if len(result.Conflicts) == 0 {
		b.WriteString(fmt.Sprintf("%s✓ No conflicts found.%s (%d services checked)\n",
			f.color(colorGreen), f.color(colorReset), result.TotalChecked))
		if len(result.OKServices) > 0 {
			b.WriteString(fmt.Sprintf("  OK: %s\n", strings.Join(result.OKServices, ", ")))
		}
		return b.String()
	}

	b.WriteString(fmt.Sprintf("%s✗ %d conflict(s) found:%s\n\n",
		f.color(colorRed), len(result.Conflicts), f.color(colorReset)))

	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "%sTYPE\tPORT\tSERVICE\tACTUAL\tPID\tSUGGESTION%s\n",
		f.color(colorBold), f.color(colorReset))

	for _, c := range result.Conflicts {
		suggestion := "-"
		if c.Suggestion > 0 {
			suggestion = fmt.Sprintf("%d", c.Suggestion)
		}
		fmt.Fprintf(w, "%s%s\t%d\t%s\t%s\t%d\t%s%s\n",
			f.color(colorRed), c.Type, c.Port, c.ServiceName, c.ActualProcess, c.ActualPID,
			suggestion, f.color(colorReset))
	}
	w.Flush()

	if len(result.OKServices) > 0 {
		b.WriteString(fmt.Sprintf("\n%sOK:%s %s\n",
			f.color(colorGreen), f.color(colorReset), strings.Join(result.OKServices, ", ")))
	}

	return b.String()
}

func (f *TerminalFormatter) FormatInfo(detail *process.ProcessDetail) string {
	if detail == nil {
		return "No process found.\n"
	}

	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)

	fmt.Fprintf(w, "%sPID:%s\t%d\n", f.color(colorBold), f.color(colorReset), detail.PID)
	fmt.Fprintf(w, "%sName:%s\t%s\n", f.color(colorBold), f.color(colorReset), detail.Name)
	fmt.Fprintf(w, "%sCommand:%s\t%s\n", f.color(colorBold), f.color(colorReset), detail.CommandLine)
	fmt.Fprintf(w, "%sUser:%s\t%s\n", f.color(colorBold), f.color(colorReset), detail.User)

	if len(detail.Ports) > 0 {
		portStrs := make([]string, len(detail.Ports))
		for i, p := range detail.Ports {
			portStrs[i] = fmt.Sprintf("%d", p)
		}
		fmt.Fprintf(w, "%sPorts:%s\t%s\n", f.color(colorBold), f.color(colorReset), strings.Join(portStrs, ", "))
	}

	fmt.Fprintf(w, "%sCPU:%s\t%.1f%%\n", f.color(colorBold), f.color(colorReset), detail.CPUPercent)
	fmt.Fprintf(w, "%sMemory:%s\t%.1f MB\n", f.color(colorBold), f.color(colorReset), detail.MemoryMB)

	if !detail.StartTime.IsZero() {
		fmt.Fprintf(w, "%sStarted:%s\t%s\n", f.color(colorBold), f.color(colorReset), detail.StartTime.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(w, "%sUptime:%s\t%s\n", f.color(colorBold), f.color(colorReset), detail.Uptime.Round(1_000_000_000))
	}

	w.Flush()
	return b.String()
}
