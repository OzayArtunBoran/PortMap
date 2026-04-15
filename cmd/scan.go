package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
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
	// TODO: Phase 2 — implement with scanner + formatter
	fmt.Println("Scanning ports...")
	fmt.Printf("Range: %s, Format: %s\n", scanFlags.portRange, scanFlags.format)
	return nil
}
