package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".portmap.yml")

	content := `
version: "1"
defaults:
  range: "3000-9999"
  strategy: nearest
services:
  frontend:
    port: 3000
    description: "React app"
    command: "npm start"
  backend:
    port: 8080
    description: "Go API"
groups:
  web:
    services: [frontend, backend]
`
	os.WriteFile(cfgPath, []byte(content), 0644)

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Version != "1" {
		t.Errorf("version = %q, want %q", cfg.Version, "1")
	}
	if cfg.Defaults.Range != "3000-9999" {
		t.Errorf("range = %q, want %q", cfg.Defaults.Range, "3000-9999")
	}
	if cfg.Defaults.Strategy != "nearest" {
		t.Errorf("strategy = %q, want %q", cfg.Defaults.Strategy, "nearest")
	}
	if len(cfg.Services) != 2 {
		t.Errorf("services count = %d, want 2", len(cfg.Services))
	}
	if cfg.Services["frontend"].Port != 3000 {
		t.Errorf("frontend port = %d, want 3000", cfg.Services["frontend"].Port)
	}
	if cfg.Services["backend"].Port != 8080 {
		t.Errorf("backend port = %d, want 8080", cfg.Services["backend"].Port)
	}
	if len(cfg.Groups) != 1 {
		t.Errorf("groups count = %d, want 1", len(cfg.Groups))
	}
	if len(cfg.Groups["web"].Services) != 2 {
		t.Errorf("web group services count = %d, want 2", len(cfg.Groups["web"].Services))
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/.portmap.yml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "bad.yml")
	os.WriteFile(cfgPath, []byte("{{invalid yaml: [[["), 0644)

	_, err := LoadConfig(cfgPath)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadConfig_EnvVarExpansion(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".portmap.yml")

	t.Setenv("PORTMAP_TEST_PORT", "4567")

	content := `
version: "1"
defaults:
  range: "3000-9999"
  strategy: nearest
services:
  myservice:
    port: ${PORTMAP_TEST_PORT}
    description: "env var test"
`
	os.WriteFile(cfgPath, []byte(content), 0644)

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Services["myservice"].Port != 4567 {
		t.Errorf("port = %d, want 4567", cfg.Services["myservice"].Port)
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".portmap.yml")

	content := `
services:
  app:
    port: 3000
`
	os.WriteFile(cfgPath, []byte(content), 0644)

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Version != "1" {
		t.Errorf("default version = %q, want %q", cfg.Version, "1")
	}
	if cfg.Defaults.Range != "3000-9999" {
		t.Errorf("default range = %q, want %q", cfg.Defaults.Range, "3000-9999")
	}
	if cfg.Defaults.Strategy != "nearest" {
		t.Errorf("default strategy = %q, want %q", cfg.Defaults.Strategy, "nearest")
	}
}

func TestSaveConfig_WritesValidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".portmap.yml")

	managed := true
	cfg := &PortmapConfig{
		Version: "1",
		Defaults: ConfigDefaults{
			Range:    "3000-9999",
			Strategy: "nearest",
		},
		Services: map[string]ServiceConfig{
			"api": {Port: 8080, Description: "API server", Managed: &managed},
		},
		Groups: map[string]GroupConfig{},
	}

	err := SaveConfig(cfgPath, cfg)
	if err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}
	if loaded.Services["api"].Port != 8080 {
		t.Errorf("reloaded port = %d, want 8080", loaded.Services["api"].Port)
	}
	if loaded.Services["api"].Description != "API server" {
		t.Errorf("reloaded description = %q, want %q", loaded.Services["api"].Description, "API server")
	}
}

func TestValidate_DuplicatePort(t *testing.T) {
	cfg := &PortmapConfig{
		Defaults: ConfigDefaults{Range: "3000-9999", Strategy: "nearest"},
		Services: map[string]ServiceConfig{
			"svc1": {Port: 3000},
			"svc2": {Port: 3000},
		},
		Groups: map[string]GroupConfig{},
	}
	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for duplicate port")
	}
}

func TestValidate_InvalidPortRange(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"negative port", -1},
		{"port above 65535", 70000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &PortmapConfig{
				Defaults: ConfigDefaults{Range: "3000-9999", Strategy: "nearest"},
				Services: map[string]ServiceConfig{
					"svc": {Port: tt.port},
				},
				Groups: map[string]GroupConfig{},
			}
			err := cfg.Validate()
			if err == nil {
				t.Errorf("expected error for port %d", tt.port)
			}
		})
	}
}

func TestValidate_InvalidStrategy(t *testing.T) {
	cfg := &PortmapConfig{
		Defaults: ConfigDefaults{Range: "3000-9999", Strategy: "invalid_strategy"},
		Services: map[string]ServiceConfig{},
		Groups:   map[string]GroupConfig{},
	}
	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for invalid strategy")
	}
}

func TestValidate_InvalidRange(t *testing.T) {
	tests := []struct {
		name  string
		range_ string
	}{
		{"alphabetic", "abc-def"},
		{"reversed", "5000-3000"},
		{"single number", "5000"},
		{"empty", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &PortmapConfig{
				Defaults: ConfigDefaults{Range: tt.range_, Strategy: "nearest"},
				Services: map[string]ServiceConfig{},
				Groups:   map[string]GroupConfig{},
			}
			err := cfg.Validate()
			if err == nil {
				t.Errorf("expected error for range %q", tt.range_)
			}
		})
	}
}

