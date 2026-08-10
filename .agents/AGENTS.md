# StockAPP — Reglas y Contexto del Proyecto

## Descripción

Sistema de gestión de inventario y ventas para negocios, con arquitectura cliente-servidor segura (Tailscale P2P). App de escritorio para Windows/Linux.

## Rama activa de desarrollo

**`feat/login-crud-y-fixes`** — contiene la versión actual con microservicios, auth, CRUD y UI.
`main` está desactualizada (solo tenía el monolito Go con un endpoint ping).

## Entorno de desarrollo

| Herramienta | Versión | Notas |
|-------------|---------|-------|
| **OS** | Fedora 44 (x86_64) | ThinkPad |
| **Go** | 1.26.5 | Backend stock-operations + sidecar |
| **Node.js** | 22.23.1 (npm 10.9.8) | Gateway stock-api + frontend SvelteKit |
| **Rust** | 1.97.1 (Cargo 1.97.1) | Shell nativo Tauri 2 |
| **Podman** | 5.8.4 | Reemplaza Docker (no hay Docker instalado) |
| **podman-compose** | 1.6.0 | Para levantar la infra |
| **Task** | 3.4.2 | Task runner (Taskfile.yml) |
| **WebKitGTK** | 2.52.5 | Dependencia de compilación de Tauri |

## Arquitectura (microservicios)

```
App Tauri+Svelte (:1420)
    → Sidecar Go (:9090) [solo producción, usa Tailscale]
        → stock-api Express/TS (:3000) [auth + proxy]
            → stock-operations Go (:8080) [lógica de negocio]
                → PostgreSQL (:5432)
```

- **stock-api** (TypeScript/Express 5): dueño de usuarios y sesiones (login/refresh/logout). Hace proxy a stock-operations para todo lo demás.
- **stock-operations** (Go): CRUD de productos, stock, ventas, facturas, proveedores, resumen. Valida sesiones por introspección contra stock-api.
- **Sidecar** (Go + tsnet): proxy local que enruta tráfico por túnel Tailscale. Solo para producción; en dev se apunta directo al gateway.

## Estructura de directorios

```
StockAPP/
├── app-desktop/              # Tauri 2 + SvelteKit (Svelte 5)
│   ├── src/                  # Frontend (rutas, componentes, lib)
│   └── src-tauri/            # Shell nativo Rust
├── backend/
│   ├── infra/                # docker-compose.yml + manifests K8s
│   └── services/
│       ├── stock-api/        # Gateway TS (Express 5, JWT, bcrypt)
│       └── stock-operations/ # API Go (net/http, lib/pq)
│           ├── model/        # esquema.sql + seed_dev.sql
│           └── pkg/api/      # Handlers, middleware, router
├── cliente-sidecar/          # Proxy Tailscale (Go + tsnet)
└── Taskfile.yml              # Task runner para infra y dev
```

## Base de datos

- **PostgreSQL 16** (Alpine) en contenedor
- Credenciales de desarrollo: `admin_dev` / `password_dev` / `stock_db`
- 8 tablas: `productos`, `negocios`, `usuarios`, `stock_interno`, `ventas`, `detalles_venta`, `refresh_tokens`, `proveedores`
- Multi-tenant por `id_negocio` (un negocio nunca ve datos de otro)

## Usuarios de prueba (seed_dev.sql)

| Email | Contraseña | Rol |
|-------|-----------|-----|
| `duenio@dev.local` | `admin123` | dueño |
| `cajero@dev.local` | `cajero123` | cajero |

## Cómo levantar para desarrollo local (sin Tailscale)

```bash
# 1. Postgres
podman run -d --name db_stock -p 5432:5432 \
  -e POSTGRES_USER=admin_dev -e POSTGRES_PASSWORD=password_dev \
  -e POSTGRES_DB=stock_db postgres:16-alpine

# 2. Esquema + seed
podman exec -i db_stock psql -U admin_dev -d stock_db < backend/services/stock-operations/model/esquema.sql
podman exec -i db_stock psql -U admin_dev -d stock_db < backend/services/stock-operations/model/seed_dev.sql

# 3. Backend Go (terminal 2)
cd backend/services/stock-operations && go run .

# 4. Gateway TS (terminal 3)
cd backend/services/stock-api && npm install && npm run dev

# 5. App desktop (terminal 4)
cd app-desktop && npm install
VITE_API_URL=http://localhost:3000 npm run tauri dev
```

O con Task: `task infra:up`, `task db:seed`, `task backend:run`, `task gateway:run`, `task app:dev`.

## Convenciones de código

### Go (stock-operations)
- Paquete `main` en raíz, lógica en `pkg/api/`, DB en `db/`, logger en `pkg/logger/`
- Router con `http.ServeMux` (Go 1.22+ patterns: `"GET /api/ruta"`)
- Middleware como funciones que envuelven `http.HandlerFunc` (`conCORS`, `conAuth`)
- JSON helpers: `responderJSON()`, `responderError()`
- Claims del token en contexto del request (`claimsDe`, `negocioDe`, `rolDe`, `usuarioDe`)
- Variables, funciones y comentarios en **español**

### TypeScript (stock-api)
- Express 5, ESM (`"type": "module"`)
- Estructura: `controllers/`, `services/`, `routes/`, `middleware/`, `errors/`, `config/`, `db/`
- Errores con clase `ApiError` (código HTTP + mensaje)
- Dev con `tsx watch`

### Svelte (app-desktop)
- Svelte 5 con runes (`$state`, `$derived`, `$effect`, `$props`)
- SvelteKit con `adapter-static` (SPA mode, SSR desactivado)
- Design system propio en `src/lib/estilos/diseno.css` (paleta "patagónica": azul petróleo `#2e6e73`)
- Componentes en `src/lib/componentes/`
- Estado de sesión en `src/lib/sesion.svelte.js` (access token en memoria, refresh en localStorage, `esDueno()` para permisos)
- API client en `src/lib/api.js` (auto-refresh ante 401)
- Variable `VITE_API_URL` configura el destino (`:9090` producción, `:3000` dev)

## Seguridad

- Passwords: bcrypt
- Access token: JWT HS256, 15 min TTL
- Refresh token: 32 bytes random → SHA-256 en DB, 30 días, rotación en cada uso
- Clave fiscal ARCA: AES-256-GCM en DB, nunca expuesta por API
- **Nunca subir**: `.env`, `tsnet-state/`, `process-compose.yaml`

## Tailscale

- El token de auth (`TS_AUTHKEY` / `TAILSCALE_AUTH_KEY`) se genera desde el panel de Tailscale
- No se guarda en archivos del proyecto — se pasa como variable de entorno
- Para desarrollo local no es necesario (se apunta directo al gateway en :3000)
- El sidecar se identifica como `cliente-stock-app` en la Tailnet
- El servidor se identifica como `stock-server-api`

## Podman como Docker

- Este proyecto usa **Podman** en lugar de Docker
- Usar `podman-compose` en vez de `docker compose`
- El Taskfile referencia `docker compose` en los tasks de infra — si se usan, reemplazar mentalmente por `podman-compose`
- Para K8s local el Taskfile usa k3d con socket de Podman
