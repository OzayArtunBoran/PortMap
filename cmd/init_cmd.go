package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ozayartunboran/portmap/internal/config"
	"github.com/ozayartunboran/portmap/internal/scanner"
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
		fmt.Println("Auto-detecting running services...")
		scn := scanner.New()
		result, err := scn.Scan(scanner.ScanOptions{
			ListenOnly: true,
			RangeStart: 1024,
			RangeEnd:   65535,
		})
		if err == nil && len(result.Ports) > 0 {
			for _, p := range result.Ports {
				if p.ProcessName == "" || p.ProcessName == "unknown (requires root)" {
					continue
				}
				name := strings.ToLower(p.ProcessName)
				// Avoid duplicates
				if _, exists := cfg.Services[name]; exists {
					continue
				}
				cfg.Services[name] = config.ServiceConfig{
					Port:        p.Port,
					Description: fmt.Sprintf("Auto-detected %s", p.ProcessName),
				}
			}
			fmt.Printf("Detected %d service(s)\n", len(cfg.Services))
		}
	}

	if err := config.SaveConfig(cfgFile, cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Printf("Created %s\n", cfgFile)
	fmt.Println("Add services with: portmap add <name> --port <port>")
	return nil
}
