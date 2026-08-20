# Plan de Implementación — Selector de Usuarios Recientes (Login estilo Windows/macOS)

Este documento detalla el diseño técnico y los cambios necesarios para transformar la pantalla de inicio de sesión (`/login`) en un selector visual de usuarios frecuentes en el equipo local.

---

## 🎯 Objetivo

Permitir que en equipos compartidos (p. ej. "Caja 1"), los empleados y dueños que ya hayan ingresado previamente en esta computadora puedan **seleccionar su perfil haciendo clic en un círculo con su avatar/nombre**, de modo que solo tengan que ingresar su contraseña sin escribir nuevamente el email.

---

## 💡 Flujo de usuario y Experiencia de Interfaz (UX)

### 1. Comportamiento en la pantalla `/login`:
- **Si NO hay usuarios recordados en este equipo** (primera vez o lista vacía):
  - Muestra el formulario tradicional de inicio de sesión (**Email** + **Contraseña**).

- **Si HAY usuarios recordados en este equipo**:
  - Muestra una cuadrícula / lista de **círculos de avatares** con las iniciales o color según rol.
  - Debajo de cada avatar se muestra el **Nombre del usuario**, **Email** y una etiqueta badge (`Dueño/a` o `Empleado/a`).
  - Al hacer clic en una tarjeta/círculo:
    - Se selecciona ese usuario.
    - Se oculta el campo de Email (o se muestra bloqueado/precompletado).
    - **El foco se coloca inmediatamente en el campo de Contraseña**.
  - Botón secundario: **"+ Ingresar con otra cuenta"** (para cambiar al formulario completo de email/password).
  - Opción de **"Olvidar cuenta en este equipo"** (botón de papelera o botón de quitar usuario).

---

## 🛠️ Cambios Propuestos

### 1. Persistencia de Usuarios Recientes (`app-desktop/src/lib/sesion.svelte.js`)

#### [MODIFY] [sesion.svelte.js](file:///home/Coria/Proyectos/StockAPP/app-desktop/src/lib/sesion.svelte.js)
- Agregar función helper `guardarUsuarioReciente(email, nombre, rol)`:
  - Lee la clave `usuarios_recientes` de `localStorage`.
  - Evita duplicados (si el email ya existía, lo actualiza y lo mueve al principio de la lista).
  - Mantiene hasta un máximo de 6 usuarios recientes en el equipo.
  - Guarda un objeto con `{ email, nombre, rol, ultimaSesion: Date.now() }`.
- Exportar función `obtenerUsuariosRecientes()` y `eliminarUsuarioReciente(email)`.
- Invocar `guardarUsuarioReciente` automáticamente dentro de `iniciarSesion(email, password)`.

---

### 2. Rediseño de la Pantalla de Login (`app-desktop/src/routes/login/+page.svelte`)

#### [MODIFY] [login/+page.svelte](file:///home/Coria/Proyectos/StockAPP/app-desktop/src/routes/login/+page.svelte)
- Leer los usuarios recientes guardados al montar la página.
- Agregar estado para el usuario seleccionado `usuarioSeleccionado` (`$state(null)`).
- **Vista A (Selector de Usuarios)**:
  - Círculos de avatares estilizados con las iniciales del nombre (ej: "D" para Dueño, "C" para Cajero).
  - Colores distintivos (Azul Petróleo para `dueño`, Verde Acua para `cajero`).
  - Al seleccionar un usuario, mostrar la tarjeta enfocada con el campo de contraseña y botón "Ingresar".
  - Botón "Cambiar de usuario" / "Usar otra cuenta".
- **Vista B (Formulario completo)**:
  - Formulario estándar de Email + Contraseña cuando se elige "+ Ingresar con otra cuenta" o no hay historial.

---

## 🧪 Plan de Verificación

1. **Primera ejecución (PC limpia)**:
   - Limpiar `localStorage.clear()`.
   - Abrir `/login` → Debe mostrar el formulario tradicional de Email + Contraseña.
2. **Primer Login**:
   - Iniciar sesión con `duenio@dev.local` / `admin123`.
   - Cerrar sesión.
3. **Selector Activo**:
   - Al volver a `/login`, debe aparecer el círculo/avatar de "Dueño Demo".
   - Al hacer clic en el círculo, el cursor salta directo a Contraseña.
4. **Segundo Login**:
   - Tocar "+ Ingresar con otra cuenta" e ingresar con `cajero@dev.local` / `cajero123`.
   - Cerrar sesión.
   - En `/login` ahora deben figurar ambos círculos (Dueño Demo y Cajero Demo).
5. **Olvidar Cuenta**:
   - Presionar la "X" u opción de olvidar en uno de los usuarios y verificar que se elimine correctamente del selector.
