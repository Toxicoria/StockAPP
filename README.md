# Sistema de Gestión de Stock 🔐

Desarrollo de control de inventario y ventas con arquitectura cliente-servidor de alta seguridad.

## 🛠️ Stack Técnico
- **Gateway:** TypeScript + Node.js/Express (usuarios, sesiones y puerta de entrada de la API).
- **Servicios:** Go (lógica de negocio interna: productos, stock, ventas, facturación).
- **DB:** PostgreSQL (Nube personal).
- **Desktop:** Tauri + SvelteKit (Svelte 5, ventana sin marco con el diseño "patagónico").
- **Red:** Tailscale (Túnel P2P encriptado vía `tsnet`).
- **IA:** Gemini API (Normalización de catálogo).

## 🗂️ Módulos Principales
- `/app-desktop`: Cliente Windows/Linux (Tauri + Svelte).
- `/backend/services/stock-api`: Gateway TS/Express — login/refresh/logout y proxy de todos los endpoints (:3000).
- `/backend/services/stock-operations`: Servicios de negocio en Go, internos detrás del gateway (:8080). Paquetes: `db/`, `pkg/api/`, `pkg/logger/`, `model/` (esquema y seed SQL).
- `/cliente-sidecar`: Proxy de red privada (Go + tsnet, :9090 → gateway :3000).
- `/infra`: Docker Compose y esquemas de DB.

### Cómo fluye una petición
```
App (Tauri) → sidecar :9090 → [tailnet] → stock-api :3000 → stock-operations :8080 → PostgreSQL
                                              ↑ ______________ /internal/session ↲
```
`stock-api` resuelve login/refresh/logout contra las tablas `usuarios`/`refresh_tokens`
y emite los JWT; para el resto valida el token y reenvía a `stock-operations`.
`stock-operations` no conoce el secreto JWT: valida cada sesión **consumiendo
`/internal/session` de stock-api** (introspección, con cache de 60 s). Ninguno de
los dos servicios de negocio se expone directamente: la única puerta es stock-api.

### Kubernetes local (kind)
`stock-api` y `stock-operations` corren como pods en el cluster kind:
```bash
task local:up       # crea el cluster (k3d + podman)
task local:deploy   # construye imágenes, las importa al cluster y aplica infra/k8s/
task local:seed     # seed dentro del pod de Postgres
task local:probar   # port-forward a stock-api :3000 para usarlo desde la app
```

---

## 📦 1. Instalación de Requisitos y Paquetes

Es fundamental contar con los compiladores y runtimes instalados antes de intentar descargar las dependencias del proyecto.

### 🐧 En Linux (Fedora - Recomendado)
1. **Motores de lenguaje:** `sudo dnf install golang nodejs rust cargo`
2. **Dependencias de compilación para Tauri:** `sudo dnf install webkit2gtk4.1-devel openssl-devel curl wget libappindicator-gtk3-devel librsvg2-devel`
3. **Docker:** `sudo dnf install dnf-plugins-core && sudo dnf config-manager --add-repo https://download.docker.com/linux/fedora/docker-ce.repo`
   `sudo dnf install docker-ce docker-ce-cli containerd.io docker-compose-plugin`
   `sudo systemctl enable --now docker`
   *Nota: Se recomienda ejecutar `sudo usermod -aG docker $USER` y reiniciar la sesión para usar Docker sin sudo.*

