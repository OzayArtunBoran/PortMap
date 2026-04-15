package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// PortmapConfig is the root configuration structure
type PortmapConfig struct {
	Version  string                   `yaml:"version"`
	Defaults ConfigDefaults           `yaml:"defaults"`
	Services map[string]ServiceConfig `yaml:"services"`
	Groups   map[string]GroupConfig   `yaml:"groups"`
}

// ConfigDefaults holds default settings for port allocation
type ConfigDefaults struct {
	Range    string `yaml:"range"`    // "3000-9999"
	Strategy string `yaml:"strategy"` // nearest | sequential | random
}

// ServiceConfig defines a single service's port assignment
type ServiceConfig struct {
	Port        int    `yaml:"port"`
	Description string `yaml:"description,omitempty"`
	Command     string `yaml:"command,omitempty"`
	HealthCheck string `yaml:"health_check,omitempty"`
	Managed     *bool  `yaml:"managed,omitempty"` // nil = true (default)
}

// IsManaged returns whether this service is managed by portmap (default: true)
func (s ServiceConfig) IsManaged() bool {
	if s.Managed == nil {
		return true
	}
	return *s.Managed
}

// GroupConfig defines a named group of services
type GroupConfig struct {
	Services []string `yaml:"services"`
}

// PortRange represents a range of port numbers
type PortRange struct {
	Start int
	End   int
}

// ParseRange parses a "start-end" string into a PortRange
func ParseRange(s string) (PortRange, error) {
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return PortRange{}, fmt.Errorf("invalid range format %q, expected start-end", s)
	}

	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return PortRange{}, fmt.Errorf("invalid range start %q: %w", parts[0], err)
	}

	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return PortRange{}, fmt.Errorf("invalid range end %q: %w", parts[1], err)
	}

	if start < 1 || end > 65535 {
		return PortRange{}, fmt.Errorf("port range must be between 1 and 65535, got %d-%d", start, end)
	}

	if start >= end {
		return PortRange{}, fmt.Errorf("range start (%d) must be less than end (%d)", start, end)
	}

	return PortRange{Start: start, End: end}, nil
}

// Contains checks if a port is within this range
func (r PortRange) Contains(port int) bool {
	return port >= r.Start && port <= r.End
}

// LoadConfig reads and parses a .portmap.yml file
func LoadConfig(path string) (*PortmapConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	// Env var expansion: ${VAR_NAME}
	expanded := os.ExpandEnv(string(data))

	var cfg PortmapConfig
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Set defaults
	if cfg.Version == "" {
		cfg.Version = "1"
	}
	if cfg.Defaults.Range == "" {
		cfg.Defaults.Range = "3000-9999"
	}
	if cfg.Defaults.Strategy == "" {
		cfg.Defaults.Strategy = "nearest"
	}
	if cfg.Services == nil {
		cfg.Services = make(map[string]ServiceConfig)
	}
	if cfg.Groups == nil {
		cfg.Groups = make(map[string]GroupConfig)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
}

// SaveConfig writes the config to a YAML file
func SaveConfig(path string, cfg *PortmapConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	header := "# PortMap configuration\n# Docs: https://github.com/ozayartunboran/portmap\n\n"
	content := header + string(data)

	return os.WriteFile(path, []byte(content), 0644)
}

// Validate checks config consistency
func (c *PortmapConfig) Validate() error {
	// Validate default range
	if _, err := ParseRange(c.Defaults.Range); err != nil {
		return fmt.Errorf("defaults.range: %w", err)
	}

	// Validate strategy
	validStrategies := map[string]bool{"nearest": true, "sequential": true, "random": true}
	if !validStrategies[c.Defaults.Strategy] {
		return fmt.Errorf("defaults.strategy must be nearest, sequential, or random, got %q", c.Defaults.Strategy)
	}

	// Check for duplicate ports
	portMap := make(map[int]string)
	for name, svc := range c.Services {
		if svc.Port < 0 || svc.Port > 65535 {
			return fmt.Errorf("service %q: port %d out of range (1-65535)", name, svc.Port)
		}
		if svc.Port > 0 {
			if existing, ok := portMap[svc.Port]; ok {
				return fmt.Errorf("duplicate port %d: assigned to both %q and %q", svc.Port, existing, name)
			}
			portMap[svc.Port] = name
		}
	}

	// Validate group references
	for gName, group := range c.Groups {
		for _, svcName := range group.Services {
			if _, exists := c.Services[svcName]; !exists {
				return fmt.Errorf("group %q references unknown service %q", gName, svcName)
			}
		}
	}

	return nil
}

// AddService adds a service to the config with validation
func (c *PortmapConfig) AddService(name string, svc ServiceConfig) error {
	if _, exists := c.Services[name]; exists {
		return fmt.Errorf("service %q already exists", name)
	}
	if svc.Port < 0 || svc.Port > 65535 {
		return fmt.Errorf("port %d out of range (1-65535)", svc.Port)
	}
	// Check for port conflicts
	if svc.Port > 0 {
		for existingName, existingSvc := range c.Services {
			if existingSvc.Port == svc.Port {
				return fmt.Errorf("port %d already assigned to %q", svc.Port, existingName)
			}
		}
	}
	c.Services[name] = svc
	return nil
}

// RemoveService removes a service and cleans up group references
func (c *PortmapConfig) RemoveService(name string) error {
	if _, exists := c.Services[name]; !exists {
		return fmt.Errorf("service %q not found", name)
	}
	delete(c.Services, name)

	// Clean up group references
	for gName, group := range c.Groups {
		filtered := make([]string, 0, len(group.Services))
		for _, s := range group.Services {
			if s != name {
				filtered = append(filtered, s)
			}
		}
		group.Services = filtered
		c.Groups[gName] = group
	}

	return nil
}

// UpdatePort updates a service's port with validation
func (c *PortmapConfig) UpdatePort(name string, newPort int) error {
	svc, exists := c.Services[name]
	if !exists {
		return fmt.Errorf("service %q not found", name)
	}
	if newPort < 1 || newPort > 65535 {
		return fmt.Errorf("port %d out of range (1-65535)", newPort)
	}
	// Check for conflicts
	for existingName, existingSvc := range c.Services {
		if existingName != name && existingSvc.Port == newPort {
			return fmt.Errorf("port %d already assigned to %q", newPort, existingName)
		}
	}
	svc.Port = newPort
	c.Services[name] = svc
	return nil
}

// GetServicePorts returns a port → service name mapping
func (c *PortmapConfig) GetServicePorts() map[int]string {
	result := make(map[int]string)
	for name, svc := range c.Services {
		if svc.Port > 0 {
			result[svc.Port] = name
		}
	}
	return result
}
