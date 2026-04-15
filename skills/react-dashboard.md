# Skill: React Dashboard

## Amaç
React + TypeScript + Tailwind + Recharts ile yönetim dashboard'u oluştur.

## Girdiler
- `project-spec.yml` → `frontend`, `api` bölümleri

## Kurallar
- Vite ile scaffold: `npm create vite@latest . -- --template react-ts`
- Tailwind CSS kullan
- Dark theme birincil
- Recharts ile grafikler (AreaChart, BarChart, LineChart, PieChart)
- lucide-react ile ikonlar
- react-router-dom ile routing
- API client: fetch wrapper, base URL `/api`
- Responsive: desktop öncelikli, mobile uyumlu
- Tek dosyada CSS (index.css veya App.css + Tailwind)
- State yönetimi: React hooks yeterli (Redux gereksiz)

## Kurulum
```bash
# Go projeleri:
mkdir -p ~/{project.name}/web/frontend
cd ~/{project.name}/web/frontend

# Python projeleri:
mkdir -p ~/{project.name}/frontend
cd ~/{project.name}/frontend

npm create vite@latest . -- --template react-ts
npm install tailwindcss @tailwindcss/vite react-router-dom recharts lucide-react
```

## Yapı

```
frontend/ (veya web/frontend/)
├── src/
│   ├── App.tsx              # Router setup
│   ├── main.tsx             # Entry point
│   ├── index.css            # Tailwind + tema CSS değişkenleri
│   ├── components/
│   │   ├── Layout.tsx       # Sidebar + topbar + main content
│   │   ├── SummaryCard.tsx  # Metrik özet kartı
│   │   ├── DataTable.tsx    # Reusable tablo bileşeni
│   │   ├── Badge.tsx        # Durum/severity badge
│   │   ├── EmptyState.tsx   # Veri yokken gösterilecek bileşen
│   │   ├── LoadingSpinner.tsx
│   │   └── ...
│   ├── pages/
│   │   ├── Dashboard.tsx    # Ana sayfa: summary + chart + tablo
│   │   └── ...
│   └── lib/
│       ├── api.ts           # API client (fetch wrapper)
│       └── types.ts         # TypeScript type tanımları
├── package.json
├── tsconfig.json
├── vite.config.ts
└── tailwind.config.ts
```

## Dosya İçerikleri

### src/index.css — Tema
```css
@import "tailwindcss";

:root {
  --bg: #0a0a0b;
  --bg-card: #141416;
  --bg-hover: #1c1c1e;
  --bg-input: #1c1c1e;
  --border: #27272a;
  --border-hover: #3f3f46;
  --text: #fafafa;
  --text-secondary: #a1a1aa;
  --text-muted: #52525b;
  --accent: #10b981;
  --accent-hover: #059669;
  --danger: #ef4444;
  --warning: #f59e0b;
  --info: #3b82f6;
  --success: #22c55e;
}

body {
  background-color: var(--bg);
  color: var(--text);
  font-family: 'JetBrains Mono', 'Geist Mono', ui-monospace, monospace;
  -webkit-font-smoothing: antialiased;
}

/* Scrollbar */
::-webkit-scrollbar { width: 6px; }
::-webkit-scrollbar-track { background: var(--bg); }
::-webkit-scrollbar-thumb { background: var(--border); border-radius: 3px; }
::-webkit-scrollbar-thumb:hover { background: var(--border-hover); }
```

### src/lib/api.ts — API client
```typescript
const BASE_URL = '/api';

interface APIResponse<T> {
  success: boolean;
  data: T;
  error?: { code: string; message: string };
}

async function fetchAPI<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    ...options,
  });

  const json: APIResponse<T> = await res.json();

  if (!json.success) {
    throw new Error(json.error?.message || `API Error: ${res.status}`);
  }

  return json.data;
}

// GET helper
export async function get<T>(path: string): Promise<T> {
  return fetchAPI<T>(path);
}

// POST helper
export async function post<T>(path: string, body: unknown): Promise<T> {
  return fetchAPI<T>(path, {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

// PUT helper
export async function put<T>(path: string, body: unknown): Promise<T> {
  return fetchAPI<T>(path, {
    method: 'PUT',
    body: JSON.stringify(body),
  });
}

// DELETE helper
export async function del<T>(path: string): Promise<T> {
  return fetchAPI<T>(path, { method: 'DELETE' });
}
```