func TestValidate_GroupReferencesUnknownService(t *testing.T) {
	cfg := &PortmapConfig{
		Defaults: ConfigDefaults{Range: "3000-9999", Strategy: "nearest"},
		Services: map[string]ServiceConfig{
			"api": {Port: 3000},
		},
		Groups: map[string]GroupConfig{
			"web": {Services: []string{"api", "nonexistent"}},
		},
	}
	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for unknown service in group")
	}
}

func TestParseRange_Valid(t *testing.T) {
	tests := []struct {
		name  string
		input string
		start int
		end   int
	}{
		{"standard range", "3000-9999", 3000, 9999},
		{"small range", "8080-8090", 8080, 8090},
		{"full range", "1-65535", 1, 65535},
		{"with spaces", " 3000 - 9999 ", 3000, 9999},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr, err := ParseRange(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if pr.Start != tt.start {
				t.Errorf("start = %d, want %d", pr.Start, tt.start)
			}
			if pr.End != tt.end {
				t.Errorf("end = %d, want %d", pr.End, tt.end)
			}
		})
	}
}

func TestParseRange_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"alphabetic", "abc-def"},
		{"single number", "5000"},
		{"empty", ""},
		{"reversed", "9000-3000"},
		{"zero start", "0-100"},
		{"exceeds max", "1-70000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseRange(tt.input)
			if err == nil {
				t.Errorf("expected error for input %q", tt.input)
			}
		})
	}
}

func TestPortRange_Contains(t *testing.T) {
	pr := PortRange{Start: 3000, End: 9999}

	tests := []struct {
		name string
		port int
		want bool
	}{
		{"below range", 2999, false},
		{"start boundary", 3000, true},
		{"mid range", 5000, true},
		{"end boundary", 9999, true},
		{"above range", 10000, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pr.Contains(tt.port)
			if got != tt.want {
				t.Errorf("Contains(%d) = %v, want %v", tt.port, got, tt.want)
			}
		})
	}
}

func TestServiceConfig_IsManaged(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name    string
		managed *bool
		want    bool
	}{
		{"nil (default true)", nil, true},
		{"explicit true", &trueVal, true},
		{"explicit false", &falseVal, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := ServiceConfig{Managed: tt.managed}
			got := svc.IsManaged()
			if got != tt.want {
				t.Errorf("IsManaged() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddService(t *testing.T) {
	cfg := &PortmapConfig{
		Services: map[string]ServiceConfig{
			"existing": {Port: 3000},
		},
	}

	// Add new service
	err := cfg.AddService("newservice", ServiceConfig{Port: 4000, Description: "new"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Services["newservice"].Port != 4000 {
		t.Errorf("new service port = %d, want 4000", cfg.Services["newservice"].Port)
	}

	// Duplicate name
	err = cfg.AddService("existing", ServiceConfig{Port: 5000})
	if err == nil {
		t.Error("expected error for duplicate service name")
	}

	// Duplicate port
	err = cfg.AddService("another", ServiceConfig{Port: 3000})
	if err == nil {
		t.Error("expected error for duplicate port")
	}

	// Invalid port
	err = cfg.AddService("badport", ServiceConfig{Port: -1})
	if err == nil {
		t.Error("expected error for negative port")
	}
}

func TestRemoveService(t *testing.T) {
	cfg := &PortmapConfig{
		Services: map[string]ServiceConfig{
			"api":    {Port: 3000},
			"worker": {Port: 4000},
		},
		Groups: map[string]GroupConfig{
			"backend": {Services: []string{"api", "worker"}},
		},
	}

	// Remove existing service
	err := cfg.RemoveService("api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, exists := cfg.Services["api"]; exists {
		t.Error("service should be deleted")
	}
	// Check group cleanup
	for _, s := range cfg.Groups["backend"].Services {
		if s == "api" {
			t.Error("removed service should be cleaned from groups")
		}
	}

	// Remove nonexistent
	err = cfg.RemoveService("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent service")
	}
}

func TestUpdatePort(t *testing.T) {
	cfg := &PortmapConfig{
		Services: map[string]ServiceConfig{
			"api":    {Port: 3000},
			"worker": {Port: 4000},
		},
	}

	// Valid update
	err := cfg.UpdatePort("api", 5000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Services["api"].Port != 5000 {
		t.Errorf("port = %d, want 5000", cfg.Services["api"].Port)
	}

	// Nonexistent service
	err = cfg.UpdatePort("missing", 6000)
	if err == nil {
		t.Error("expected error for nonexistent service")
	}

	// Port conflict
	err = cfg.UpdatePort("api", 4000)
	if err == nil {
		t.Error("expected error for port conflict")
	}

	// Invalid port
	err = cfg.UpdatePort("api", 0)
	if err == nil {
		t.Error("expected error for port 0")
	}

	err = cfg.UpdatePort("api", 70000)
	if err == nil {
		t.Error("expected error for port > 65535")
	}
}

func TestGetServicePorts(t *testing.T) {
	cfg := &PortmapConfig{
		Services: map[string]ServiceConfig{
			"api":    {Port: 3000},
			"worker": {Port: 4000},
			"noport": {Port: 0},
		},
	}
	ports := cfg.GetServicePorts()
	if len(ports) != 2 {
		t.Errorf("ports count = %d, want 2", len(ports))
	}
	if ports[3000] != "api" {
		t.Errorf("port 3000 = %q, want %q", ports[3000], "api")
	}
	if ports[4000] != "worker" {
		t.Errorf("port 4000 = %q, want %q", ports[4000], "worker")
	}
}
