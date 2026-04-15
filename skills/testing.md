# Skill: Testing

## Amaç
Kapsamlı test suite oluştur. Her public fonksiyon/method için testler, edge case'ler, table-driven/parametrize patternler.

## Girdiler
- `project-spec.yml` → `quality`, `features` bölümleri
- Mevcut kaynak kod (test edilecek modüller)

## Kurallar
- Her public fonksiyon/method için en az 1 test
- Edge case'ler: nil/None, boş input, hatalı input, sınır değerler
- Table-driven testler (Go) veya parametrize testler (Python)
- External servisler (HTTP, DB, dosya sistemi) mock'lanır
- Test isimleri açıklayıcı: `Test{Fonksiyon}_{Senaryo}_{BeklenenSonuç}`
- Test dosyaları, test edilen dosyanın yanında (Go) veya `tests/` dizininde (Python)

---

## Go Testleri

### Yapı
```
internal/{package}/
├── {file}.go
└── {file}_test.go    # Aynı dizinde, aynı package
```

### Tekil test pattern
```go
func TestFunctionName_ValidInput_ReturnsExpected(t *testing.T) {
    // Arrange
    input := "test-value"
    expected := "expected-result"

    // Act
    result, err := FunctionName(input)

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != expected {
        t.Errorf("got %q, want %q", result, expected)
    }
}
```

### Table-driven test pattern
```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected OutputType
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    validInput,
            expected: validOutput,
            wantErr:  false,
        },
        {
            name:     "empty input",
            input:    "",
            expected: zeroValue,
            wantErr:  true,
        },
        {
            name:     "nil input",
            input:    nil,
            expected: zeroValue,
            wantErr:  true,
        },
        {
            name:     "boundary value",
            input:    maxValue,
            expected: boundaryOutput,
            wantErr:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := FunctionName(tt.input)

            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if !tt.wantErr && result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### HTTP mock pattern (Go)
```go
func TestHTTPClient(t *testing.T) {
    // Mock server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // İstek doğrulama
        if r.Method != http.MethodPost {
            t.Errorf("expected POST, got %s", r.Method)
        }
        if r.URL.Path != "/expected/path" {
            t.Errorf("unexpected path: %s", r.URL.Path)
        }

        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status": "ok"}`))
    }))
    defer server.Close()

    // server.URL'yi kullanarak test et
    client := NewClient(server.URL)
    err := client.DoSomething()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

### Temp dosya/dizin pattern (Go)
```go
func TestFileOperation(t *testing.T) {
    // Geçici dizin
    tmpDir := t.TempDir()

    // Test dosyası oluştur
    testFile := filepath.Join(tmpDir, "test.yml")
    os.WriteFile(testFile, []byte("content"), 0644)

    // Test et
    result, err := ProcessFile(testFile)
    // ...
}
```

### Config test pattern (Go)
```go
func TestLoadConfig_ValidFile(t *testing.T) {
    tmpDir := t.TempDir()
    cfgPath := filepath.Join(tmpDir, ".config.yml")

    content := `
project: test-project
environments:
  dev:
    env_files:
      - .env.dev
`
    os.WriteFile(cfgPath, []byte(content), 0644)

    cfg, err := LoadConfig(cfgPath)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if cfg.Project != "test-project" {
        t.Errorf("project = %q, want %q", cfg.Project, "test-project")
    }
}

func TestLoadConfig_FileNotFound(t *testing.T) {
    _, err := LoadConfig("/nonexistent/path.yml")
    if err == nil {
        t.Error("expected error for nonexistent file")
    }
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
    tmpDir := t.TempDir()
    cfgPath := filepath.Join(tmpDir, "bad.yml")
    os.WriteFile(cfgPath, []byte("{{invalid yaml"), 0644)

    _, err := LoadConfig(cfgPath)
    if err == nil {
        t.Error("expected error for invalid YAML")
    }
}
```

---

## Python Testleri

### Yapı
```
backend/tests/
├── __init__.py
├── conftest.py           # Shared fixtures
├── test_{module}.py
└── ...
```

### conftest.py — Fixtures
```python
import pytest
import pytest_asyncio
from httpx import AsyncClient, ASGITransport
from sqlalchemy.ext.asyncio import create_async_engine, async_sessionmaker, AsyncSession

from app.main import app
from app.database import Base
from app.dependencies import get_db


# Test DB
TEST_DATABASE_URL = "sqlite+aiosqlite:///test.db"

engine = create_async_engine(TEST_DATABASE_URL)
TestSession = async_sessionmaker(engine, class_=AsyncSession, expire_on_commit=False)


@pytest_asyncio.fixture
async def db():
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    async with TestSession() as session:
        yield session
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.drop_all)


@pytest_asyncio.fixture
async def client(db):
    async def override_get_db():
        yield db

    app.dependency_overrides[get_db] = override_get_db

    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as c:
        yield c

    app.dependency_overrides.clear()
```

### Tekil test pattern
```python
class TestClassName:
    def test_method_valid_input_returns_expected(self):
        # Arrange
        input_data = "test-value"
        expected = "expected-result"

        # Act
        result = function_name(input_data)

        # Assert
        assert result == expected

    def test_method_empty_input_raises_error(self):
        with pytest.raises(ValueError):
            function_name("")
```

### Parametrize pattern
```python
@pytest.mark.parametrize("input_val,expected", [
    ("valid", "result"),
    ("", None),
    (None, None),
    ("boundary", "boundary-result"),
])
def test_function_parametrized(input_val, expected):
    result = function_name(input_val)
    assert result == expected
```

### FastAPI endpoint test pattern
```python
@pytest.mark.asyncio
async def test_health_endpoint(client):
    response = await client.get("/health")
    assert response.status_code == 200
    data = response.json()
    assert data["success"] is True
    assert data["data"]["status"] == "ok"


@pytest.mark.asyncio
async def test_create_resource(client):
    payload = {"name": "test", "value": 42}
    response = await client.post("/api/v1/resources/", json=payload)
    assert response.status_code == 201
    data = response.json()
    assert data["success"] is True
    assert data["data"]["name"] == "test"


@pytest.mark.asyncio
async def test_get_nonexistent_resource(client):
    response = await client.get("/api/v1/resources/999")
    assert response.status_code == 404
```

### Mock pattern (Python)
```python
from unittest.mock import AsyncMock, patch


@pytest.mark.asyncio
async def test_external_api_call():
    mock_response = AsyncMock()
    mock_response.json.return_value = {"result": "ok"}
    mock_response.status_code = 200

    with patch("app.services.external.httpx.AsyncClient.get", return_value=mock_response):
        result = await service.fetch_data()
        assert result == {"result": "ok"}
```

---

## Test Kategorileri Checklist

Her feature için şu kategorilerde test yaz:

1. **Happy path** — Normal çalışma senaryosu
2. **Edge cases** — Boş input, nil/None, sınır değerler
3. **Error handling** — Hatalı input, dosya bulunamadı, network hatası
4. **Validation** — Geçersiz format, eksik field, tip uyumsuzluğu
5. **Integration** — Modüller arası etkileşim (mock'suz, DB ile)

## Doğrulama
- Tüm testler geçer: `make test`
- Hata mesajları açıklayıcı (got X, want Y formatında)
- Mock'lar external bağımlılıkları tamamen izole eder
- Geçici dosyalar test sonrası temizlenir
