# Sistema de Gestión de Stock 🔐

Desarrollo de control de inventario y ventas con arquitectura cliente-servidor.

## 🛠️ Stack Técnico
- **Backend:** Go (API centralizada en Docker).
- **DB:** PostgreSQL (Nube personal) + SQLite (Local App).
- **Desktop:** Tauri + Svelte (Optimizado para bajos recursos).
- **Red:** Tailscale (Túnel P2P encriptado).
- **IA:** Gemini API (Normalización de catálogo).

## 🗂️ Módulos Principales
- `/app-desktop`: Cliente Windows 10 x64.
- `/backend-go`: Lógica de negocio y sincronización.
- `/infra`: Docker Compose y esquemas de DB.

## 🔒 Notas de Seguridad
- Acceso restringido vía MFA (App/Email).
- Comunicación cifrada de extremo a extremo.
- Base de datos multi-tenant aislada.