### src/components/Layout.tsx
```tsx
import { ReactNode } from 'react';
import { NavLink } from 'react-router-dom';
import { /* ikonlar */ } from 'lucide-react';

interface LayoutProps {
  children: ReactNode;
}

export default function Layout({ children }: LayoutProps) {
  return (
    <div className="flex h-screen">
      {/* Sidebar */}
      <aside className="w-60 border-r flex flex-col"
        style={{ backgroundColor: 'var(--bg-card)', borderColor: 'var(--border)' }}>
        {/* Logo */}
        <div className="p-4 border-b" style={{ borderColor: 'var(--border)' }}>
          <h1 className="text-lg font-bold">{/* project.display_name */}</h1>
        </div>
        {/* Navigation */}
        <nav className="flex-1 p-3 space-y-1">
          {/* NavLink'ler — project-spec'ten gelen sayfalar */}
        </nav>
        {/* Version */}
        <div className="p-3 text-xs" style={{ color: 'var(--text-muted)' }}>
          v1.0.0
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 overflow-auto p-6">
        {children}
      </main>
    </div>
  );
}
```

### src/components/SummaryCard.tsx
```tsx
interface SummaryCardProps {
  title: string;
  value: number | string;
  change?: number;
  icon: ReactNode;
  color: string;
}

export default function SummaryCard({ title, value, change, icon, color }: SummaryCardProps) {
  return (
    <div className="rounded-lg p-4 border"
      style={{ backgroundColor: 'var(--bg-card)', borderColor: 'var(--border)' }}>
      <div className="flex items-center justify-between mb-2">
        <span className="text-sm" style={{ color: 'var(--text-secondary)' }}>{title}</span>
        <span style={{ color }}>{icon}</span>
      </div>
      <div className="text-2xl font-bold">{value}</div>
      {change !== undefined && (
        <div className="text-xs mt-1" style={{ color: change >= 0 ? 'var(--success)' : 'var(--danger)' }}>
          {change >= 0 ? '↑' : '↓'} {Math.abs(change)}%
        </div>
      )}
    </div>
  );
}
```

### src/pages/Dashboard.tsx — Örnek ana sayfa
```tsx
import { useEffect, useState } from 'react';
import { /* ikonlar */ } from 'lucide-react';
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';
import SummaryCard from '../components/SummaryCard';
import { get } from '../lib/api';

export default function Dashboard() {
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    get('/stats').then(setStats).finally(() => setLoading(false));
  }, []);

  if (loading) return <div>Loading...</div>;

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Dashboard</h1>

      {/* Summary cards */}
      <div className="grid grid-cols-4 gap-4 mb-6">
        {/* SummaryCard'lar — project-spec'e göre */}
      </div>

      {/* Trend chart */}
      <div className="rounded-lg border p-4 mb-6"
        style={{ backgroundColor: 'var(--bg-card)', borderColor: 'var(--border)' }}>
        <h2 className="text-lg font-semibold mb-4">Trend</h2>
        <ResponsiveContainer width="100%" height={300}>
          <AreaChart data={/* chart data */}>
            <XAxis dataKey="date" stroke="var(--text-muted)" />
            <YAxis stroke="var(--text-muted)" />
            <Tooltip
              contentStyle={{
                backgroundColor: 'var(--bg-card)',
                border: '1px solid var(--border)',
                borderRadius: '8px',
              }}
            />
            <Area type="monotone" dataKey="value" stroke="var(--accent)" fill="var(--accent)" fillOpacity={0.1} />
          </AreaChart>
        </ResponsiveContainer>
      </div>

      {/* Recent items table */}
      <div className="rounded-lg border"
        style={{ backgroundColor: 'var(--bg-card)', borderColor: 'var(--border)' }}>
        {/* DataTable — son kayıtlar */}
      </div>
    </div>
  );
}
```

### vite.config.ts
```typescript
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:{backend_port}',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
});
```

## Build & Embed

### Go projelerde embed
```go
//go:embed frontend/dist
var frontendFS embed.FS
```

### Python projelerde serve
```python
from fastapi.staticfiles import StaticFiles
app.mount("/", StaticFiles(directory="static", html=True), name="static")
```

## Doğrulama
- `npm run build` hatasız
- Tüm sayfalar render oluyor
- Dark theme doğru uygulanmış
- API çağrıları çalışıyor (backend ayaktayken)
- Responsive: 1024px ve üzerinde düzgün görünüm
