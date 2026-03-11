# Sistema de Gestión de Stock 🔐

Desarrollo de control de inventario y ventas con arquitectura cliente-servidor de alta seguridad.

## 🛠️ Stack Técnico
- **Backend:** Go (API centralizada en Docker).
- **DB:** PostgreSQL (Nube personal) + SQLite (Local App).
- **Desktop:** Tauri + SvelteKit + TypeScript (Optimizado para bajos recursos).
- **Red:** Tailscale (Túnel P2P encriptado vía `tsnet`).
- **IA:** Gemini API (Normalización de catálogo).

## 🗂️ Módulos Principales
- `/app-desktop`: Cliente Windows/Linux (Tauri + Svelte).
- `/backend`: Lógica de negocio y API (Go).
- `/cliente-sidecar`: Proxy de red privada (Go + tsnet).
- `/infra`: Docker Compose y esquemas de DB.

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

* **Módulos de Go:** `cd backend && go mod tidy && cd ../cliente-sidecar && go mod tidy`
* **Módulos de Node:** `cd app-desktop && npm install`
* **Imágenes de Docker:** `cd infra && docker compose pull`

---

## 🚀 3. Comandos para Iniciar el Desarrollo (3 Terminales)

Se deben ejecutar los procesos en tres terminales independientes en el siguiente orden estricto:

### Paso 1: Infraestructura (Base de Datos)
* **Linux:** `cd infra && docker compose up -d`
* **Windows:** `cd infra; docker compose up -d`

### Paso 2: Sidecar (Proxy Seguro)
* **Linux:** `cd cliente-sidecar && export TS_AUTHKEY="tskey-auth-XXX" && go run main.go`
* **Windows:** `cd cliente-sidecar; $env:TS_AUTHKEY="tskey-auth-XXX"; go run main.go`

### Paso 3: App Desktop (Frontend)
* **Ambos:** `cd app-desktop && npm run tauri dev`

---

## 🛑 4. Detener el Proyecto
* **Contenedores:** `cd infra && docker compose stop`
* **Procesos Locales:** Presionar `Ctrl + C` en las terminales de Sidecar y App.

---

## ⚠️ 5. Notas Esenciales y Consideraciones

* **Identidad de Red:** El Frontend debe comunicarse siempre con `http://localhost:9090` (Sidecar).
* **Auth Keys:** Es necesario generar una `TS_AUTHKEY` reusable desde el panel de Tailscale para desarrollo.
* **Persistencia:** Los datos se conservan en el volumen de Docker incluso al detener los contenedores con `stop`.
* **Seguridad:** No subir al repositorio archivos `.env` ni la carpeta `cliente-sidecar/tsnet-state/`.
* **Conflictos:** Verificar que los puertos **5432** (DB) y **9090** (Proxy) estén libres.

> **Regla de Oro:** El dispositivo de desarrollo debe estar autenticado en la misma Tailnet que la infraestructura para garantizar la conectividad del túnel.