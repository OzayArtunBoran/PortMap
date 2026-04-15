package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ozayartunboran/portmap/internal/config"
)

var removeCmd = &cobra.Command{
	Use:   "remove <service-name>",
	Short: "Remove a service from config",
	Long:  `Remove a service and its port assignment from the .portmap.yml config file.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRemove,
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if _, exists := cfg.Services[name]; !exists {
		return fmt.Errorf("service %q not found", name)
	}

	delete(cfg.Services, name)

	// Remove from any groups that reference it
	for gName, group := range cfg.Groups {
		filtered := make([]string, 0, len(group.Services))
		for _, s := range group.Services {
			if s != name {
				filtered = append(filtered, s)
			}
		}
		group.Services = filtered
		cfg.Groups[gName] = group
	}

	if err := config.SaveConfig(cfgFile, cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Printf("Removed service %q\n", name)
	return nil
}
