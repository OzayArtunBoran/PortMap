package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ozayartunboran/portmap/internal/config"
	"github.com/ozayartunboran/portmap/internal/formatter"
	"github.com/ozayartunboran/portmap/internal/scanner"
)

var scanFlags struct {
	port       int
	portRange  string
	format     string
	filter     string
	tcpOnly    bool
	udpOnly    bool
	listenOnly bool
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan active ports on this machine",
	Long:  `Scan all active TCP/UDP ports and show which process is using each one.`,
	RunE:  runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().IntVarP(&scanFlags.port, "port", "p", 0, "scan specific port (0 = all)")
	scanCmd.Flags().StringVarP(&scanFlags.portRange, "range", "r", "1-65535", "port range to scan (e.g. 3000-9999)")
	scanCmd.Flags().StringVarP(&scanFlags.format, "format", "f", "terminal", "output format: terminal, json, markdown, compact")
	scanCmd.Flags().StringVar(&scanFlags.filter, "filter", "", "filter by process name (substring match)")
	scanCmd.Flags().BoolVar(&scanFlags.tcpOnly, "tcp-only", false, "show only TCP ports")
	scanCmd.Flags().BoolVar(&scanFlags.udpOnly, "udp-only", false, "show only UDP ports")
	scanCmd.Flags().BoolVar(&scanFlags.listenOnly, "listen-only", false, "show only LISTEN state ports")
}

func runScan(cmd *cobra.Command, args []string) error {
	portRange, err := config.ParseRange(scanFlags.portRange)
	if err != nil {
		return fmt.Errorf("invalid range: %w", err)
	}

	opts := scanner.ScanOptions{
		Port:       scanFlags.port,
		RangeStart: portRange.Start,
		RangeEnd:   portRange.End,
		TCPOnly:    scanFlags.tcpOnly,
		UDPOnly:    scanFlags.udpOnly,
		ListenOnly: scanFlags.listenOnly,
		Filter:     scanFlags.filter,
	}

	scn := scanner.New()
	result, err := scn.Scan(opts)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	fmtr, err := formatter.New(scanFlags.format, noColor)
	if err != nil {
		return err
	}

	fmt.Print(fmtr.FormatScan(result))
	return nil
}