### 🪟 En Windows (PowerShell como Administrador)
1. **Motores de lenguaje:** `winget install GoLang.Go.1.25` / `winget install OpenJS.NodeJS.LTS` / `winget install Rustlang.Rustup`
2. **Herramientas de C++:** Instalar [Visual Studio Build Tools](https://visualstudio.microsoft.com/visual-cpp-build-tools/) (Seleccionar la carga de trabajo: "Desarrollo para el escritorio con C++").
3. **WebView2:** Instalar el [Evergreen Bootstrapper](https://developer.microsoft.com/en-us/microsoft-edge/webview2/).
4. **Docker:** Instalar [Docker Desktop](https://www.docker.com/products/docker-desktop/) (Asegurar motor WSL2 activo).

---

## 🛠 2. Descarga de Dependencias del Proyecto

Una vez instalados los motores de arriba, ejecuta lo siguiente en la raíz de `StockAPP`:

* **Módulos de Go:** `cd backend/services/stock-operations && go mod tidy && cd ../cliente-sidecar && go mod tidy`
* **Módulos de Node:** `cd app-desktop && npm install && cd ../backend/services/stock-api && npm install`
* **Imágenes de Docker:** `cd infra && docker compose pull`

---

## 🚀 3. Comandos para Iniciar el Desarrollo

Con [Task](https://taskfile.dev) instalado (desarrollo 100 % local, sin túnel):

```bash
task infra:up        # Postgres (+ API en contenedores)
task db:esquema      # aplica backend/services/stock-operations/model/esquema.sql (idempotente, sirve sobre bases vivas)
task db:seed         # admin@dev.local/admin123 y cajero@dev.local/cajero123
task backend:run     # servicio Go interno en :8080
task gateway:run     # gateway TS/Express en :3000
task app:dev         # app Tauri (usa VITE_API_URL=http://localhost:3000 de .env.development)
```

Para el modo remoto (app hablando con el servidor real por Tailscale), en vez de
`backend:run`/`gateway:run` locales se usa el sidecar:

* **Linux:** `cd cliente-sidecar && export TS_AUTHKEY="tskey-auth-XXX" && go run main.go`
* **Windows:** `cd cliente-sidecar; $env:TS_AUTHKEY="tskey-auth-XXX"; go run main.go`

y la app apunta a `http://localhost:9090` (el default cuando no hay `.env.development`).

---

## 🛑 4. Detener el Proyecto
* **Contenedores:** `cd infra && docker compose stop`
* **Procesos Locales:** Presionar `Ctrl + C` en las terminales de Sidecar y App.

---

## ⚠️ 5. Notas Esenciales y Consideraciones

* **Identidad de Red:** En remoto el Frontend habla con `http://localhost:9090` (Sidecar); en desarrollo local, con `http://localhost:3000` (gateway).
* **Auth Keys:** Es necesario generar una `TS_AUTHKEY` reusable desde el panel de Tailscale para desarrollo.
* **Persistencia:** Los datos se conservan en el volumen de Docker incluso al detener los contenedores con `stop`.
* **Seguridad:** No subir al repositorio archivos `.env` ni la carpeta `cliente-sidecar/tsnet-state/`.
* **Conflictos:** Verificar que los puertos **5432** (DB), **3000** (gateway), **8080** (Go) y **9090** (Proxy) estén libres.

> **Regla de Oro:** El dispositivo de desarrollo debe estar autenticado en la misma Tailnet que la infraestructura para garantizar la conectividad del túnel.

---

## 🧾 6. Facturación y ARCA

Hoy la facturación es **numeración local** (comprobante interno tipo C para
monotributo): al tocar «Facturar» en Ventas, el negocio asigna el próximo
número (`C 0001-00000214`) de forma atómica. **No** se emite comprobante fiscal.

El esquema ya quedó **ARCA-ready** para integrar WSFEv1 más adelante sin migrar
datos ni tocar la UI: `ventas` tiene `tipo_comprobante` (11 = Factura C),
`nro_comprobante`, `cae` y `cae_vencimiento` (los dos últimos quedan NULL hasta
integrar el web service). El camino futuro es:

1. Generar el certificado con el CUIT en el portal de ARCA y darlo de alta para WSAA.
2. Implementar el cliente WSAA (login CMS) y WSFEv1 (`FECAESolicitar`) primero en homologación.
3. Reemplazar el asignador local: el número y el CAE pasan a venir de ARCA
   (ojo con el campo `CondicionIVAReceptorId`, obligatorio desde el manual v4.x).

La **clave fiscal** se carga desde la app (Inicio → Configuración, solo admin)
y se guarda cifrada (`negocios.clave_fiscal_cifrada`, AES-256-GCM) — la API
nunca la devuelve, solo informa si hay una configurada. El cliente WSAA del
punto 2 va a ser quien la descifre internamente (`pkg/api/secreto.go`) para
autenticarse contra ARCA.

## 🔒 7. Seguridad

- Secretos en reposo: la clave fiscal se cifra con AES-256-GCM antes de
  guardarse; la clave de cifrado sale de `CONFIG_SECRET_KEY` (mismo patrón
  que `JWT_SECRET` — con default de desarrollo, **obligatoria en producción**).
- Escapado: Svelte escapa automáticamente todo lo que se interpola en el
  markup (`{expresión}`); la app no usa `{@html}` en ningún lado, así que un
  nombre de producto o proveedor con caracteres raros no puede inyectar HTML.
- Headers HTTP: tanto `stock-api` como `stock-operations` responden
  `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY` y
  `Referrer-Policy: no-referrer` en cada respuesta.
- Todas las consultas SQL van parametrizadas (`$1`, `$2`, ...) — nunca se arma
  una query concatenando texto del usuario.