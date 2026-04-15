package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/ozayartunboran/portmap/internal/scanner"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var watchFlags struct {
	interval  int
	portRange string
}

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Live port monitoring",
	Long:  `Continuously monitor ports and display a live-updating table in the terminal. Press q to quit.`,
	RunE:  runWatch,
}

func init() {
	rootCmd.AddCommand(watchCmd)

	watchCmd.Flags().IntVarP(&watchFlags.interval, "interval", "i", 2, "refresh interval in seconds")
	watchCmd.Flags().StringVarP(&watchFlags.portRange, "range", "r", "1-65535", "port range to watch")
}

// portKey uniquely identifies a port entry
type portKey struct {
	Port     int
	Protocol string
	PID      int
}

// highlightEntry tracks highlight state for new/closed ports
type highlightEntry struct {
	info      scanner.PortInfo
	remaining int  // cycles remaining for highlight
	closed    bool // true = closing animation
}

func runWatch(cmd *cobra.Command, args []string) error {
	opts, err := parseWatchRange(watchFlags.portRange)
	if err != nil {
		return fmt.Errorf("invalid range: %w", err)
	}
	opts.ListenOnly = true

	interval := time.Duration(watchFlags.interval) * time.Second
	if interval < time.Second {
		interval = time.Second
	}

	sc := scanner.New()

	// Set terminal to raw mode for key input
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer term.Restore(fd, oldState)

	// Hide cursor
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")

	// Signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	// Key input goroutine
	quitCh := make(chan struct{})
	go func() {
		buf := make([]byte, 1)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil || n == 0 {
				continue
			}
			if buf[0] == 'q' || buf[0] == 'Q' || buf[0] == 3 { // 3 = Ctrl+C
				close(quitCh)
				return
			}
		}
	}()

	highlights := make(map[portKey]*highlightEntry)
	var prevPorts map[portKey]scanner.PortInfo
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Render immediately on first tick
	render := func() {
		result, err := sc.Scan(opts)
		if err != nil {
			return
		}

		currPorts := make(map[portKey]scanner.PortInfo)
		for _, p := range result.Ports {
			k := portKey{p.Port, p.Protocol, p.PID}
			currPorts[k] = p
		}

		if prevPorts != nil {
			// New ports → green highlight
			for k, p := range currPorts {
				if _, existed := prevPorts[k]; !existed {
					highlights[k] = &highlightEntry{info: p, remaining: 3}
				}
			}
			// Closed ports → red highlight then remove
			for k, p := range prevPorts {
				if _, exists := currPorts[k]; !exists {
					highlights[k] = &highlightEntry{info: p, remaining: 1, closed: true}
				}
			}
		}

		output := renderWatchTable(result, highlights, watchFlags.portRange, noColor)

		// Clear screen and move cursor to top
		fmt.Print("\033[2J\033[H")
		fmt.Print(output)

		// Decrement highlight counters
		for k, h := range highlights {
			h.remaining--
			if h.remaining <= 0 {
				delete(highlights, k)
			}
		}

		prevPorts = currPorts
	}

	render()
	for {
		select {
		case <-ticker.C:
			render()
		case <-sigCh:
			return nil
		case <-quitCh:
			return nil
		}
	}
}

// renderWatchTable builds the TUI output string. Exported logic for testing.
func renderWatchTable(result *scanner.ScanResult, highlights map[portKey]*highlightEntry, portRange string, disableColor bool) string {
	var b strings.Builder

	colorFn := func(code string) string {
		if disableColor {
			return ""
		}
		return code
	}

	const (
		reset  = "\033[0m"
		bold   = "\033[1m"
		green  = "\033[32m"
		red    = "\033[31m"
		yellow = "\033[33m"
		blue   = "\033[34m"
		dim    = "\033[2m"
	)

	b.WriteString(fmt.Sprintf("%s portmap watch %s\n\n", colorFn(bold), colorFn(reset)))

	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "%sPORT\tPROTO\tPID\tPROCESS\tUSER\tSTATE%s\n",
		colorFn(bold), colorFn(reset))

	// Build sorted list of ports to display
	type displayPort struct {
		info  scanner.PortInfo
		color string
	}

	var ports []displayPort

	// Current ports
	for _, p := range result.Ports {
		k := portKey{p.Port, p.Protocol, p.PID}
		c := ""
		if h, ok := highlights[k]; ok && !h.closed {
			c = colorFn(green) // new port highlight
		} else {
			// State-based coloring
			switch strings.ToUpper(p.State) {
			case "LISTEN", "UNCONN":
				c = colorFn(green)
			case "ESTABLISHED":
				c = colorFn(yellow)
			case "TIME_WAIT", "CLOSE_WAIT":
				c = colorFn(blue)
			}
		}
		ports = append(ports, displayPort{info: p, color: c})
	}

	// Closed ports (still showing in red)
	for _, h := range highlights {
		if h.closed {
			ports = append(ports, displayPort{info: h.info, color: colorFn(red)})
		}
	}

	sort.Slice(ports, func(i, j int) bool {
		return ports[i].info.Port < ports[j].info.Port
	})

	for _, dp := range ports {
		p := dp.info
		fmt.Fprintf(w, "%s%d\t%s\t%d\t%s\t%s\t%s%s\n",
			dp.color, p.Port, p.Protocol, p.PID, p.ProcessName, p.User, p.State, colorFn(reset))
	}
	w.Flush()

	// Status bar
	now := time.Now().Format("15:04:05")
	b.WriteString(fmt.Sprintf("\n%sWatching %s | %d ports | q to quit | %s%s\n",
		colorFn(dim), portRange, result.Total, now, colorFn(reset)))

	return b.String()
}

func parseWatchRange(rangeStr string) (scanner.ScanOptions, error) {
	opts := scanner.ScanOptions{}
	parts := strings.SplitN(rangeStr, "-", 2)
	if len(parts) == 1 {
		port, err := strconv.Atoi(parts[0])
		if err != nil {
			return opts, err
		}
		opts.Port = port
		return opts, nil
	}
	start, err := strconv.Atoi(parts[0])
	if err != nil {
		return opts, err
	}
	end, err := strconv.Atoi(parts[1])
	if err != nil {
		return opts, err
	}
	if start > end || start < 0 || end > 65535 {
		return opts, fmt.Errorf("invalid range %d-%d", start, end)
	}
	opts.RangeStart = start
	opts.RangeEnd = end
	return opts, nil
}
