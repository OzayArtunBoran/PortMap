package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/ozayartunboran/portmap/internal/scanner"
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
	port, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid port number: %s", args[0])
	}

	scn := scanner.New()
	result, err := scn.Scan(scanner.ScanOptions{
		Port: port,
	})
	if err != nil {
		return fmt.Errorf("scan port: %w", err)
	}

	if len(result.Ports) == 0 {
		fmt.Printf("Nothing running on port %d\n", port)
		return nil
	}

	target := result.Ports[0]
	if target.PID == 0 {
		return fmt.Errorf("cannot determine PID for port %d (try running with sudo)", port)
	}

	// Don't kill ourselves
	if target.PID == os.Getpid() {
		return fmt.Errorf("refusing to kill own process (PID %d)", target.PID)
	}

	fmt.Printf("Process: %s (PID %d) on port %d\n", target.ProcessName, target.PID, port)

	if !killFlags.yes {
		fmt.Printf("Kill this process? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(answer)) != "y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	proc, err := os.FindProcess(target.PID)
	if err != nil {
		return fmt.Errorf("find process: %w", err)
	}

	if killFlags.force {
		if err := proc.Signal(syscall.SIGKILL); err != nil {
			return fmt.Errorf("SIGKILL failed: %w", err)
		}
		fmt.Printf("Sent SIGKILL to PID %d\n", target.PID)
		return nil
	}

	// Send SIGTERM first
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("SIGTERM failed: %w", err)
	}
	fmt.Printf("Sent SIGTERM to PID %d, waiting...\n", target.PID)

	// Wait up to 5 seconds
	deadline := time.After(5 * time.Second)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			fmt.Printf("Process still running after 5s. Use --force to SIGKILL.\n")
			return nil
		case <-ticker.C:
			if err := proc.Signal(syscall.Signal(0)); err != nil {
				fmt.Printf("Process %d terminated.\n", target.PID)
				return nil
			}
		}
	}
}
