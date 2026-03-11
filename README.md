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

---
## 📦 1. Instalación de Requisitos y Paquetes

Es fundamental contar con los compiladores y runtimes instalados antes de intentar descargar las dependencias del proyecto.

### 🐧 En Linux (Fedora - Recomendado)
Ejecuta los siguientes comandos para instalar el stack completo:

1. **Motores de lenguaje:**
   `sudo dnf install golang nodejs rust cargo`
2. **Dependencias de compilación para Tauri:**
   `sudo dnf install webkit2gtk4.1-devel openssl-devel curl wget libappindicator-gtk3-devel librsvg2-devel`
3. **Docker:**
   `sudo dnf install dnf-plugins-core && sudo dnf config-manager --add-repo https://download.docker.com/linux/fedora/docker-ce.repo`
   `sudo dnf install docker-ce docker-ce-cli containerd.io docker-compose-plugin`
   `sudo systemctl enable --now docker`

### 🪟 En Windows (PowerShell como Administrador)
Se recomienda el uso de un gestor de paquetes como `choco` o `winget`, o instaladores oficiales:

1. **Motores de lenguaje:**
   `winget install GoLang.Go.1.25`
   `winget install OpenJS.NodeJS.LTS`
   `winget install Rustlang.Rustup`
2. **Herramientas de C++ (Vital para Tauri/Rust):**
   Descargar [Visual Studio Build Tools](https://visualstudio.microsoft.com/visual-cpp-build-tools/) y seleccionar la carga de trabajo "Desarrollo para el escritorio con C++".
3. **WebView2:**
   Instalar el [Evergreen Bootstrapper](https://developer.microsoft.com/en-us/microsoft-edge/webview2/) de Microsoft.
4. **Docker:**
   Instalar [Docker Desktop](https://www.docker.com/products/docker-desktop/) y asegurarse de que el motor WSL2 esté activo.

---

## 🛠 2. Descarga de Dependencias del Proyecto

Una vez instalados los motores de arriba, ejecuta lo siguiente en la raíz de `StockAPP`:

* **Módulos de Go (Backend y Sidecar):**
  `cd backend && go mod tidy && cd ../cliente-sidecar && go mod tidy`
* **Módulos de Node (Frontend):**
  `cd app-desktop && npm install`
* **Imágenes de Docker:**
  `cd infra && docker compose pull`
---
---
## 🛠 1. Configuración del Entorno y Dependencias

Antes de iniciar, ambos sistemas deben contar con: **Docker**, **Go (v1.25.7+)**, **Node.js (v20+)** y **Rust/Cargo**.

### 🐧 Linux (Fedora/Ubuntu/Debian)
1. **Infraestructura:** `cd infra && sudo docker compose pull`
2. **Backend & Sidecar:** `cd backend && go mod download` y `cd ../cliente-sidecar && go mod download`
3. **Frontend:** `cd app-desktop && npm install`
4. **Dependencias Tauri:** Ejecutar `sudo dnf install webkit2gtk4.1-devel openssl-devel curl wget` (o equivalente en `apt` para Debian/Ubuntu).

### 🪟 Windows (PowerShell)
1. **Infraestructura:** `cd infra; docker compose pull`
2. **Backend & Sidecar:** `cd backend; go mod download; cd ../cliente-sidecar; go mod download`
3. **Frontend:** `cd app-desktop; npm install`
4. **Dependencias:** Instalar **C++ Build Tools** (vía Visual Studio Installer) y el runtime de **WebView2**.

---

## 🚀 2. Comandos para Iniciar el Desarrollo

Se deben ejecutar los procesos en tres terminales independientes en el siguiente orden estricto:

### Paso 1: Infraestructura (Base de Datos y Red)
* **Linux:** `cd infra && sudo docker compose up -d`
* **Windows:** `cd infra; docker compose up -d`

### Paso 2: Sidecar (Proxy Seguro)
* **Linux:** `cd cliente-sidecar && export TS_AUTHKEY="tskey-auth-XXX" && go run main.go`
* **Windows:** `cd cliente-sidecar; $env:TS_AUTHKEY="tskey-auth-XXX"; go run main.go`

### Paso 3: App Desktop (Interfaz Svelte)
* **Linux/Windows:** `cd app-desktop && npm run tauri dev`

---

## 🛑 3. Comandos para Detener el Proyecto

### Detener Contenedores
* **Linux:** `cd infra && sudo docker compose stop`
* **Windows:** `cd infra; docker compose stop`

### Detener Procesos Locales
* En las terminales de **Sidecar** y **App**, presionar `Ctrl + C`.

---

## ⚠️ 4. Notas Esenciales y Consideraciones



* **Identidad de Red:** El Frontend debe comunicarse obligatoriamente con `http://localhost:9090` para que el tráfico sea tunelizado por el Sidecar.
* **Auth Keys:** Es necesario generar una `TS_AUTHKEY` válida y reusable desde el panel de administración de Tailscale para el entorno de desarrollo.
* **Persistencia:** Los datos de la base de datos se conservan en el volumen de Docker incluso si se detienen los contenedores con `stop`.
* **Seguridad:** No se deben commitear archivos `.env` o la carpeta `cliente-sidecar/tsnet-state/` para proteger las credenciales de la red privada.
* **Conflictos:** Verificar que el puerto **5432** (DB) y **9090** (Proxy) no estén ocupados por servicios locales previos.

> **Regla de Oro:** Para que el sistema funcione, el dispositivo de desarrollo debe tener acceso a la misma Tailnet (red privada) que los servicios de infraestructura.
