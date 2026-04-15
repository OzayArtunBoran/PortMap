package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var killFlags struct {
	force bool
	yes   bool
}

var killCmd = &cobra.Command{
	Use:   "kill <port>",
	Short: "Kill process occupying a specific port",
	Long:  `Find the process using a specific port and terminate it. Sends SIGTERM by default, use --force for SIGKILL.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runKill,
}

func init() {
	rootCmd.AddCommand(killCmd)

	killCmd.Flags().BoolVar(&killFlags.force, "force", false, "force kill (SIGKILL instead of SIGTERM)")
	killCmd.Flags().BoolVarP(&killFlags.yes, "yes", "y", false, "skip confirmation prompt")
}

func runKill(cmd *cobra.Command, args []string) error {
	// TODO: Phase 2 — implement with scanner + process kill
	fmt.Printf("Kill process on port %s (force=%v, yes=%v) — not yet implemented\n", args[0], killFlags.force, killFlags.yes)
	return nil
}
