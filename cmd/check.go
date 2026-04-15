package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var checkFlags struct {
	format string
	fix    bool
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for port conflicts against config",
	Long:  `Compare configured port assignments with actually running ports and report conflicts.`,
	RunE:  runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)

	checkCmd.Flags().StringVarP(&checkFlags.format, "format", "f", "terminal", "output format: terminal, json, markdown, compact")
	checkCmd.Flags().BoolVar(&checkFlags.fix, "fix", false, "auto-resolve conflicts by assigning free ports")
}

func runCheck(cmd *cobra.Command, args []string) error {
	// TODO: Phase 2 — implement with config + scanner + detector
	fmt.Printf("Checking port conflicts using config: %s\n", cfgFile)
	return nil
}
