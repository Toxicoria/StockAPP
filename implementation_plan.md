# Implementación de Onboarding y Sidecar en Tauri

Este plan detalla cómo transformar la app en un ejecutable autónomo que incluya el proxy (sidecar) y cómo implementar el flujo de primera configuración.

## ⚠️ User Review Required

- **Empaquetado del Sidecar**: Vamos a integrar el binario de Go (`cliente-sidecar`) dentro del ejecutable final de Tauri. Tauri se encargará de iniciarlo en segundo plano cuando la app arranque.
- **Persistencia de Configuración**: Usaremos `tauri-plugin-store` para guardar el Token de Tailscale localmente en la máquina del usuario para que la app lo recuerde entre reinicios.

## ❓ Open Questions

> [!WARNING]
> **1. ¿Registro de negocio nuevo o actualización?**
> Cuando el usuario ingrese "nombre del negocio, info de contacto, dueño", ¿esto significa que el sistema debe **registrar un nuevo negocio desde cero** en la base de datos central (y crearle un usuario admin), o simplemente está completando los datos de un negocio que ya existe? Si es un registro desde cero, necesitamos crear un endpoint nuevo en el backend (`POST /api/registro`).

> [!WARNING]
> **2. Campos faltantes en la Base de Datos**
> Actualmente la tabla `negocios` tiene `nombre_negocio`, `direccion` y `cuit`. ¿Necesitamos modificar `esquema.sql` para agregar columnas como `nombre_dueno`, `telefono_contacto` y `email_contacto`?

> [!WARNING]
> **3. Token de Tailscale (Auth Key)**
> Para que cada app tenga un token único, el usuario (o el instalador) tendrá que generar una "Auth Key" desechable o reutilizable desde el panel de Tailscale y pegarla en la pantalla de bienvenida. ¿Es este el flujo esperado?

## 🛠️ Proposed Changes

---

### 1. Backend (stock-api y stock-operations)

Si decidimos que el onboarding debe **crear** un nuevo negocio desde cero:
#### [NEW] `backend/services/stock-api/src/controllers/registroController.ts`
- Nuevo endpoint público `POST /api/registro` que reciba los datos iniciales, llame a Go para insertar el negocio y cree el primer usuario admin con su contraseña.
#### [MODIFY] `backend/services/stock-operations/pkg/api/router.go`
- Agregar ruta para registro o actualizar `PUT /api/negocio` para soportar los nuevos campos (dueño, contacto).
#### [MODIFY] `backend/services/stock-operations/model/esquema.sql`
- Agregar columnas como `nombre_dueno`, `telefono_contacto` si confirmamos que son necesarias.

---

### 2. Integración Tauri + Sidecar (Rust)

#### [MODIFY] `app-desktop/src-tauri/tauri.conf.json`
- Configurar `bundle.externalBin` para incluir el binario compilado de `cliente-sidecar`.
#### [MODIFY] `app-desktop/src-tauri/Cargo.toml`
- Agregar dependencias `tauri-plugin-store` (para guardar el token) y `tauri-plugin-shell` (para lanzar el sidecar).
#### [MODIFY] `app-desktop/src-tauri/src/main.rs`
- Escribir comandos Tauri que el frontend pueda llamar:
  - `guardar_token(token: String)`
  - `leer_token() -> Option<String>`
  - `iniciar_sidecar(token: String)`: Lanza el binario de Go con `TS_AUTHKEY=token` y el directorio de estado seteado en un path local del usuario (`AppData`).

---

### 3. Frontend (SvelteKit)

#### [NEW] `app-desktop/src/routes/setup/+page.svelte`
- Nueva pantalla de Onboarding.
- Paso 1: Pedir el **Token de Tailscale**.
  - Al ingresarlo, se llama a Tauri para guardarlo e iniciar el sidecar.
- Paso 2: Pantalla de conexión exitosa y formulario de **Datos del Negocio** / **Crear Usuario Admin**.
- Paso 3: Completar y redirigir al Hub (`/`).
#### [MODIFY] `app-desktop/src/routes/+layout.svelte`
- Lógica de redirección en el arranque: si Tauri dice que NO hay token guardado, redirigir a `/setup`. Si lo hay, iniciar el sidecar silenciosamente y proceder a intentar revivir la sesión como hace actualmente.

---

## 🧪 Verification Plan

### Manual Verification
1. **Compilación**: Compilar el sidecar para el SO objetivo (ej: Linux) y luego hacer `npm run tauri build` para verificar que el binario se empaqueta correctamente.
2. **First Run**: Limpiar el estado local de Tauri. Abrir la app compilada y verificar que arranca en la pantalla de "Setup".
3. **Flujo Sidecar**: Ingresar un Auth Key real de Tailscale. Verificar en los logs que el sidecar conecta a la tailnet y levanta en el puerto 9090 interno.
4. **Guardado**: Llenar los datos del negocio, cerrar la app y volver a abrirla. Verificar que el token se lee de memoria, el sidecar arranca solo y redirige al Home (sin pedir configuración de nuevo).
