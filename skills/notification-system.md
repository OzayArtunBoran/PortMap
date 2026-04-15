# Skill: Notification System (Slack + Email + Discord + Telegram)

## Amaç
4 bildirim kanalını tek bir sistemde implement et. Ortak interface, kanal bazında format optimizasyonu.

## Girdiler
- `project-spec.yml` → `notifications` bölümü

## Kurallar
- Her kanal aynı interface'i / base class'ı implement eder
- Config eksikse kanal sessizce atlanır — hata logla ama diğerlerini engelleme
- HTTP çağrıları: 10 saniye timeout, 1 retry (exponential backoff yok, basit retry)
- Mesaj formatları kanal başına optimize edilir
- Tüm kanallar aynı Notification struct/model'ini alır, kendi formatına çevirir

---

## Ortak Veri Yapısı

### Go
```go
type Notification struct {
    Title    string            // Bildirim başlığı
    Summary  string            // Kısa özet (1-2 cümle)
    Details  []NotificationDetail // Detay listesi
    Severity string            // info | warning | critical
    Metadata map[string]string // Ek bilgiler (proje adı, zaman, vs.)
}

type NotificationDetail struct {
    Label string
    Value string
}
```

### Python
```python
from dataclasses import dataclass, field

@dataclass
class NotificationDetail:
    label: str
    value: str

@dataclass
class Notification:
    title: str
    summary: str
    details: list[NotificationDetail] = field(default_factory=list)
    severity: str = "info"  # info | warning | critical
    metadata: dict[str, str] = field(default_factory=dict)
```

---

## Interface / Base Class

### Go
```go
package notifier

type Notifier interface {
    Name() string
    Send(n Notification) error
    Validate() error
}

// NewNotifiers config'e göre aktif kanalları oluşturur
func NewNotifiers(cfg config.NotificationConfig) []Notifier {
    var notifiers []Notifier
    if cfg.Slack != nil && cfg.Slack.WebhookURL != "" {
        notifiers = append(notifiers, NewSlackNotifier(cfg.Slack))
    }
    if cfg.Email != nil && cfg.Email.SMTPHost != "" {
        notifiers = append(notifiers, NewEmailNotifier(cfg.Email))
    }
    if cfg.Discord != nil && cfg.Discord.WebhookURL != "" {
        notifiers = append(notifiers, NewDiscordNotifier(cfg.Discord))
    }
    if cfg.Telegram != nil && cfg.Telegram.BotToken != "" {
        notifiers = append(notifiers, NewTelegramNotifier(cfg.Telegram))
    }
    return notifiers
}

// SendAll tüm kanallara gönderir, hataları toplar
func SendAll(notifiers []Notifier, n Notification) []error {
    var errs []error
    for _, notifier := range notifiers {
        if err := notifier.Send(n); err != nil {
            log.Printf("[WARN] notification error (%s): %v", notifier.Name(), err)
            errs = append(errs, fmt.Errorf("%s: %w", notifier.Name(), err))
        }
    }
    return errs
}
```

### Python
```python
from abc import ABC, abstractmethod
import logging

logger = logging.getLogger(__name__)

class BaseNotifier(ABC):
    @property
    @abstractmethod
    def name(self) -> str: ...

    @abstractmethod
    async def send(self, notification: Notification) -> bool: ...

    @abstractmethod
    def validate(self) -> bool: ...

async def send_all(notifiers: list[BaseNotifier], notification: Notification) -> list[str]:
    errors = []
    for notifier in notifiers:
        try:
            await notifier.send(notification)
        except Exception as e:
            logger.warning(f"Notification error ({notifier.name}): {e}")
            errors.append(f"{notifier.name}: {e}")
    return errors
```

---

## Kanal Implementasyonları

### 1. Slack

**Yöntem:** HTTP POST → Webhook URL
**Format:** Block Kit (zengin mesaj)

```go
type SlackNotifier struct {
    webhookURL string
    channel    string
}

func (s *SlackNotifier) Send(n Notification) error {
    // Severity → renk
    colors := map[string]string{
        "info": "#36a64f", "warning": "#daa038", "critical": "#a30200",
    }

    // Block Kit payload
    payload := map[string]interface{}{
        "blocks": []map[string]interface{}{
            {"type": "header", "text": map[string]string{"type": "plain_text", "text": n.Title}},
            {"type": "section", "text": map[string]string{"type": "mrkdwn", "text": n.Summary}},
        },
        "attachments": []map[string]interface{}{
            {
                "color": colors[n.Severity],
                "fields": buildSlackFields(n.Details[:min(5, len(n.Details))]),
            },
        },
    }

    return httpPost(s.webhookURL, payload, 10*time.Second)
}
```

### 2. Email

**Yöntem:** SMTP + TLS
**Format:** HTML template

```go
type EmailNotifier struct {
    host       string
    port       int
    username   string
    password   string
    from       string
    recipients []string
}

func (e *EmailNotifier) Send(n Notification) error {
    // HTML template
    htmlBody := buildEmailHTML(n)

    // SMTP bağlantısı
    auth := smtp.PlainAuth("", e.username, e.password, e.host)

    msg := buildMIMEMessage(e.from, e.recipients, n.Title, htmlBody)

    addr := fmt.Sprintf("%s:%d", e.host, e.port)

    // Port'a göre TLS stratejisi
    if e.port == 465 {
        // Implicit TLS
        return sendWithTLS(addr, auth, e.from, e.recipients, msg)
    }
    // STARTTLS (port 587)
    return smtp.SendMail(addr, auth, e.from, e.recipients, msg)
}
```

