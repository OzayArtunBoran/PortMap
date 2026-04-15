# Skill: Python FastAPI

## Amaç
FastAPI + SQLAlchemy + Alembic tabanlı async API oluştur.

## Girdiler
- `project-spec.yml` → `api`, `database_schema`, `stack.python` bölümleri

## Kurallar
- Async SQLAlchemy 2.0: `create_async_engine`, `async_sessionmaker`
- `pydantic-settings` ile config yönetimi (`.env` dosyasından)
- Her model ayrı dosyada: `app/models/{model}.py`
- Her router ayrı dosyada: `app/api/{resource}.py`
- Pydantic schemas: `app/schemas/{resource}.py` (request/response ayrı class)
- Service katmanı: `app/services/{service}.py` (iş mantığı router'da değil, burada)
- Celery + Redis task queue (gerekiyorsa)
- Alembic migrations
- Response format standart: `{"success": bool, "data": ..., "error": ...}`

## Yapı

```
backend/
├── app/
│   ├── __init__.py
│   ├── main.py            # FastAPI app, middleware, lifespan
│   ├── config.py           # pydantic-settings
│   ├── database.py         # Async engine + session
│   ├── dependencies.py     # get_db, get_current_user
│   ├── models/             # SQLAlchemy ORM modelleri
│   │   ├── __init__.py
│   │   └── {model}.py
│   ├── api/                # FastAPI routers
│   │   ├── __init__.py
│   │   ├── router.py       # Ana router (tüm sub-router'ları birleştirir)
│   │   └── {resource}.py
│   ├── schemas/            # Pydantic request/response şemaları
│   │   ├── __init__.py
│   │   ├── common.py       # APIResponse, PaginatedResponse
│   │   └── {resource}.py
│   ├── services/           # Business logic
│   │   ├── __init__.py
│   │   └── {service}.py
│   ├── tasks/              # Celery tasks (opsiyonel)
│   │   ├── __init__.py
│   │   ├── celery_app.py
│   │   └── {task}.py
│   └── utils/
│       ├── __init__.py
│       └── {util}.py
├── alembic/
│   ├── env.py
│   └── versions/
├── alembic.ini
├── requirements.txt
├── Dockerfile
├── pytest.ini
└── tests/
    ├── __init__.py
    ├── conftest.py
    └── test_{module}.py
```

## Dosya İçerikleri

### app/config.py
```python
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    # App
    app_name: str = "{project.display_name}"
    debug: bool = False
    api_prefix: str = "{api.prefix}"

    # Database
    database_url: str = "postgresql+asyncpg://{name}:{name}@localhost:5432/{name}"

    # Redis (Celery varsa)
    redis_url: str = "redis://localhost:6379/0"

    # Auth (OAuth varsa)
    jwt_secret: str = "change-me-in-production"
    jwt_algorithm: str = "HS256"
    jwt_expiration_hours: int = 24

    # Notifications (bildirim sistemi varsa)
    slack_webhook_url: str = ""
    discord_webhook_url: str = ""
    telegram_bot_token: str = ""
    telegram_chat_id: str = ""
    smtp_host: str = ""
    smtp_port: int = 587
    smtp_username: str = ""
    smtp_password: str = ""
    smtp_from: str = ""
    notification_recipients: list[str] = []

    # Domain-specific settings
    # ...project-spec'ten gelen ayarlar

    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8")


settings = Settings()
```

### app/database.py
```python
from sqlalchemy.ext.asyncio import create_async_engine, async_sessionmaker, AsyncSession
from sqlalchemy.orm import DeclarativeBase

from app.config import settings

engine = create_async_engine(
    settings.database_url,
    echo=settings.debug,
    pool_pre_ping=True,
)

async_session = async_sessionmaker(
    engine,
    class_=AsyncSession,
    expire_on_commit=False,
)


class Base(DeclarativeBase):
    pass
```

### app/dependencies.py
```python
from typing import AsyncGenerator

from sqlalchemy.ext.asyncio import AsyncSession

from app.database import async_session


async def get_db() -> AsyncGenerator[AsyncSession, None]:
    async with async_session() as session:
        try:
            yield session
            await session.commit()
        except Exception:
            await session.rollback()
            raise


# Auth dependency (OAuth/JWT varsa)
# async def get_current_user(
#     token: str = Depends(oauth2_scheme),
#     db: AsyncSession = Depends(get_db),
# ) -> User:
#     ...
```

### app/main.py
```python
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.config import settings
from app.api.router import api_router


@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    print(f"{settings.app_name} starting...")
    yield
    # Shutdown
    print(f"{settings.app_name} shutting down...")


app = FastAPI(
    title=settings.app_name,
    lifespan=lifespan,
    docs_url=f"{settings.api_prefix}/docs",
    redoc_url=f"{settings.api_prefix}/redoc",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(api_router, prefix=settings.api_prefix)


@app.get("/health")
async def health():
    return {"success": True, "data": {"status": "ok"}}
```

### app/schemas/common.py
```python
from typing import Any, Optional

from pydantic import BaseModel


class APIResponse(BaseModel):
    success: bool
    data: Optional[Any] = None
    error: Optional[dict] = None


class PaginatedResponse(BaseModel):
    success: bool = True
    data: list[Any]
    total: int
    page: int
    per_page: int


def success_response(data: Any) -> dict:
    return {"success": True, "data": data}


def error_response(code: str, message: str) -> dict:
    return {"success": False, "error": {"code": code, "message": message}}
```

### app/api/router.py
```python
from fastapi import APIRouter

# from app.api import {resource1}, {resource2}

api_router = APIRouter()

# api_router.include_router({resource1}.router, prefix="/{resource1}s", tags=["{resource1}s"])
# api_router.include_router({resource2}.router, prefix="/{resource2}s", tags=["{resource2}s"])
```

### app/api/{resource}.py — Router pattern
```python
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.ext.asyncio import AsyncSession

from app.dependencies import get_db
from app.schemas.common import success_response, error_response
from app.services.{service} import {Service}

router = APIRouter()


@router.get("/")
async def list_{resource}s(
    db: AsyncSession = Depends(get_db),
    skip: int = 0,
    limit: int = 50,
):
    service = {Service}(db)
    items = await service.list(skip=skip, limit=limit)
    return success_response(items)


@router.get("/{id}")
async def get_{resource}(
    id: int,
    db: AsyncSession = Depends(get_db),
):
    service = {Service}(db)
    item = await service.get(id)
    if not item:
        raise HTTPException(status_code=404, detail="Not found")
    return success_response(item)


@router.post("/", status_code=201)
async def create_{resource}(
    data: {Create}Schema,
    db: AsyncSession = Depends(get_db),
):
    service = {Service}(db)
    item = await service.create(data)
    return success_response(item)
```

### app/services/{service}.py — Service pattern
```python
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.models.{model} import {Model}


class {Service}:
    def __init__(self, db: AsyncSession):
        self.db = db

    async def list(self, skip: int = 0, limit: int = 50) -> list[dict]:
        result = await self.db.execute(
            select({Model}).offset(skip).limit(limit)
        )
        items = result.scalars().all()
        return [item.to_dict() for item in items]

    async def get(self, id: int) -> dict | None:
        result = await self.db.execute(
            select({Model}).where({Model}.id == id)
        )
        item = result.scalar_one_or_none()
        return item.to_dict() if item else None

    async def create(self, data) -> dict:
        item = {Model}(**data.model_dump())
        self.db.add(item)
        await self.db.flush()
        await self.db.refresh(item)
        return item.to_dict()
```

### Celery setup (opsiyonel)
```python
# app/tasks/celery_app.py
from celery import Celery
from celery.schedules import crontab

from app.config import settings

celery_app = Celery(
    settings.app_name.lower(),
    broker=settings.redis_url,
    backend=settings.redis_url,
)

celery_app.conf.beat_schedule = {
    # Zamanlanmış görevler
    # "task-name": {
    #     "task": "app.tasks.{task}.{function}",
    #     "schedule": crontab(hour=9, minute=0),
    # },
}
```

### requirements.txt
```
fastapi>=0.110.0
uvicorn[standard]>=0.27.0
sqlalchemy[asyncio]>=2.0.0
asyncpg>=0.29.0
alembic>=1.13.0
pydantic-settings>=2.1.0
python-jose[cryptography]>=3.3.0
httpx>=0.27.0
celery>=5.3.0
redis>=5.0.0
ruff>=0.3.0
pytest>=8.0.0
pytest-asyncio>=0.23.0
```

## Bağımlılıklar
```bash
pip install -r backend/requirements.txt --break-system-packages
```

## Doğrulama
- `docker compose up` hatasız
- `curl http://localhost:8000/health` → `{"success":true,"data":{"status":"ok"}}`
- `curl http://localhost:8000/api/v1/docs` → Swagger UI
- `pytest` geçer
- `alembic upgrade head` hatasız
