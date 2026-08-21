# StockAPP 🛒📦

Sistema de gestión de inventario y ventas para negocios, diseñado con una arquitectura cliente-servidor moderna, rápida y segura.

---

## 🌟 Características Principales

- **Gestión de Stock e Inventario:** Control total de catálogo maestro, productos internos, precios y stock en tiempo real.
- **Punto de Venta (POS):** Interfaz fluida y optimizada para el registro de ventas rápidas, cobros y búsqueda por lector de código de barras (EAN).
- **Control de Acceso y Roles:** Permisos diferenciados para perfiles de **Dueño** (administración global) y **Cajero** (operación de ventas).
- **Multinegocio:** Arquitectura multi-tenant con aislamiento completo de datos por negocio.
- **Experiencia de Usuario Premium:** Aplicación de escritorio nativa, moderna y ágil con soporte para atajos de teclado y notificaciones en tiempo real.

---

## 🛠️ Tecnologías Utilizadas

- **Frontend Desktop:** [Tauri](https://tauri.app/) + [Svelte 5](https://svelte.dev/) (SvelteKit)
- **Servicios Backend:** [Go](https://go.dev/)
- **Gateway & Autenticación:** [Node.js](https://nodejs.org/) + Express (TypeScript)
- **Base de Datos:** [PostgreSQL](https://www.postgresql.org/)
- **Seguridad y Red:** Túnel P2P encriptado punto a punto.

---

## 🚀 Requisitos Previos

Para compilar o ejecutar el proyecto en modo desarrollo se requiere contar con:

- **Node.js** (v20+)
- **Go** (v1.22+)
- **Rust** & dependencias para Tauri
- **Podman** / **Docker**

---

## 💻 Ejecución en Desarrollo

1. **Clonar el repositorio:**
   ```bash
   git clone https://github.com/tu-usuario/StockAPP.git
   cd StockAPP
   ```

2. **Instalar dependencias y levantar el entorno local:**
   Seguir la guía interna de inicio para configurar la base de datos y los servicios.

3. **Iniciar la app desktop:**
   ```bash
   cd app-desktop
   npm install
   npm run tauri dev
   ```

---

## 🔒 Seguridad

- **Cifrado de datos:** Cifrado AES-256-GCM para datos sensibles en reposo.
- **Autenticación:** Gestión de sesiones mediante tokens JWT y tokens de refresco con rotación.
- **Comunicaciones:** Tráfico encriptado end-to-end entre cliente y servidor.
