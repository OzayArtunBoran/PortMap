package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var infoFlags struct {
	port   int
	pid    int
	format string
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show detailed info about a port or process",
	Long:  `Display detailed information about what's running on a specific port or about a specific process ID.`,
	RunE:  runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)

	infoCmd.Flags().IntVarP(&infoFlags.port, "port", "p", 0, "port number")
	infoCmd.Flags().IntVar(&infoFlags.pid, "pid", 0, "process ID")
	infoCmd.Flags().StringVarP(&infoFlags.format, "format", "f", "terminal", "output format: terminal, json, markdown, compact")
}

func runInfo(cmd *cobra.Command, args []string) error {
	if infoFlags.port == 0 && infoFlags.pid == 0 {
		return fmt.Errorf("specify --port or --pid")
	}
	// TODO: Phase 2 — implement with scanner + process
	fmt.Printf("Getting info for port=%d pid=%d\n", infoFlags.port, infoFlags.pid)
	return nil
}
