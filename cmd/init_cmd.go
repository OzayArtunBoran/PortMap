package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ozayartunboran/portmap/internal/config"
)

var initFlags struct {
	detect bool
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a .portmap.yml config file",
	Long:  `Interactively create a .portmap.yml configuration file for your project.`,
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolVarP(&initFlags.detect, "detect", "d", false, "auto-detect running services and populate config")
}

func runInit(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	// Check if config already exists
	if _, err := os.Stat(cfgFile); err == nil {
		fmt.Printf("Config file %s already exists. Overwrite? [y/N]: ", cfgFile)
		answer, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(answer)) != "y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	cfg := &config.PortmapConfig{
		Version: "1",
		Defaults: config.ConfigDefaults{
			Range:    "3000-9999",
			Strategy: "nearest",
		},
		Services: make(map[string]config.ServiceConfig),
		Groups:   make(map[string]config.GroupConfig),
	}

	if initFlags.detect {
		// TODO: Phase 2 — auto-detect running services using scanner
		fmt.Println("Auto-detecting running services...")
	}

	if err := config.SaveConfig(cfgFile, cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Printf("Created %s\n", cfgFile)
	fmt.Println("Add services with: portmap add <name> --port <port>")
	return nil
}
