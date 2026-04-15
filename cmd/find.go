package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ozayartunboran/portmap/internal/allocator"
	"github.com/ozayartunboran/portmap/internal/config"
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
	portRange, err := config.ParseRange(findFlags.portRange)
	if err != nil {
		return fmt.Errorf("invalid range: %w", err)
	}

	alloc := allocator.New()
	ports, err := alloc.FindFree(allocator.AllocRequest{
		PreferredPort: findFlags.near,
		Range:         portRange,
		Strategy:      allocator.StrategyNearest,
		Count:         findFlags.count,
	})
	if err != nil {
		return fmt.Errorf("find free ports: %w", err)
	}

	for _, port := range ports {
		fmt.Println(port)
	}
	return nil
}
