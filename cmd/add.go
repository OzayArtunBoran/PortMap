package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ozayartunboran/portmap/internal/config"
)

var addFlags struct {
	port        int
	description string
	command     string
}

var addCmd = &cobra.Command{
	Use:   "add <service-name>",
	Short: "Add a service to config",
	Long:  `Add a new service with a port assignment to the .portmap.yml config file.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().IntVarP(&addFlags.port, "port", "p", 0, "port number (0 = auto-assign)")
	addCmd.Flags().StringVarP(&addFlags.description, "description", "d", "", "service description")
	addCmd.Flags().StringVar(&addFlags.command, "command", "", "start command for this service")
}

func runAdd(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if _, exists := cfg.Services[name]; exists {
		return fmt.Errorf("service %q already exists", name)
	}

	managed := true
	svc := config.ServiceConfig{
		Port:        addFlags.port,
		Description: addFlags.description,
		Command:     addFlags.command,
		Managed:     &managed,
	}

	// TODO: Phase 2 — if port == 0, use allocator to auto-assign

	cfg.Services[name] = svc

	if err := config.SaveConfig(cfgFile, cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Printf("Added service %q on port %d\n", name, svc.Port)
	return nil
}
