# Diseño de Instalador Multiplataforma, Sistema de Actualizaciones y Vinculación (StockAPP)

Este documento resume las especificaciones técnicas, decisiones de arquitectura y flujos de trabajo acordados para el desarrollo del instalador, el wizard de primer inicio, el sistema de actualizaciones automáticas y la vinculación móvil mediante QR.

---

## 📋 Índice

1. [Arquitectura General e Instalador Multiplataforma](#1-arquitectura-general-e-instalador-multiplataforma)
2. [Wizard de Primer Inicio y Códigos de Activación](#2-wizard-de-primer-inicio-y-códigos-de-activación)
3. [Automatización con API de Tailscale](#3-automatización-con-api-de-tailscale)
4. [Sistema de Actualizaciones Automáticas (Tauri Updater)](#4-sistema-de-actualizaciones-automáticas-tauri-updater)
5. [Vinculación Futura con Dispositivos Móviles mediante QR](#5-vinculación-futura-con-dispositivos-móviles-mediante-qr)
6. [Plan de Acción para Ejecución](#6-plan-de-acción-para-ejecución)

---

## 1. Arquitectura General e Instalador Multiplataforma

### Objetivos
- Generar empaquetados e instaladores ejecutables e independientes para **Linux** (`.AppImage`, `.deb`) y **Windows** (`.exe` NSIS, `.msi`).
- Incluir el proxy Go embebido (**`cliente-sidecar`**) compilado para la plataforma correspondiente.

### Diagrama de Arquitectura
```
[ Instalador Tauri (Linux .AppImage/.deb | Windows .exe/.msi) ]
  ├── App Frontend (Svelte 5 + Design System Patagónico)
  ├── Rust Core (Gestión del subproceso sidecar y tauri-plugin-updater)
  ├── Almacenamiento Local (~/.config/stockapp/ o %APPDATA%\stockapp\)
  │     ├── config.json (Modo de conexión, URL, tokens)
  │     └── tsnet-state/ (Credenciales persistentes del túnel P2P)
  └── Binario Sidecar Embebido:
        ├── cliente-sidecar-x86_64-unknown-linux-gnu (Linux)
        └── cliente-sidecar-x86_64-pc-windows-msvc.exe (Windows)
              └── [ Túnel P2P Tailscale ] ──> Gateway stock-api (:3000) ──> stock-operations (:8080)
```

### Persistencia de Datos de Usuario
El sidecar mantendrá el estado del túnel en la carpeta de configuración del usuario según el sistema operativo:
- **Linux**: `~/.config/stockapp/`
- **Windows**: `%APPDATA%\stockapp\`

---

## 2. Wizard de Primer Inicio y Códigos de Activación

Para no exponer la clave larga de Tailscale (`tskey-auth-kXXXXXX-...`), se implementa un sistema de **Códigos Cortos de Activación** (ejemplo: `STOK-8472-9103`).

### Base de Datos (`codigos_activacion`)
```sql
CREATE TABLE codigos_activacion (
    id_codigo SERIAL PRIMARY KEY,
    codigo VARCHAR(50) UNIQUE NOT NULL,       -- ej: 'STOK-8472-9103'
    id_negocio INT REFERENCES negocios(id_negocio) ON DELETE CASCADE,
    ts_auth_key TEXT NOT NULL,                -- El token real de Tailscale
    destino_api VARCHAR(255) DEFAULT 'http://stock-server-api:3000',
    usado BOOLEAN DEFAULT FALSE,
    max_usos INT DEFAULT 1,                   -- Límite de computadoras por código
    usos_actuales INT DEFAULT 0,
    creado_en TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expira_en TIMESTAMP                       -- Fecha límite opcional
);
```

### Flujo del Wizard de 3 Pasos
1. **Paso 0 — Conexión y Activación de Licencia**:
   - El usuario ingresa el código corto (ej: `STOK-8472-9103`) o selecciona *Conexión Directa / Red Local* (IP fija `http://192.168.x.x:3000`).
   - La app consulta a `POST /api/activar-equipo` en el gateway público.
   - El servidor responde con la `ts_auth_key` y la dirección del búnker.
   - Rust inicia el `cliente-sidecar` automáticamente y confirma la conexión con un indicador en tiempo real.
2. **Paso 1 — Datos del Negocio**:
   - Registro de Nombre del negocio, CUIT, Dirección, Teléfono, Email.
3. **Paso 2 — Cuenta Administrador**:
   - Creación del usuario dueño inicial (`email_admin`, `password` mínimo 8 caracteres).
   - Redirección automática a la pantalla de Login (`/login`).

### Recuperación desde Pantalla de Login
- Se agrega un botón **"Configurar Servidor / Red"** en `/login` para permitir modificar el código o la IP en caso de que cambie la infraestructura de red.

---

## 3. Automatización con API de Tailscale

Para evitar copiar y pegar tokens manualmente desde el panel de administración de Tailscale:
- Se integra con la API oficial de Tailscale (`POST https://api.tailscale.com/api/v2/tailnet/{tailnet}/keys`) mediante credenciales **OAuth Client ID / Secret**.
- Cuando el administrador crea un nuevo negocio desde el panel admin (`front-admin`), el backend solicita automáticamente la Auth Key a Tailscale y genera el código corto `STOK-XXXX-XXXX`.

---

## 4. Sistema de Actualizaciones Automáticas (Tauri Updater)

El sistema de actualizaciones utiliza `@tauri-apps/plugin-updater` con firma criptográfica de binarios (`npx tauri signer generate`).

### Manifiesto JSON Servido por `stock-api` (`GET /api/app/update`)
```json
{
  "version": "1.2.0",
  "notes": "Mejoras de rendimiento en el módulo de ventas y caja.",
  "pub_date": "2026-08-12T15:00:00Z",
  "critica": false,
  "version_minima": "1.0.0",
  "platforms": {
    "linux-x86_64": {
      "signature": "dW50cnVzdGVkIGNvbW1lbnQ6...",
      "url": "https://actualizaciones.stockapp.com/v1.2.0/StockAPP.AppImage.tar.gz"
    },
    "windows-x86_64": {
      "signature": "dW50cnVzdGVkIGNvbW1lbnQ6...",
      "url": "https://actualizaciones.stockapp.com/v1.2.0/StockAPP_1.2.0_x64-setup.nsis.zip"
    }
  }
}
```

### Modos de Actualización
- **Actualización Opcional (`critica: false`)**:
  - Muestra una insignia/badge discreta en la barra de título superior: `🚀 Nueva versión v1.2.0 disponible`.
  - El usuario decide cuándo descargar e instalar sin interrumpir su trabajo.
- **Actualización Forzosa / Crítica (`critica: true` o `version_local < version_minima`)**:
  - Muestra un modal emergente bloqueante en la interfaz.
  - Impide operar en el sistema hasta presionar "Actualizar ahora".

---

## 5. Vinculación Futura con Dispositivos Móviles mediante QR

Permite conectar dispositivos móviles (Android / iOS) sin que el empleado escriba credenciales ni direcciones de red.

### Flujo de Emparejamiento por QR
1. El usuario entra a **Ajustes ➔ Vincular Dispositivo** en la App de Escritorio.
2. La app genera un **Código QR** con un Token de Emparejamiento Temporal (OTP de 5 minutos) o Auth Key reutilizable.
3. El celular abre la app móvil de StockAPP y escanea el QR con la cámara.
4. El celular almacena las credenciales en su almacenamiento encriptado seguro (`Keychain` en iOS / `EncryptedSharedPreferences` en Android).
5. El servidor registra el dispositivo (ej: *"Samsung Galaxy - Caja 2"*) y permite al dueño revocar accesos en cualquier momento.

---

## 6. Plan de Acción para Ejecución

### Tareas de Infraestructura y Backend
1. Crear tabla `codigos_activacion` en `esquema.sql` y `seed_dev.sql`.
2. Crear endpoint `POST /api/activar-equipo` y `GET /api/app/update` en `stock-api`.
3. Configurar script Go para compilar el sidecar en Linux y Windows.

### Tareas en la App de Escritorio (Tauri + Svelte)
1. Agregar módulo Rust para gestión de procesos sidecar (`iniciar_sidecar`, `detener_sidecar`, `cargar_config_conexion`).
2. Configurar `tauri-plugin-updater` y firma de claves.
3. Crear `conexion.svelte.js` y `actualizaciones.svelte.js` para manejo del estado reactivo con runes Svelte 5.
4. Implementar el Paso 0 en `setup/+page.svelte` (Canje del código `STOK-XXXX-XXXX`).
5. Diseñar la insignia de actualización opcional en `BarraTitulo.svelte` y el `ModalActualizacionForzosa.svelte`.

### Tareas de Compilación y Packaging (`Taskfile.yml`)
- `task build:sidecar:linux`: Compila binario Linux.
- `task build:sidecar:windows`: Compila binario Windows.
- `task build:app:linux`: Empaqueta `.AppImage` y `.deb`.
- `task build:app:windows`: Empaqueta ejecutable `.exe` / `.msi`.
