package formatter

import (
	"encoding/json"

	"github.com/ozayartunboran/portmap/internal/detector"
	"github.com/ozayartunboran/portmap/internal/process"
	"github.com/ozayartunboran/portmap/internal/scanner"
)

// JSONFormatter outputs machine-readable JSON
type JSONFormatter struct{}

func (f *JSONFormatter) FormatScan(result *scanner.ScanResult) string {
	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data) + "\n"
}

func (f *JSONFormatter) FormatCheck(result *detector.CheckResult) string {
	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data) + "\n"
}

func (f *JSONFormatter) FormatInfo(detail *process.ProcessDetail) string {
	if detail == nil {
		return "{}\n"
	}
	data, _ := json.MarshalIndent(detail, "", "  ")
	return string(data) + "\n"
}