**HTML template yapısı:**
```html
<div style="font-family: -apple-system, sans-serif; max-width: 600px; margin: 0 auto;">
  <h2 style="color: {severity_color};">{title}</h2>
  <p>{summary}</p>
  <table style="width: 100%; border-collapse: collapse;">
    <tr style="background: #f5f5f5;">
      <th style="padding: 8px; text-align: left;">Item</th>
      <th style="padding: 8px; text-align: left;">Value</th>
    </tr>
    <!-- details loop -->
  </table>
  <hr>
  <p style="color: #888; font-size: 12px;">Sent by {project.display_name}</p>
</div>
```

### 3. Discord

**Yöntem:** HTTP POST → Webhook URL
**Format:** Embed (zengin kart)

```go
type DiscordNotifier struct {
    webhookURL string
}

func (d *DiscordNotifier) Send(n Notification) error {
    colors := map[string]int{
        "info": 0x3B82F6, "warning": 0xF59E0B, "critical": 0xEF4444,
    }

    fields := make([]map[string]interface{}, 0, min(10, len(n.Details)))
    for _, detail := range n.Details[:min(10, len(n.Details))] {
        fields = append(fields, map[string]interface{}{
            "name": detail.Label, "value": detail.Value, "inline": true,
        })
    }

    payload := map[string]interface{}{
        "embeds": []map[string]interface{}{
            {
                "title":       n.Title,
                "description": n.Summary,
                "color":       colors[n.Severity],
                "fields":      fields,
                "timestamp":   time.Now().UTC().Format(time.RFC3339),
            },
        },
    }

    return httpPost(d.webhookURL, payload, 10*time.Second)
}
```

### 4. Telegram

**Yöntem:** HTTP POST → Bot API
**Format:** HTML parse mode

```go
type TelegramNotifier struct {
    botToken string
    chatID   string
}

func (t *TelegramNotifier) Send(n Notification) error {
    severityEmoji := map[string]string{
        "info": "ℹ️", "warning": "⚠️", "critical": "🚨",
    }

    var msg strings.Builder
    msg.WriteString(fmt.Sprintf("<b>%s %s</b>\n\n", severityEmoji[n.Severity], n.Title))
    msg.WriteString(fmt.Sprintf("%s\n\n", n.Summary))

    for _, d := range n.Details[:min(5, len(n.Details))] {
        msg.WriteString(fmt.Sprintf("• <b>%s:</b> %s\n", d.Label, d.Value))
    }

    url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)
    payload := map[string]interface{}{
        "chat_id":    t.chatID,
        "text":       msg.String(),
        "parse_mode": "HTML",
    }

    return httpPost(url, payload, 10*time.Second)
}
```

---

## HTTP Helper (ortak)

```go
func httpPost(url string, payload interface{}, timeout time.Duration) error {
    body, _ := json.Marshal(payload)

    client := &http.Client{Timeout: timeout}

    var lastErr error
    for attempt := 0; attempt < 2; attempt++ { // 1 retry
        resp, err := client.Post(url, "application/json", bytes.NewReader(body))
        if err != nil {
            lastErr = err
            time.Sleep(time.Second) // basit bekleme
            continue
        }
        defer resp.Body.Close()

        if resp.StatusCode >= 200 && resp.StatusCode < 300 {
            return nil
        }
        respBody, _ := io.ReadAll(resp.Body)
        lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
    }

    return fmt.Errorf("after retries: %w", lastErr)
}
```

---

## Escalation (Opsiyonel)

project-spec'te `escalation.enabled: true` ise:

```
İlk bildirim: Slack
5 dakika sonra onay yoksa → Email
15 dakika sonra → Telegram (acil durum)
```

Escalation, bildirim gönderimi sırasında değil, ayrı bir scheduled task/goroutine ile yönetilir.

---

## Config Pattern (YAML)

```yaml
notifications:
  slack:
    webhook_url: "${SLACK_WEBHOOK_URL}"
    channel: "#alerts"
  email:
    smtp_host: "smtp.gmail.com"
    smtp_port: 587
    username: "${SMTP_USERNAME}"
    password: "${SMTP_PASSWORD}"
    from: "alerts@example.com"
    recipients:
      - "admin@example.com"
  discord:
    webhook_url: "${DISCORD_WEBHOOK_URL}"
  telegram:
    bot_token: "${TELEGRAM_BOT_TOKEN}"
    chat_id: "${TELEGRAM_CHAT_ID}"
```

---

## Doğrulama
- `Validate()` eksik config'i doğru yakalar (webhook URL boş, SMTP host yok)
- `Send()` hatalı webhook'ta panic yapmaz, hata döner
- `NewNotifiers()` config olmayan kanalları atlar
- `SendAll()` bir kanal hata verse diğerleri çalışmaya devam eder
- Unit testler: HTTP mock ile her kanal test edilir (httptest.NewServer)
