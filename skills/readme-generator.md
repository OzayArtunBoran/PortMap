# Skill: README Generator

## Amaç
Profesyonel, kapsamlı, kopyala-yapıştır çalışır README.md oluştur.

## Girdiler
- `project-spec.yml` → `readme`, `project`, `cli`, `api`, `docker`, `notifications` bölümleri

## Dil
İngilizce (her zaman).

---

## README Yapısı

```markdown
# {display_name}

> {tagline}

[![Go Version](https://img.shields.io/badge/go-{version}-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-{license}-blue)](LICENSE)
[![CI](https://github.com/{author.github}/{name}/actions/workflows/ci.yml/badge.svg)](https://github.com/{author.github}/{name}/actions)
[![Docker](https://img.shields.io/badge/docker-ready-2496ED?logo=docker)](Dockerfile)

---

## The Problem

{2-3 cümle: bu araç olmadan hayat nasıl zor. Empati kur, teknik jargon az.}

## The Solution

{display_name} {ne yapar — 2-3 cümle}.

**Key features:**
- ✅ Feature 1 — kısa açıklama
- ✅ Feature 2 — kısa açıklama
- ✅ Feature 3 — kısa açıklama
- ✅ Feature 4 — kısa açıklama

## Quick Start

### Install

**Go install:**
```bash
go install github.com/{author.github}/{name}@latest
```

**Binary download:**
```bash
curl -sSL https://github.com/{author.github}/{name}/releases/latest/download/{name}_linux_amd64.tar.gz | tar xz
sudo mv {name} /usr/local/bin/
```

**Docker:**
```bash
docker pull ghcr.io/{author.github}/{name}:latest
```

### First Run

```bash
# Initialize config
{binary_name} init

# Run
{binary_name} {primary_command}
```

**Example output:**
```
{gerçekçi terminal çıktısı — renkli olmasa da yapıyı gösterir}
```

## Configuration

{binary_name} uses a `.{name}.yml` config file:

```yaml
{tam config örneği — tüm alanlar açıklamalı}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `field1` | string | `""` | Açıklama |
| `field2` | int | `0` | Açıklama |

## Usage

### CLI Reference

| Command | Description |
|---------|-------------|
| `{name} {cmd1}` | {açıklama} |
| `{name} {cmd2}` | {açıklama} |
| `{name} {cmd3}` | {açıklama} |

### Examples

```bash
# Örnek 1
{name} {command} --flag value

# Örnek 2
{name} {command} --format json | jq .
```

## Docker

### Build & Run

```bash
docker build -t {name} .
docker run -v $(pwd)/.{name}.yml:/config/.{name}.yml {name}
```

### Docker Compose

```bash
docker compose up -d
```

## Notifications

{display_name} supports 4 notification channels:

### Slack
```yaml
notifications:
  slack:
    webhook_url: "${SLACK_WEBHOOK_URL}"
    channel: "#alerts"
```

### Email
```yaml
notifications:
  email:
    smtp_host: "smtp.gmail.com"
    smtp_port: 587
    username: "${SMTP_USERNAME}"
    password: "${SMTP_PASSWORD}"
    from: "alerts@example.com"
    recipients: ["admin@example.com"]
```

### Discord
```yaml
notifications:
  discord:
    webhook_url: "${DISCORD_WEBHOOK_URL}"
```

### Telegram
```yaml
notifications:
  telegram:
    bot_token: "${TELEGRAM_BOT_TOKEN}"
    chat_id: "${TELEGRAM_CHAT_ID}"
```

## CI/CD Integration

### GitHub Actions

```yaml
- uses: {author.github}/{name}-action@v1
  with:
    config-path: '.{name}.yml'
```

### Pre-commit Hook

```bash
cp ci/pre-commit-hook.sh .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

## Development

```bash
# Clone
git clone https://github.com/{author.github}/{name}.git
cd {name}

# Build
make build

# Test
make test

# Lint
make lint
```

## Contributing

Contributions welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md).

1. Fork the repo
2. Create a feature branch (`git checkout -b feat/amazing-feature`)
3. Commit changes (`git commit -m 'feat: add amazing feature'`)
4. Push (`git push origin feat/amazing-feature`)
5. Open a Pull Request

## License

{license} — see [LICENSE](LICENSE) for details.

---

Built by [{author.name}]({author.website})
```

---

## Kurallar

1. **Badge'ler** shields.io formatında, gerçek URL'lerle
2. **Kod örnekleri** kopyala-yapıştır çalışır olmalı — hatalı syntax yok
3. **Quick Start** 5 dakikada tamamlanabilir olmalı
4. **Config örneği** tüm alanları göstermeli, yorum satırlarıyla
5. **CLI tablosu** tüm komutları içermeli
6. **Bildirim bölümü** sadece `notifications.enabled: true` ise eklenir
7. **CI/CD bölümü** sadece `ci_cd.github_actions.enabled: true` ise eklenir
8. **Docker bölümü** sadece `docker.dockerfile` tanımlıysa eklenir
9. **Terminal çıktısı** gerçekçi, yapıyı gösteren örnek
10. **Footer** her zaman "Built by" ile biter

---

## Ek Bölümler (opsiyonel)

- **Architecture:** Sistem diyagramı (metin olarak — ASCII veya Mermaid)
- **Screenshots:** Dashboard varsa placeholder'lar
- **Self-Hosting Guide:** SaaS projeler için (min gereksinimler, reverse proxy, SSL, backup)
- **API Reference:** API projeleri için endpoint tablosu + Swagger link
- **FAQ:** Sık sorulan sorular

---

## Doğrulama
- Tüm linkler çalışıyor (badge URL'leri, repo linkleri)
- Kod blokları doğru dil etiketiyle işaretli (```bash, ```yaml, ```go)
- Bölüm sırası mantıklı: problem → çözüm → kurulum → kullanım → katkı
- Gereksiz bölüm yok (bildirim yoksa bildirim bölümü yok)
