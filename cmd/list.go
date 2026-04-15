package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ozayartunboran/portmap/internal/config"
)

var listFlags struct {
	format string
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all services from config with their port status",
	Long:  `Display all configured services, their port assignments, and whether each port is currently in use.`,
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringVarP(&listFlags.format, "format", "f", "terminal", "output format: terminal, json, markdown, compact")
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if len(cfg.Services) == 0 {
		fmt.Println("No services configured. Add one with: portmap add <name> --port <port>")
		return nil
	}

	// TODO: Phase 2 — check port status with scanner, format with formatter
	for name, svc := range cfg.Services {
		status := "unknown"
		fmt.Printf("  %-20s :%d  %s  %s\n", name, svc.Port, status, svc.Description)
	}

	return nil
}
