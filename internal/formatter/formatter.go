package formatter

import (
	"fmt"

	"github.com/ozayartunboran/portmap/internal/detector"
	"github.com/ozayartunboran/portmap/internal/process"
	"github.com/ozayartunboran/portmap/internal/scanner"
)

// Formatter defines the output formatting interface
type Formatter interface {
	FormatScan(result *scanner.ScanResult) string
	FormatCheck(result *detector.CheckResult) string
	FormatInfo(detail *process.ProcessDetail) string
}

// New returns a formatter for the given format name
func New(format string, noColor bool) (Formatter, error) {
	switch format {
	case "terminal":
		return &TerminalFormatter{NoColor: noColor}, nil
	case "json":
		return &JSONFormatter{}, nil
	case "markdown":
		return &MarkdownFormatter{}, nil
	case "compact":
		return &CompactFormatter{}, nil
	default:
		return nil, fmt.Errorf("unknown format %q (available: terminal, json, markdown, compact)", format)
	}
}
