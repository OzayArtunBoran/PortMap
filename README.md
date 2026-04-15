# PortMap

> Never fight over ports again

[![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![CI](https://github.com/ozayartunboran/portmap/actions/workflows/ci.yml/badge.svg)](https://github.com/ozayartunboran/portmap/actions)

---

## The Problem

You're running 5 services locally. You start your frontend — port 3000 is already in use. You check, kill, restart. Repeat daily. Tracking which service uses which port across your stack is a constant source of friction that breaks your flow.

## The Solution

PortMap is a CLI tool that scans your local ports, detects conflicts against your project's port assignments, suggests free alternatives, and manages everything through a simple config file.

**Key features:**

- Scan active ports and see which process owns each one
- Detect conflicts between your config and what's actually running
- Find free ports with smart allocation strategies (nearest, sequential, random)
- Live monitoring with a real-time terminal dashboard
- Kill processes occupying specific ports
- Manage service-to-port mappings via `.portmap.yml`

## Quick Start

### Install

```bash
go install github.com/ozayartunboran/portmap@latest
```

### First Run

```bash
# See what's running on your ports
portmap scan

# Create a config file for your project
portmap init --detect

# Check for conflicts
portmap check
```

**Example output:**

```
PORT   PROTO  PID    PROCESS    USER      STATE
3000   tcp    12834  node       dev       LISTEN
5432   tcp    1021   postgres   postgres  LISTEN
6379   tcp    1045   redis      redis     LISTEN
8080   tcp    15672  main       dev       LISTEN

Scanned 1-65535 in 124ms | 4 active ports | linux
```

## Configuration

PortMap uses a `.portmap.yml` config file:

```yaml
version: "1"

defaults:
  range: "3000-9999"
  strategy: "nearest"    # nearest | sequential | random

services:
  frontend:
    port: 3000
    description: "React dev server"
    command: "npm run dev"
    health_check: "http://localhost:3000"

  api:
    port: 8080
    description: "Go API server"
    command: "go run ."
    health_check: "http://localhost:8080/health"

  database:
    port: 5432
    description: "PostgreSQL"
    managed: false         # portmap won't start/stop this service

  redis:
    port: 6379
    description: "Redis"
    managed: false

groups:
  full-stack:
    services: ["frontend", "api", "database", "redis"]

  backend-only:
    services: ["api", "database", "redis"]
```

| Field | Type | Description |
|-------|------|-------------|
| `version` | string | Config format version |
| `defaults.range` | string | Default port search range |
| `defaults.strategy` | string | Allocation strategy: `nearest`, `sequential`, `random` |
| `services.<name>.port` | int | Assigned port number |
| `services.<name>.description` | string | Human-readable service description |
| `services.<name>.command` | string | Command to start the service |
| `services.<name>.health_check` | string | URL for health check |
| `services.<name>.managed` | bool | Whether portmap manages this service (default: true) |
| `groups.<name>.services` | list | List of service names in this group |

## Usage

### CLI Reference

| Command | Description | Key Flags |
|---------|-------------|-----------|
| `portmap scan` | Scan active ports | `--range`, `--format`, `--filter`, `--tcp-only`, `--listen-only` |
| `portmap check` | Check for port conflicts | `--config`, `--fix`, `--format` |
| `portmap find` | Find free ports | `--near`, `--count`, `--range` |
| `portmap info` | Show port/process details | `--port`, `--pid`, `--format` |
| `portmap init` | Create `.portmap.yml` | `--detect` |
| `portmap add <name>` | Add a service to config | `--port`, `--description`, `--command` |
| `portmap remove <name>` | Remove a service from config | — |
| `portmap list` | List configured services | `--format` |
| `portmap watch` | Live port monitoring | `--interval`, `--range` |
| `portmap kill <port>` | Kill process on a port | `--force`, `--yes` |

### Examples

```bash
# Scan a specific range
portmap scan --range 3000-9999

# Pipe JSON output to jq
portmap scan --format json | jq '.ports[] | select(.state == "LISTEN")'

# Auto-fix conflicts by reassigning ports
portmap check --fix

# Find 5 free ports near 3000
portmap find --near 3000 --count 5

# Get detailed info about port 8080
portmap info --port 8080

# Kill whatever is on port 3000
portmap kill 3000

# Watch ports with 1-second refresh
portmap watch --interval 1
```

## Docker

### Build & Run

```bash
docker build -t portmap .
docker run --rm portmap scan
```

Mount a config file:

```bash
docker run --rm -v $(pwd)/.portmap.yml:/config/.portmap.yml portmap check --config /config/.portmap.yml
```

## Development

```bash
# Clone
git clone https://github.com/ozayartunboran/portmap.git
cd portmap

# Build
make build

# Test
make test

# Lint
make lint
```

## Contributing

Contributions welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT — see [LICENSE](LICENSE) for details.

---

Built by [Özay Artun Boran](https://ozayartunboran.com)
