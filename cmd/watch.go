package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var watchFlags struct {
	interval  int
	portRange string
}

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Live port monitoring",
	Long:  `Continuously monitor ports and display a live-updating table in the terminal. Press q to quit.`,
	RunE:  runWatch,
}

func init() {
	rootCmd.AddCommand(watchCmd)

	watchCmd.Flags().IntVarP(&watchFlags.interval, "interval", "i", 2, "refresh interval in seconds")
	watchCmd.Flags().StringVarP(&watchFlags.portRange, "range", "r", "1-65535", "port range to watch")
}

func runWatch(cmd *cobra.Command, args []string) error {
	// TODO: Phase 4 — implement TUI with ANSI escape codes
	fmt.Printf("Watch mode: interval=%ds range=%s (not yet implemented)\n", watchFlags.interval, watchFlags.portRange)
	return nil
}
