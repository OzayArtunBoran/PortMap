package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
	noColor bool

	ver     string
	buildAt string
)

// SetVersionInfo sets version info from build-time ldflags
func SetVersionInfo(v, b string) {
	ver = v
	buildAt = b
}

var rootCmd = &cobra.Command{
	Use:   "portmap",
	Short: "Never fight over ports again",
	Long: `PortMap — Local development port collision resolver.

Shows what's running on each port, detects conflicts against your config,
suggests free ports, and manages port assignments for all your services.`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", ".portmap.yml", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")

	if ver == "" {
		ver = "dev"
	}
	rootCmd.Version = ver
	rootCmd.SetVersionTemplate(fmt.Sprintf("portmap %s (built %s)\n", ver, buildAt))
}
