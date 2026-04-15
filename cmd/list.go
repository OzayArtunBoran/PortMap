package cmd

import (
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ozayartunboran/portmap/internal/allocator"
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

	// Sort service names for consistent output
	names := make([]string, 0, len(cfg.Services))
	for name := range cfg.Services {
		names = append(names, name)
	}
	sort.Strings(names)

	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "SERVICE\tPORT\tSTATUS\tDESCRIPTION\n")

	for _, name := range names {
		svc := cfg.Services[name]
		status := "free"
		if svc.Port > 0 && !allocator.IsPortFree(svc.Port) {
			status = "in-use"
		}
		managed := ""
		if !svc.IsManaged() {
			managed = " (unmanaged)"
		}
		fmt.Fprintf(w, "%s\t%d\t%s\t%s%s\n", name, svc.Port, status, svc.Description, managed)
	}
	w.Flush()

	fmt.Print(b.String())
	return nil
}
