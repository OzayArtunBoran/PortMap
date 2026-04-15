# Skill: Go CLI (Cobra + Viper)

## Amaç
Cobra + Viper tabanlı CLI uygulaması oluştur.

## Girdiler
- `project-spec.yml` → `cli`, `stack.go` bölümleri

## Kurallar
- Her komut ayrı dosyada: `cmd/{command}.go`
- Root komut global flag'leri tanımlar: `--config`, `--verbose`, `--no-color`, `--version`
- Build-time version injection: `var version, buildTime string` (`main.go`'da)
- Config dosyası YAML formatında, Viper ile okunur
- Tüm iş mantığı `internal/` altında — `cmd/` sadece CLI arayüzü, iş mantığı içermez
- Hata yönetimi: `RunE` kullan (`Run` değil), hataları yukarı propagate et

## Yapı

```
cmd/
├── root.go          # Root komut, global flag'ler, Viper init
├── {command1}.go    # Her komut kendi dosyasında
├── {command2}.go
└── ...
internal/
├── config/
│   └── config.go    # Config struct'ları + LoadConfig + SaveConfig
├── {domain1}/
│   └── {domain1}.go # İş mantığı modülü
├── {domain2}/
│   └── {domain2}.go
└── ...
main.go              # cmd.Execute() çağrısı + version vars
```

## Dosya İçerikleri

### main.go
```go
package main

import (
    "{module}/cmd"
)

var (
    version   string
    buildTime string
)

func main() {
    cmd.SetVersionInfo(version, buildTime)
    cmd.Execute()
}
```

### cmd/root.go
```go
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
    
    ver       string
    buildAt   string
)

func SetVersionInfo(v, b string) {
    ver = v
    buildAt = b
}

var rootCmd = &cobra.Command{
    Use:   "{binary_name}",
    Short: "{project.tagline}",
    Long:  "{project.description}",
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}

func init() {
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", ".{name}.yml", "config file path")
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
    rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")
    
    rootCmd.Version = ver
    rootCmd.SetVersionTemplate(fmt.Sprintf("{{.Name}} %s (built %s)\n", ver, buildAt))
}
```

### cmd/{command}.go — Komut dosyası pattern
```go
package cmd

import (
    "fmt"
    
    "{module}/internal/config"
    "{module}/internal/{domain}"
    "github.com/spf13/cobra"
)

var {command}Flags struct {
    flagName string
    // ...project-spec'ten gelen flag'ler
}

var {command}Cmd = &cobra.Command{
    Use:   "{name} [args]",
    Short: "{description}",
    Long:  "{detaylı açıklama}",
    Args:  cobra.{argValidator},  // NoArgs, ExactArgs(n), MinimumNArgs(n)
    RunE:  run{Command},
}

func init() {
    rootCmd.AddCommand({command}Cmd)
    
    {command}Cmd.Flags().StringVarP(&{command}Flags.flagName, "flag", "f", "default", "description")
    // ...diğer flag'ler
}

func run{Command}(cmd *cobra.Command, args []string) error {
    // 1. Config yükle (gerekiyorsa)
    cfg, err := config.LoadConfig(cfgFile)
    if err != nil {
        return fmt.Errorf("config load: %w", err)
    }
    
    // 2. İş mantığını çağır (internal/ altından)
    result, err := {domain}.DoSomething(cfg, args)
    if err != nil {
        return fmt.Errorf("{operation}: %w", err)
    }
    
    // 3. Çıktı formatla
    // ...
    
    return nil
}
```

### internal/config/config.go — Config yönetimi
```go
package config

import (
    "fmt"
    "os"
    
    "gopkg.in/yaml.v3"
)

type Config struct {
    // project-spec.yml'den gelen tüm config alanları
    Version string `yaml:"version"`
    // ...domain-specific fields
}

// LoadConfig YAML dosyasını okur ve env var expansion yapar
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read config: %w", err)
    }
    
    // Env var expansion: ${VAR_NAME}
    expanded := os.ExpandEnv(string(data))
    
    var cfg Config
    if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
        return nil, fmt.Errorf("parse config: %w", err)
    }
    
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("validate config: %w", err)
    }
    
    return &cfg, nil
}

// SaveConfig config'i YAML dosyasına yazar
func SaveConfig(path string, cfg *Config) error {
    data, err := yaml.Marshal(cfg)
    if err != nil {
        return fmt.Errorf("marshal config: %w", err)
    }
    return os.WriteFile(path, data, 0644)
}

// Validate config'in tutarlılığını kontrol eder
func (c *Config) Validate() error {
    // Domain-specific validasyonlar
    return nil
}
```

### init komutu pattern (interaktif)
```go
func runInit(cmd *cobra.Command, args []string) error {
    reader := bufio.NewReader(os.Stdin)
    
    // Mevcut config varsa onay iste
    if _, err := os.Stat(cfgFile); err == nil {
        fmt.Printf("Config file %s already exists. Overwrite? [y/N]: ", cfgFile)
        answer, _ := reader.ReadString('\n')
        if strings.TrimSpace(strings.ToLower(answer)) != "y" {
            fmt.Println("Aborted.")
            return nil
        }
    }
    
    cfg := &config.Config{Version: "1"}
    
    // İnteraktif sorular
    fmt.Print("Project name: ")
    name, _ := reader.ReadString('\n')
    cfg.ProjectName = strings.TrimSpace(name)
    
    // ... diğer sorular
    
    return config.SaveConfig(cfgFile, cfg)
}
```

## Bağımlılıklar
```bash
go get github.com/spf13/cobra@latest
go get github.com/spf13/viper@latest
go get gopkg.in/yaml.v3
```

## Doğrulama
- `go build ./...` hatasız
- `./bin/{name} --help` tüm komutları gösterir
- `./bin/{name} --version` versiyon bilgisi gösterir
- Her komutun `--help` çıktısı anlamlı flag açıklamalarına sahip
- `./bin/{name} init` interaktif akış çalışıyor
