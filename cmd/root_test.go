package cmd

import (
	"testing"
)

func TestRootCommand_Help(t *testing.T) {
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("root --help failed: %v", err)
	}
}

func TestRootCommand_Version(t *testing.T) {
	SetVersionInfo("1.0.0-test", "2026-01-01")
	if ver != "1.0.0-test" {
		t.Errorf("ver = %q, want %q", ver, "1.0.0-test")
	}
	if buildAt != "2026-01-01" {
		t.Errorf("buildAt = %q, want %q", buildAt, "2026-01-01")
	}
}

func TestRootCommand_Flags(t *testing.T) {
	// Test that persistent flags are defined
	f := rootCmd.PersistentFlags()

	if f.Lookup("config") == nil {
		t.Error("missing --config flag")
	}
	if f.Lookup("verbose") == nil {
		t.Error("missing --verbose flag")
	}
	if f.Lookup("no-color") == nil {
		t.Error("missing --no-color flag")
	}
}

func TestScanCommand_Flags(t *testing.T) {
	flags := scanCmd.Flags()

	expectedFlags := []string{"port", "range", "format", "filter", "tcp-only", "udp-only", "listen-only"}
	for _, name := range expectedFlags {
		if flags.Lookup(name) == nil {
			t.Errorf("missing --%s flag on scan command", name)
		}
	}
}

func TestScanCommand_Registered(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "scan" {
			found = true
			break
		}
	}
	if !found {
		t.Error("scan command not registered with root")
	}
}
