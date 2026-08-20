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
- **stock-operations** (Go): CRUD de productos, stock, ventas, facturas, proveedores, resumen, usuarios. Valida sesiones por introspección contra stock-api.
- **Sidecar** (Go + tsnet): proxy local que enruta tráfico por túnel Tailscale. Solo para producción; en dev se apunta directo al gateway.
- **front-admin** (Futuro): panel web independiente para administración del sistema.

## Roles del sistema

- **`dueño`**: Cliente que usa la app. Control total de su negocio, gestión de productos, caja y usuarios empleados.
- **`cajero`**: Empleado del dueño. Acceso a caja (ventas), consulta de stock y cobro.

## Base de datos

- **PostgreSQL 16** (Alpine) en contenedor Podman
- Credenciales de desarrollo: `admin_dev` / `password_dev` / `stock_db`
- 8 tablas: `productos` (catálogo maestro global), `negocios`, `usuarios`, `stock_interno` (inventario del negocio), `ventas`, `detalles_venta`, `refresh_tokens`, `proveedores`
- Multi-tenant por `id_negocio` (un negocio nunca ve datos de otro)

## Usuarios de prueba (seed_dev.sql)

| Email | Contraseña | Rol |
|-------|-----------|-----|
| `duenio@dev.local` | `admin123` | dueño |
| `cajero@dev.local` | `cajero123` | cajero |

## Catálogo Maestro (21.800+ Productos Argentinos)

- Archivo de catálogo: `backend/services/stock-operations/productos_limpios.csv`
- Script importador: `cd backend/services/stock-operations && go run pkg/importar_db.go`

## Cómo levantar para desarrollo local (sin Tailscale)

```bash
# 1. Postgres
podman run -d --name db_stock -p 5432:5432 \
  -e POSTGRES_USER=admin_dev -e POSTGRES_PASSWORD=password_dev \
  -e POSTGRES_DB=stock_db docker.io/library/postgres:16-alpine

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

## Convenciones de código y Componentes

### Go (stock-operations)
- Router con `http.ServeMux` (Go 1.22+ patterns: `"GET /api/ruta"`)
- Middleware: `conCORS`, `conAuth`
- JSON helpers: `responderJSON()`, `responderError()`
- Claims: `claimsDe`, `negocioDe`, `rolDe`, `usuarioDe`
- Soporte para `cantidad_presentacion` y `unidad_medida` en productos e inventario.

### Svelte (app-desktop)
- Svelte 5 con runes (`$state`, `$derived`, `$effect`, `$props`)
- Design system propio en `src/lib/estilos/diseno.css` (paleta "patagónica": azul petróleo `#2e6e73`)
- Iconografía: **SVGs puros coloreados** (sin emojis textuales en componentes UI)
- Notificaciones: Sistema global de Toasts (`$lib/toast.svelte.js` + `Toast.svelte`)
- Carga de productos: Autocompletado desde catálogo maestro + Modo Carga Masiva con lector EAN
- Estado de sesión en `src/lib/sesion.svelte.js` (`esDueno()` para permisos)

## Seguridad

- Passwords: bcrypt
- Access token: JWT HS256, 15 min TTL
- Refresh token: 32 bytes random → SHA-256 en DB, 30 días, rotación en cada uso
- **Nunca subir**: `.env`, `tsnet-state/`, `process-compose.yaml`

## Podman como Docker

- Este proyecto usa **Podman** en lugar de Docker
- Usar `podman-compose` en vez de `docker compose`
