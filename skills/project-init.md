# Skill: Project Init

## Amaç
Projenin iskeletini oluştur: dizin yapısı, CLAUDE.md, Makefile, .gitignore, config dosyaları.

## Girdiler
- `project-spec.yml` → `project`, `stack`, `file_structure`, `quality` bölümleri

## Adımlar

### 1. Çalışma dizini oluştur
```bash
mkdir -p ~/{project.name}
cd ~/{project.name}
git init
```

### 2. project-spec.yml'i kopyala
```bash
cp /kaynak/project-spec.yml ~/{project.name}/project-spec.yml
```

### 3. Kullanılacak skill dosyalarını kopyala
project-spec.yml → `skills` listesindeki her skill dosyasını `skills/` dizinine kopyala:
```bash
mkdir -p ~/{project.name}/skills
cp /kaynak/skills/{skill-adı}.md ~/{project.name}/skills/
```

### 4. CLAUDE.md oluştur
Şablon: `templates/CLAUDE.md.template` dosyasını referans al.
project-spec.yml'den değerleri doldur:
- `{project.display_name}` → display_name
- `{project.description}` → description
- Stack bilgileri → stack bölümünden
- Dosya yapısı → `file_structure.tree`
- Build/test komutları → `quality` bölümünden
- Features listesi → `features` bölümünden (phase sırasıyla)
- Skills listesi → `skills` bölümünden

### 5. Makefile oluştur

**Go projeler:**
```makefile
BINARY_NAME={cli.binary_name}
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

.PHONY: build test lint clean install

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) .

test:
	go test ./... -v -count=1

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

tidy:
	go mod tidy

install: build
	cp bin/$(BINARY_NAME) $(GOPATH)/bin/

dev: build
	./bin/$(BINARY_NAME)
```

**Python projeler:**
```makefile
.PHONY: dev test lint clean migrate seed

dev:
	docker compose up

test:
	docker compose exec backend pytest -v

lint:
	docker compose exec backend ruff check .

migrate:
	docker compose exec backend alembic upgrade head

seed:
	docker compose exec backend python -m app.utils.seed

clean:
	docker compose down -v
```

### 6. .gitignore oluştur

**Go:**
```
bin/
*.db
*.sqlite
.env
dist/
node_modules/
vendor/
.DS_Store
*.log
```

**Python:**
```
__pycache__/
*.pyc
.env
*.egg-info/
dist/
node_modules/
.venv/
*.db
.DS_Store
*.log
```

### 7. Dile göre proje init

**Go:**
```bash
go mod init {stack.go.module}
```

**Python:**
```bash
mkdir -p backend/app
touch backend/app/__init__.py
# requirements.txt oluştur (stack bilgilerine göre)
```

### 8. Doğrulama
- `git status` — dosyalar mevcut
- Go: `go build ./...` hatasız (boş main.go bile olsa)
- Python: `pip install -r requirements.txt` hatasız

### 9. Commit
```bash
git add .
git commit -m "feat: initial project structure and CLAUDE.md"
```

## Çıktı
- Proje dizini hazır
- CLAUDE.md oluşturulmuş
- Makefile, .gitignore mevcut
- Git init + ilk commit yapılmış
- Sonraki faz başlayabilir
