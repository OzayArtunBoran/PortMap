package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var findFlags struct {
	count     int
	near      int
	portRange string
}

var findCmd = &cobra.Command{
	Use:   "find",
	Short: "Find free ports",
	Long:  `Find available ports in a given range, optionally near a preferred port number.`,
	RunE:  runFind,
}

func init() {
	rootCmd.AddCommand(findCmd)

	findCmd.Flags().IntVarP(&findFlags.count, "count", "n", 1, "number of free ports to find")
	findCmd.Flags().IntVar(&findFlags.near, "near", 3000, "find ports near this number")
	findCmd.Flags().StringVarP(&findFlags.portRange, "range", "r", "3000-9999", "search range")
}

func runFind(cmd *cobra.Command, args []string) error {
	// TODO: Phase 2 — implement with allocator
	fmt.Printf("Finding %d free port(s) near %d in range %s\n", findFlags.count, findFlags.near, findFlags.portRange)
	return nil
}
