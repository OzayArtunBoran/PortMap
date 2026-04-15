# Skill: CI/CD

## Amaç
GitHub Actions workflow'ları, GoReleaser config, ve pre-commit hook oluştur.

## Girdiler
- `project-spec.yml` → `ci_cd`, `stack`, `quality` bölümleri

---

## GitHub Actions — CI Workflow

### .github/workflows/ci.yml

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      # ── Go ──
      - uses: actions/setup-go@v5
        with:
          go-version: '{stack.version}'
      - uses: golangci/golangci-lint-action@v4
        with:
          version: latest

      # ── Python ──
      # - uses: actions/setup-python@v5
      #   with:
      #     python-version: '{stack.version}'
      # - run: pip install ruff
      # - run: ruff check .

  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      # ── Go ──
      - uses: actions/setup-go@v5
        with:
          go-version: '{stack.version}'
      - run: go test ./... -v -race -coverprofile=coverage.txt
      - uses: codecov/codecov-action@v4
        with:
          file: ./coverage.txt
        if: github.event_name == 'push'

      # ── Python ──
      # - uses: actions/setup-python@v5
      #   with:
      #     python-version: '{stack.version}'
      # - run: pip install -r backend/requirements.txt
      # - run: pytest -v --cov --cov-report=xml
      # - uses: codecov/codecov-action@v4
      #   if: github.event_name == 'push'

  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
      - uses: actions/checkout@v4

      # ── Go ──
      - uses: actions/setup-go@v5
        with:
          go-version: '{stack.version}'
      - run: go build -o /dev/null .

      # ── Docker ──
      - run: docker build -t {name}:ci .
```

---

## GitHub Actions — Release Workflow (Go projeleri)

### .github/workflows/release.yml

```yaml
name: Release

on:
  push:
    tags: ['v*']

permissions:
  contents: write
  packages: write

jobs:
  release:
    name: Release
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: '{stack.version}'

      - uses: goreleaser/goreleaser-action@v5
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

## GoReleaser Config (Go projeleri)

### .goreleaser.yml

```yaml
version: 2

builds:
  - binary: {binary_name}
    env:
      - CGO_ENABLED=0
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.buildTime={{.Date}}
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64

archives:
  - format: tar.gz
    name_template: >-
      {{.ProjectName}}_{{.Version}}_{{.Os}}_{{.Arch}}
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: 'checksums.txt'

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^chore:'

release:
  github:
    owner: {author.github}
    name: {project.name}
  draft: false
  prerelease: auto
```

---

## GitHub Actions — Docker Publish (opsiyonel)

### .github/workflows/docker.yml

```yaml
name: Docker

on:
  push:
    tags: ['v*']

permissions:
  packages: write

jobs:
  docker:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - uses: docker/metadata-action@v5
        id: meta
        with:
          images: ghcr.io/${{ github.repository }}
          tags: |
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha

      - uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
```

---

## Pre-commit Hook

### ci/pre-commit-hook.sh

```bash
#!/bin/bash
set -e

echo "🔍 Running pre-commit checks..."

# ── Go ──
echo "→ go vet"
go vet ./...

echo "→ go test"
go test ./... -count=1 -short

# ── Python ──
# echo "→ ruff check"
# ruff check .
# echo "→ pytest"
# pytest -x -q

echo "✅ All checks passed!"
```

### Kurulum talimatı (README'ye eklenecek)
```bash
# Hook'u kur
cp ci/pre-commit-hook.sh .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

---

## GitHub Actions — Composite Action (opsiyonel, reusable)

Eğer proje bir CI tool'u ise (DriftGuard gibi), başkalarının kullanması için composite action:

### ci/github-action.yml

```yaml
name: '{display_name}'
description: '{tagline}'

inputs:
  config-path:
    description: 'Config file path'
    required: false
    default: '.{name}.yml'
  fail-on-error:
    description: 'Fail the workflow on errors'
    required: false
    default: 'true'

runs:
  using: composite
  steps:
    - name: Install {display_name}
      shell: bash
      run: |
        curl -sSL https://github.com/{author.github}/{name}/releases/latest/download/{name}_linux_amd64.tar.gz | tar xz
        sudo mv {name} /usr/local/bin/

    - name: Run {display_name}
      shell: bash
      run: |
        {name} scan --config ${{ inputs.config-path }}
      continue-on-error: ${{ inputs.fail-on-error != 'true' }}
```

---

## Doğrulama
- YAML syntax geçerli: `python -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))"`
- GoReleaser config geçerli: `goreleaser check` (kuruluysa)
- Pre-commit hook çalıştırılabilir: `chmod +x ci/pre-commit-hook.sh && bash ci/pre-commit-hook.sh`
