# Skill: Quality Checks

## Amaç
Her pipeline fazı sonunda çalıştırılacak kalite kontrolleri. Build, test, lint, CLAUDE.md güncelleme ve commit.

## Girdiler
- `project-spec.yml` → `quality` bölümü

---

## Pipeline Faz Sonu Prosedürü

Her faz tamamlandığında şu adımları sırayla uygula:

### 1. Build Kontrolü

```bash
# Go:
go build ./...

# Python:
python -c "from app.main import app; print('OK')"

# Frontend:
cd {frontend_dir} && npm run build

# Docker:
docker build -t {name}:test .
```

**❌ Build başarısız →** Dur, hatayı düzelt, tekrar dene. Sonraki adıma geçme.

### 2. Test Kontrolü

```bash
# Go:
go test ./... -v -count=1

# Python:
pytest -v

# Kısa versiyon (CI/pre-commit):
go test ./... -short -count=1
pytest -x -q
```

**❌ Test başarısız →** Dur, hatayı düzelt, tekrar dene. Sonraki adıma geçme.

### 3. Lint Kontrolü

```bash
# Go:
go vet ./...
# veya golangci-lint varsa:
golangci-lint run ./...

# Python:
ruff check .
# ruff format --check .
```

**⚠️ Lint uyarısı →** Düzelt (mümkünse). Kritik değilse not al ve devam et.

### 4. CLAUDE.md Güncelleme

Her faz sonunda CLAUDE.md'nin `Current Status` bölümünü güncelle:

```markdown
## Current Status
- [x] Phase 1: Scaffold
- [x] Phase 2: Core Features        ← tamamlanan faz
- [ ] Phase 3: Tests                 ← sonraki faz
- [ ] Phase 4: API/UI
- [ ] Phase 5: Notifications
- [ ] Phase 6: Docker + CI/CD
- [ ] Phase 7: README + GitHub
```

Ayrıca gerekiyorsa:
- Yeni struct/type eklendiyse `Key Types` bölümünü güncelle
- Yeni dosya eklendiyse `Project Structure` bölümünü güncelle
- Yeni komut eklendiyse `Build & Run` bölümünü güncelle

### 5. Commit

```bash
git add .
git commit -m "{prefix}: {açıklayıcı mesaj}"
```

**Conventional commit prefix'leri:**

| Prefix | Kullanım | Örnek |
|--------|----------|-------|
| `feat:` | Yeni özellik | `feat: add port scanning with proc parsing` |
| `fix:` | Hata düzeltme | `fix: handle empty config file gracefully` |
| `test:` | Test ekleme/değişiklik | `test: add edge case tests for allocator` |
| `chore:` | Bakım işleri | `chore: dockerize and add CI/CD workflows` |
| `docs:` | Dokümantasyon | `docs: add comprehensive README` |
| `refactor:` | Kod yeniden yapılandırma | `refactor: extract formatter interface` |

**Faz → Commit eşleştirmesi:**

| Faz | Prefix | Örnek |
|-----|--------|-------|
| Phase 1: Scaffold | `feat:` | `feat: initial project structure and CLAUDE.md` |
| Phase 2: Core | `feat:` | `feat: implement port scanner and conflict detector` |
| Phase 3: Tests | `test:` | `test: add comprehensive test suite` |
| Phase 4: API/UI | `feat:` | `feat: add dashboard frontend` |
| Phase 5: Notifications | `feat:` | `feat: add notification channels` |
| Phase 6: Docker+CI | `chore:` | `chore: dockerize and add CI/CD workflows` |
| Phase 7: README | `chore:` | `chore: add README, contributing docs, and license` |

---

## Hata Durumunda Davranış

### Build hatası
1. Hata mesajını oku
2. İlgili dosyayı bul ve düzelt
3. Tekrar build et
4. Başarılı olana kadar 1-3'ü tekrarla
5. Düzeltme başarılıysa devam et

### Test hatası
1. Hangi test başarısız olduğunu belirle
2. Test mi yanlış, kod mu yanlış — karar ver
3. İlgili tarafı düzelt
4. Tüm testleri tekrar çalıştır
5. Başarılı olana kadar tekrarla

### Import/dependency hatası
```bash
# Go:
go mod tidy

# Python:
pip install {eksik paket} --break-system-packages
```

---

## Temizlik Kontrolleri (Phase 7 sonunda)

Son fazda ekstra kontroller:

1. **Placeholder temizliği:** Kodda "TODO", "placeholder", "FIXME" kalmadığından emin ol
   ```bash
   grep -rn "TODO\|placeholder\|FIXME" --include="*.go" --include="*.py" --include="*.ts"
   ```

2. **Gereksiz yorum:** Açıklayıcı olmayan veya eski yorumları temizle

3. **Kullanılmayan import:** 
   ```bash
   # Go: go vet zaten yakalar
   # Python: ruff check --select F401
   ```

4. **Dosya izinleri:**
   ```bash
   # Script dosyaları çalıştırılabilir olmalı
   chmod +x ci/pre-commit-hook.sh
   chmod +x scripts/*.sh
   ```

5. **Sensitive data kontrolü:**
   ```bash
   # Hardcoded secret, password, token kalmadığından emin ol
   grep -rn "password\|secret\|token\|api_key" --include="*.go" --include="*.py" | grep -v "_test\." | grep -v "example\|placeholder\|changeme"
   ```

---

## Doğrulama Checklist

Her faz sonunda bu listeyi mental olarak kontrol et:

- [ ] Build başarılı
- [ ] Testler geçiyor
- [ ] Lint uyarısı yok (veya kabul edilebilir)
- [ ] CLAUDE.md güncel
- [ ] Commit yapıldı (conventional commit formatında)
- [ ] Gereksiz dosya (*.log, *.tmp) repo'ya eklenmedi
