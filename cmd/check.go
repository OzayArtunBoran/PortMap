package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ozayartunboran/portmap/internal/allocator"
	"github.com/ozayartunboran/portmap/internal/config"
	"github.com/ozayartunboran/portmap/internal/detector"
	"github.com/ozayartunboran/portmap/internal/formatter"
	"github.com/ozayartunboran/portmap/internal/scanner"
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
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	scn := scanner.New()
	det := detector.New(cfg, scn)
	result, err := det.Check()
	if err != nil {
		return fmt.Errorf("check failed: %w", err)
	}

	if checkFlags.fix && result.HasConflicts() {
		defaultRange, _ := config.ParseRange(cfg.Defaults.Range)
		alloc := allocator.New()

		for _, conflict := range result.Conflicts {
			if conflict.Type != detector.ConflictOccupied || conflict.Suggestion <= 0 {
				continue
			}
			// Verify suggestion is still free
			if !allocator.IsPortFree(conflict.Suggestion) {
				suggested, err := alloc.FindFree(allocator.AllocRequest{
					PreferredPort: conflict.Port,
					Range:         defaultRange,
					Strategy:      allocator.StrategyNearest,
					Count:         1,
				})
				if err != nil || len(suggested) == 0 {
					continue
				}
				conflict.Suggestion = suggested[0]
			}
			if err := cfg.UpdatePort(conflict.ServiceName, conflict.Suggestion); err == nil {
				fmt.Printf("Fixed: %s port %d → %d\n", conflict.ServiceName, conflict.Port, conflict.Suggestion)
			}
		}
		if err := config.SaveConfig(cfgFile, cfg); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
	}

	fmtr, err := formatter.New(checkFlags.format, noColor)
	if err != nil {
		return err
	}

	fmt.Print(fmtr.FormatCheck(result))

	if result.HasConflicts() {
		// Return non-zero exit code for CI usage
		return fmt.Errorf("%d conflict(s) found", len(result.Conflicts))
	}
	return nil
}
