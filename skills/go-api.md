# Skill: Go API (Echo Framework)

## Amaç
Echo framework ile REST API oluştur. SQLite veya PostgreSQL backend.

## Girdiler
- `project-spec.yml` → `api`, `database_schema` bölümleri

## Kurallar
- Echo v4 kullan
- Raw SQL tercih et, ORM kullanma
- CGO-free SQLite driver: `modernc.org/sqlite` veya `github.com/ncruces/go-sqlite3`
- Response format standart: `{"success": bool, "data": ..., "error": ...}`
- Middleware: CORS, request logging, recovery, Content-Type
- Graceful shutdown: SIGINT, SIGTERM
- Context timeout: her request için 30s

## Yapı

```
internal/store/
├── store.go         # Store interface
└── sqlite.go        # SQLite implementasyonu (veya postgres.go)
web/
├── server.go        # Echo server setup, route registration, middleware
├── handlers/
│   ├── {resource}.go # Her resource'un handler'ları
│   └── ...
├── middleware/
│   └── auth.go      # Auth middleware (gerekiyorsa)
└── frontend/        # Embed edilecek frontend (varsa)
    └── dist/
```

## Dosya İçerikleri

### web/server.go
```go
package web

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "{module}/internal/config"
    "{module}/internal/store"
)

type Server struct {
    echo   *echo.Echo
    store  store.Store
    config *config.Config
}

func NewServer(cfg *config.Config, s store.Store) *Server {
    e := echo.New()
    e.HideBanner = true
    
    // Middleware
    e.Use(middleware.Recover())
    e.Use(middleware.RequestID())
    e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
        Format: "${time_rfc3339} ${method} ${uri} ${status} ${latency_human}\n",
    }))
    e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
        AllowOrigins: []string{"*"},
        AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
    }))
    
    srv := &Server{echo: e, store: s, config: cfg}
    srv.registerRoutes()
    return srv
}

func (s *Server) registerRoutes() {
    api := s.echo.Group("/api")
    
    api.GET("/health", s.handleHealth)
    // ... project-spec'ten gelen endpoint'ler
}

func (s *Server) Start(addr string) error {
    // Graceful shutdown
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    go func() {
        if err := s.echo.Start(addr); err != nil && err != http.ErrServerClosed {
            s.echo.Logger.Fatal("server error:", err)
        }
    }()

    <-ctx.Done()
    fmt.Println("\nShutting down...")
    
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    return s.echo.Shutdown(shutdownCtx)
}
```

### Response helpers
```go
package web

import (
    "github.com/labstack/echo/v4"
)

type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *APIError   `json:"error,omitempty"`
}

type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

func success(c echo.Context, data interface{}) error {
    return c.JSON(200, APIResponse{Success: true, Data: data})
}

func successStatus(c echo.Context, status int, data interface{}) error {
    return c.JSON(status, APIResponse{Success: true, Data: data})
}

func fail(c echo.Context, status int, code, msg string) error {
    return c.JSON(status, APIResponse{
        Success: false,
        Error:   &APIError{Code: code, Message: msg},
    })
}

func handleHealth(c echo.Context) error {
    return success(c, map[string]string{"status": "ok"})
}
```

### internal/store/store.go — Store interface
```go
package store

type Store interface {
    Migrate() error
    Close() error
    // Resource-specific methods...
    // project-spec'teki database_schema'dan üretilir
}
```

### internal/store/sqlite.go — SQLite implementasyonu
```go
package store

import (
    "database/sql"
    "fmt"
    
    _ "modernc.org/sqlite"
)

type SQLiteStore struct {
    db *sql.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
    db, err := sql.Open("sqlite", dbPath)
    if err != nil {
        return nil, fmt.Errorf("open db: %w", err)
    }
    
    // SQLite optimizasyonları
    db.Exec("PRAGMA journal_mode=WAL")
    db.Exec("PRAGMA foreign_keys=ON")
    db.Exec("PRAGMA busy_timeout=5000")
    
    s := &SQLiteStore{db: db}
    if err := s.Migrate(); err != nil {
        return nil, fmt.Errorf("migrate: %w", err)
    }
    return s, nil
}

func (s *SQLiteStore) Migrate() error {
    queries := []string{
        // project-spec'teki database_schema'dan üretilir
        `CREATE TABLE IF NOT EXISTS ... (
            ...
        )`,
    }
    for _, q := range queries {
        if _, err := s.db.Exec(q); err != nil {
            return fmt.Errorf("migration: %w", err)
        }
    }
    return nil
}

func (s *SQLiteStore) Close() error {
    return s.db.Close()
}
```

### cmd/server.go — Server komutu
```go
var serverCmd = &cobra.Command{
    Use:   "server",
    Short: "Start the dashboard server",
    RunE:  runServer,
}

var serverFlags struct {
    port   int
    host   string
    dbPath string
}

func init() {
    rootCmd.AddCommand(serverCmd)
    serverCmd.Flags().IntVar(&serverFlags.port, "port", 9090, "server port")
    serverCmd.Flags().StringVar(&serverFlags.host, "host", "0.0.0.0", "server host")
    serverCmd.Flags().StringVar(&serverFlags.dbPath, "db-path", "{name}.db", "database file path")
}

func runServer(cmd *cobra.Command, args []string) error {
    cfg, err := config.LoadConfig(cfgFile)
    if err != nil {
        return err
    }
    
    db, err := store.NewSQLiteStore(serverFlags.dbPath)
    if err != nil {
        return fmt.Errorf("init db: %w", err)
    }
    defer db.Close()
    
    srv := web.NewServer(cfg, db)
    addr := fmt.Sprintf("%s:%d", serverFlags.host, serverFlags.port)
    fmt.Printf("Starting server on %s\n", addr)
    return srv.Start(addr)
}
```

## Frontend Embed (dashboard varsa)
```go
package web

import (
    "embed"
    "io/fs"
    "net/http"
)

//go:embed frontend/dist
var frontendFS embed.FS

func (s *Server) registerFrontend() {
    distFS, _ := fs.Sub(frontendFS, "frontend/dist")
    fileServer := http.FileServer(http.FS(distFS))
    
    s.echo.GET("/*", echo.WrapHandler(http.StripPrefix("/", fileServer)))
}
```

## Bağımlılıklar
```bash
go get github.com/labstack/echo/v4
go get modernc.org/sqlite  # veya pgx
```

## Doğrulama
- `curl http://localhost:{port}/api/health` → `{"success":true,"data":{"status":"ok"}}`
- Tüm endpoint'ler curl ile test edilebilir
- DB migration hatasız çalışır
- Graceful shutdown: Ctrl+C ile temiz kapanma
