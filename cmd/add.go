package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ozayartunboran/portmap/internal/allocator"
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

	port := addFlags.port
	if port == 0 {
		// Auto-assign a port
		defaultRange, err := config.ParseRange(cfg.Defaults.Range)
		if err != nil {
			defaultRange = config.PortRange{Start: 3000, End: 9999}
		}

		// Exclude ports already in use by config
		var exclude []int
		for _, svc := range cfg.Services {
			if svc.Port > 0 {
				exclude = append(exclude, svc.Port)
			}
		}

		alloc := allocator.New()
		ports, err := alloc.FindFree(allocator.AllocRequest{
			Range:    defaultRange,
			Strategy: allocator.AllocStrategy(cfg.Defaults.Strategy),
			Count:    1,
			Exclude:  exclude,
		})
		if err != nil {
			return fmt.Errorf("auto-assign port: %w", err)
		}
		port = ports[0]
	}

	managed := true
	svc := config.ServiceConfig{
		Port:        port,
		Description: addFlags.description,
		Command:     addFlags.command,
		Managed:     &managed,
	}

	if err := cfg.AddService(name, svc); err != nil {
		return err
	}

	if err := config.SaveConfig(cfgFile, cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Printf("Added service %q on port %d\n", name, port)
	return nil
}
