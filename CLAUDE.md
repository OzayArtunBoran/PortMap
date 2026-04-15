# PortMap — CLAUDE.md

> Bu dosyayı her prompt başında oku. Projeyi baştan tarama — tüm bağlam burada.

---

## Proje Özeti
Local development port collision resolver. Shows what's running on each port,
detects conflicts, suggests free ports, and manages port assignments for all
your services via a config file.

## Tech Stack
- Language: Go 1.22+
- CLI: Cobra + Viper
- Module: github.com/ozayartunboran/portmap
- Config: .portmap.yml (YAML)
- No database, no frontend, no API server

## Proje Yapısı
```
portmap/
├── CLAUDE.md
├── project-spec.yml
├── skills/
├── cmd/
│   ├── root.go        # Root komut, global flags
│   ├── scan.go        # Port tarama
│   ├── check.go       # Config'e göre çakışma kontrolü
│   ├── find.go        # Boş port bulma
│   ├── info.go        # Port/PID detay bilgisi
│   ├── init_cmd.go    # Config dosyası oluşturma
│   ├── add.go         # Config'e servis ekleme
│   ├── remove.go      # Config'den servis silme
│   ├── list.go        # Config'deki servisleri listeleme
│   ├── watch.go       # Canlı port izleme (TUI)
│   └── kill.go        # Port'u kullanan process'i öldürme
├── internal/
│   ├── config/
│   │   └── config.go      # PortmapConfig, ServiceConfig, LoadConfig, SaveConfig
│   ├── scanner/
│   │   ├── scanner.go     # Scanner interface, PortInfo, ScanResult
│   │   ├── linux.go       # Linux: /proc/net/tcp parser
│   │   └── darwin.go      # macOS: lsof parser
│   ├── detector/
│   │   └── detector.go    # Conflict detection, ConflictType, CheckResult
│   ├── allocator/
│   │   └── allocator.go   # Free port finder, AllocStrategy
│   ├── process/
│   │   └── process.go     # Process detail lookup
│   └── formatter/
│       ├── formatter.go   # Formatter interface
│       ├── terminal.go    # Colored table output
│       ├── json.go        # JSON output
│       ├── markdown.go    # Markdown table output
│       └── compact.go     # One-line compact output
├── .portmap.example.yml
├── main.go
├── go.mod / go.sum
├── Makefile
└── .gitignore
```

## Key Types

### internal/config
- PortmapConfig: Version, Defaults, Services map, Groups map
- ServiceConfig: Port, Description, Command, HealthCheck, Managed
- ConfigDefaults: Range, Strategy

### internal/scanner
- PortInfo: Port, Protocol, PID, ProcessName, User, State, CommandLine, StartTime
- ScanResult: Ports, ScannedRange, Duration, Platform
- Scanner interface: Scan(opts) → ScanResult

### internal/detector
- ConflictType: OCCUPIED, DUPLICATE, RANGE_OVERLAP
- Conflict: Port, Type, ServiceName, ExpectedService, ActualProcess, ActualPID, Suggestion, Message
- CheckResult: Conflicts, OKServices, TotalChecked

### internal/allocator
- AllocStrategy: nearest, sequential, random
- AllocRequest: PreferredPort, Range, Strategy, Count, Exclude
- PortRange: Start, End

### internal/formatter
- Formatter interface: FormatScan, FormatCheck, FormatInfo

## Build & Run
```bash
make build    # → bin/portmap
make test     # → go test ./...
make lint     # → golangci-lint (veya go vet)
```

## Current Status
- [x] Phase 1: Scaffold
- [x] Phase 2: Core Features
- [ ] Phase 3: Tests
- [ ] Phase 4: API/UI (watch mode)
- [ ] Phase 5: Notifications (N/A — skip)
- [ ] Phase 6: Docker + CI/CD
- [ ] Phase 7: README + GitHub
