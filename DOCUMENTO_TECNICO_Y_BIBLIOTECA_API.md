# StockAPP — Especificación Técnica Integral, Arquitectura de Comunicaciones y Biblioteca de APIs

**Versión del Documento:** 1.1.0 (Producción)  
**Clasificación:** Documentación Técnica de Arquitectura e Integración  
**Fecha de Publicación:** Septiembre 2026  

---

## 📑 Contenido del Documento

1. [Visión General y Alcance del Sistema](#1-visión-general-y-alcance-del-sistema)
2. [Arquitectura de Microservicios y Topología de Red](#2-arquitectura-de-microservicios-y-topología-de-red)
3. [Modelo de Comunicaciones Inter-Servicios](#3-modelo-de-comunicaciones-inter-servicios)
   - 3.1 App Desktop hacia Sidecar Local
   - 3.2 Sidecar hacia Gateway Remoto (Túnel P2P Tailscale)
   - 3.3 Gateway (stock-api) hacia Backend de Negocio (stock-operations)
   - 3.4 Backend de Negocio hacia Gateway (Introspección de Sesiones)
   - 3.5 Microservicios hacia PostgreSQL
   - 3.6 Front-Admin hacia Gateway
4. [Modelo de Datos y Base de Datos (PostgreSQL 16)](#4-modelo-de-datos-y-base-de-datos-postgresql-16)
   - 4.1 Diagrama Entidad-Relación (ERD)
   - 4.2 Diccionario Exhaustivo de Tablas
   - 4.3 Catálogo Maestro Nacional (21.800+ Productos)
5. [Biblioteca Exhaustiva de APIs (API Catalog & Reference)](#5-biblioteca-exhaustiva-de-apis-api-catalog--reference)
   - 5.1 Módulo 1: Salud y Verificación
   - 5.2 Módulo 2: Autenticación de Clientes y Terminales (stock-api)
   - 5.3 Módulo 3: Autenticación SuperAdmin con 2FA TOTP (stock-api)
   - 5.4 Módulo 4: Introspección Interna de Sesiones (stock-api)
   - 5.5 Módulo 5: Productos y Catálogo Maestro (stock-operations)
   - 5.6 Módulo 6: Stock e Inventario por Comercio (stock-operations)
   - 5.7 Módulo 7: Ventas, Tickets y Facturación (stock-operations)
   - 5.8 Módulo 8: Proveedores y Aumentos Masivos (stock-operations)
   - 5.9 Módulo 9: Resumen de Caja y Configuración de Negocio (stock-operations)
   - 5.10 Módulo 10: Usuarios y Empleados (stock-operations)
   - 5.11 Módulo 11: Administración Global de Inquilinos (stock-operations)
6. [Mecanismos de Seguridad y Políticas Criptográficas](#6-mecanismos-de-seguridad-y-políticas-criptográficas)
7. [Guía de Despliegue y Configuración de Entorno](#7-guía-de-despliegue-y-configuración-de-entorno)

---

## 1. Visión General y Alcance del Sistema

**StockAPP** es una plataforma integral de gestión de inventario, punto de venta (POS) y facturación para comercios minoristas (kioscos, minimercados, almacenes). Su diseño desacoplado resuelve las restricciones de hardware y conectividad que suelen tener los pequeños comercios en Argentina:

- **Rendimiento Nativo en Mostrador:** Desarrollado con Tauri v2 y Svelte 5, garantizando consumos mínimos de memoria RAM (< 90 MB) y respuesta instantánea al escaneo de código de barras.
- **Catálogo Argentino Precargado:** Incluye 21.800+ artículos con códigos EAN-13, marcas y descripciones oficiales que eliminan el trabajo de tipiado manual.
- **Aislamiento Multi-Tenant Estricto:** Cada comercio opera con datos lógicamente aislados mediante la clave foránea `id_negocio`.
- **Cero Exposición de Puertos en el Comercio:** El cliente se conecta mediante un túnel P2P seguro (Tailscale / WireGuard), evitando configurar reglas NAT o abrir puertos en los routers locales (MikroTik, etc.).

---

## 2. Arquitectura de Microservicios y Topología de Red

El sistema está dividido en capas especializadas para separar la interfaz de usuario, la autenticación, la lógica transaccional y la base de datos:

```mermaid
graph TD
    subgraph CLIENTE_LOCAL [Dispositivo del Cliente - Linux / Windows]
        UI[App Desktop Tauri v2 + Svelte 5 :1420]
        Sidecar[Sidecar Go tsnet :9090]
        UI -->|HTTP Local| Sidecar
    end

    subgraph RED_TAILSCALE [Red Cifrada P2P WireGuard / Tailnet]
        Sidecar -.->|Túnel P2P Cifrado| Gateway
        AdminBrowser[Navegador Admin / Celular] -.->|Túnel P2P Cifrado| FrontAdmin
    end

    subgraph SERVIDOR_PROD [Servidor Remoto Linux - Red Bridge stockapp_network]
        FrontAdmin[front-admin Nginx :5173]
        Gateway[stock-api Express/TS :3000]
        Operations[stock-operations Go :8080]
        DB[(PostgreSQL 16 :5432)]

        FrontAdmin -->|Proxy inverso /api/| Gateway
        Gateway -->|Reverse Proxy HTTP con JWT| Operations
        Operations -->|Introspección /internal/session| Gateway
        Gateway -->|Consultas Auth / Sesiones| DB
        Operations -->|Consultas Negocio / Ventas / Stock| DB
    end

    style CLIENTE_LOCAL fill:#f0fdfa,stroke:#0f766e,stroke-width:2px
    style RED_TAILSCALE fill:#f0f9ff,stroke:#0284c7,stroke-width:2px,stroke-dasharray: 5 5
    style SERVIDOR_PROD fill:#f8fafc,stroke:#475569,stroke-width:2px
    style DB fill:#fdf2f8,stroke:#be185d,stroke-width:2px
    style Gateway fill:#eff6ff,stroke:#2563eb,stroke-width:2px
    style Operations fill:#fefce8,stroke:#ca8a04,stroke-width:2px
```

---

## 3. Modelo de Comunicaciones Inter-Servicios

### 3.1 App Desktop hacia Sidecar Local
- **Protocolo:** HTTP/1.1 sin TLS (comunicación interna sobre la interfaz loopback `127.0.0.1`).
- **Puerto:** `9090`.
- **Propósito:** La interfaz de Svelte enruta todas sus peticiones API (`/api/*`) a través de este proxy local. En entornos de desarrollo apunta directo al gateway (`:3000`); en producción apunta al sidecar.

### 3.2 Sidecar hacia Gateway Remoto (Túnel P2P Tailscale)
- **Protocolo:** WireGuard Mesh (UDP puerto `41641` o túnel DERP TCP `443`).
- **Cifrado:** ChaCha20-Poly1305 / Noise Protocol.
- **Propósito:** El sidecar utiliza la librería `tsnet` de Go para actuar como un nodo de red efímero autenticado con una clave de un solo uso. Enruta las peticiones de forma transparente hacia la IP de Tailscale del servidor (`100.x.y.z:3000`).

### 3.3 Gateway (stock-api) hacia Backend de Negocio (stock-operations)
- **Protocolo:** HTTP/1.1 interno en la red de contenedores de Podman/Docker (`stockapp_network`).
- **Dirección Upstream:** `http://api_go:8080`.
- **Flujo:** 
  1. `stock-api` recibe la petición en el puerto `3000`.
  2. Valida la presencia y firma del header `Authorization: Bearer <token>`.
  3. Si la ruta pertenece a `/api/admin/*`, verifica estrictamente que el claim `rol` sea `'superadmin'`.
  4. Hace proxy reverso preservando el método HTTP, cuerpo de la petición y los headers de autorización.

### 3.4 Backend de Negocio hacia Gateway (Introspección de Sesiones)
- **Protocolo:** HTTP/1.1 interno (`http://api_ts:3000/internal/session`).
- **Propósito:** `stock-operations` no comparte el secreto JWT con `stock-api` (dueño de sesiones). Para validar el token, `stock-operations` consulta este endpoint interno.
- **Optimización por Cache:** Implementa una `sync.Map` con TTL de 60 segundos por token para evitar la penalización de un roundtrip HTTP en cada petición.

### 3.5 Microservicios hacia PostgreSQL
- **Protocolo:** Protocolo binario de PostgreSQL (puerto `5432` privado).
- **Pool de Conexiones:**
  - `stock-api`: Biblioteca `pg` con pool configurado.
  - `stock-operations`: Driver oficial `github.com/lib/pq` con `sql.DB` y reintentos automáticos de conexión (hasta 15 intentos cada 2 segundos).

### 3.6 Front-Admin hacia Gateway
- **Protocolo:** HTTP/1.1 servido por Nginx en el puerto `5173` (o `80` dentro del contenedor).
- **Enrutamiento:** Las peticiones de frontend a `/api/*` son capturadas por la regla `location /api/ { proxy_pass http://api_ts:3000/api/; }`, permitiendo que la SPA funcione con rutas relativas (`API_BASE = ''`) en cualquier dominio o IP sin requerir configuración de CORS.

---

## 4. Modelo de Datos y Base de Datos (PostgreSQL 16)

### 4.1 Diagrama Entidad-Relación (ERD)

```mermaid
erDiagram
    super_admins {
        int id_admin PK
        string usuario UK
        string password_hash
        string nombre
        string email
        string totp_secret
        boolean totp_activado
        timestamp fecha_alta
    }

    negocios ||--o{ usuarios : "emplea"
    negocios ||--o{ stock_interno : "inventaria"
    negocios ||--o{ ventas : "factura"
    negocios ||--o{ proveedores : "registra"
    negocios ||--o{ dispositivos_clientes : "vincula"

    negocios {
        int id_negocio PK
        string nombre_negocio
        string direccion
        string cuit
        string telefono
        string email_negocio
        string nombre_dueno
        string punto_venta
        int proximo_numero_factura
        timestamp fecha_alta
    }

    usuarios ||--o{ refresh_tokens : "genera"
    usuarios ||--o{ ventas : "opera"

    usuarios {
        int id_usuario PK
        int id_negocio FK
        string usuario
        string nombre
        string email
        string password_hash
        string rol
        boolean activo
    }

    productos ||--o{ stock_interno : "referenciado_en"
    productos {
        string id_producto PK "EAN o UUID"
        string productos_ean "EAN-13"
        string productos_descripcion
        string productos_cantidad_presentacion
        string productos_unidad_medida_presentacion
        string productos_marca
    }

    stock_interno {
        int id_stock PK
        int id_negocio FK
        string id_producto FK
        numeric cantidad_disponible
        numeric precio_venta
        numeric precio_costo
        numeric stock_minimo
        int id_proveedor FK
    }

    proveedores ||--o{ stock_interno : "abastece"
    proveedores {
        int id_proveedor PK
        int id_negocio FK
        string nombre
        string telefono
        string email
    }

    ventas ||--|{ detalles_venta : "contiene"
    ventas {
        int id_venta PK
        int id_negocio FK
        int id_usuario FK
        timestamp fecha_venta
        numeric total
        string metodo_pago
        string factura
        smallint tipo_comprobante
    }

    detalles_venta {
        int id_detalle PK
        int id_venta FK
        string id_producto FK
        numeric cantidad
        numeric precio_unitario
        numeric subtotal
    }

    dispositivos_clientes {
        int id_dispositivo PK
        int id_negocio FK
        string device_id UK
        string nombre_dispositivo
        string tipo_dispositivo
        timestamp fecha_vinculacion
        timestamp ultimo_acceso
    }
```

---

## 5. Biblioteca Exhaustiva de APIs (API Catalog & Reference)

Todos los endpoints que devuelven JSON usan codificación UTF-8. Las rutas protegidas exigen el encabezado HTTP:
```http
Authorization: Bearer <JWT_ACCESS_TOKEN>
```

---

### 5.1 Módulo 1: Salud y Verificación

#### `GET /api/ping`
- **Descripción:** Health check público del gateway y conectividad general.
- **Seguridad:** Pública (sin token).
- **Respuesta 200 OK:**
```json
{
  "status": "online",
  "message": "stock-api is up"
}
```

---

### 5.2 Módulo 2: Autenticación de Clientes y Terminales (`stock-api`)

#### `POST /api/auth/login`
- **Descripción:** Inicio de sesión de usuarios de comercio (dueños y cajeros). Valida la contraseña con `bcrypt`, valida la terminal activa y genera el par de tokens.
- **Seguridad:** Pública.
- **Payload:**
```json
{
  "usuario": "carlos",
  "password": "miPasswordSegura",
  "dispositivo": "Caja Principal (ThinkPad T480)"
}
```
- **Respuesta 200 OK:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsIn...",
  "refresh_token": "9a8b7c6d5e4f3a2b1c...",
  "usuario": {
    "id_usuario": 14,
    "id_negocio": 3,
    "nombre": "Carlos Gómez",
    "rol": "dueño",
    "permisos": ["vender", "stock", "usuarios", "precios", "reportes"]
  }
}
```
- **Errores:** `400 Bad Request` (campos faltantes), `401 Unauthorized` (credenciales inválidas).

---

#### `POST /api/auth/refresh`
- **Descripción:** Intercambia un refresh token vigente por un nuevo par de access token y refresh token (rotación automática).
- **Seguridad:** Pública (requiere refresh token en body).
- **Payload:**
```json
{
  "refresh_token": "9a8b7c6d5e4f3a2b1c..."
}
```
- **Respuesta 200 OK:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsIn...",
  "refresh_token": "nuevo_token_rotado_32_bytes..."
}
```

---

#### `POST /api/auth/logout`
- **Descripción:** Cierra la sesión revocando y eliminando el refresh token de la base de datos.
- **Seguridad:** Pública.
- **Payload:** `{ "refresh_token": "..." }`
- **Respuesta 204 No Content:** Sin cuerpo.

---

#### `POST /api/auth/obtener-tailscale-key`
- **Descripción:** Valida la cuenta del cliente y genera una Auth Key de Tailscale de 90 días para conectar una nueva computadora al túnel P2P (control de hasta 4 terminales).
- **Seguridad:** Pública (valida credenciales del dueño en el cuerpo).
- **Payload:**
```json
{
  "email": "duenio@comercio.com",
  "password": "password123",
  "device_id": "dev-a8b2-c3d4",
  "nombre_dispositivo": "Terminal Mostrador 2",
  "tipo_dispositivo": "desktop"
}
```
- **Respuesta 200 OK:**
```json
{
  "ts_auth_key": "tskey-auth-kXXXXX-...",
  "id_negocio": 3,
  "dispositivos_activos": 2,
  "max_dispositivos": 4
}
```
- **Errores:** `403 Forbidden` ("Límite máximo de 4 dispositivos alcanzado para este negocio").

---

#### `GET /api/auth/negocios/:id_negocio/dispositivos`
- **Descripción:** Lista todas las terminales registradas y activas de un negocio.
- **Seguridad:** Requiere Bearer Token (Rol: `superadmin` o `dueño` del mismo negocio).
- **Respuesta 200 OK:**
```json
[
  {
    "id_dispositivo": 5,
    "device_id": "dev-a8b2-c3d4",
    "nombre_dispositivo": "Caja Mostrador",
    "tipo_dispositivo": "desktop",
    "fecha_vinculacion": "2026-09-08T10:00:00Z",
    "ultimo_acceso": "2026-09-08T15:30:00Z"
  }
]
```

---

#### `DELETE /api/auth/negocios/:id_negocio/dispositivos/:id_dispositivo`
- **Descripción:** Desvincula una terminal liberando el cupo del comercio.
- **Seguridad:** Requiere Bearer Token (Rol: `superadmin` o `dueño` del mismo negocio).
- **Respuesta 200 OK:** `{ "ok": true }`

---

### 5.3 Módulo 3: Autenticación SuperAdmin con 2FA TOTP (`stock-api`)

#### `POST /api/admin/login` (Paso 1)
- **Descripción:** Autentica usuario y contraseña del administrador global. Comprueba si tiene 2FA configurado y genera un token temporal.
- **Seguridad:** Pública.
- **Payload:**
```json
{
  "usuario": "admin",
  "password": "claveSuperSegura"
}
```
- **Respuesta 200 OK (Si es Onboarding inicial):**
```json
{
  "requiere_2fa": true,
  "setup_totp": true,
  "qr_code": "data:image/png;base64,iVBORw0KGgo...",
  "secret": "JBSWY3DPEHPK3PXP",
  "temp_token": "eyJhbGciOiJIUzI1NiIsIn..."
}
```
- **Respuesta 200 OK (Si ya tiene 2FA activo):**
```json
{
  "requiere_2fa": true,
  "setup_totp": false,
  "temp_token": "eyJhbGciOiJIUzI1NiIsIn..."
}
```

---

#### `POST /api/admin/totp/verificar` (Paso 2)
- **Descripción:** Valida el código de 6 dígitos numéricos contra el secreto temporal o persistido. Si es válido, activa el 2FA en PostgreSQL y emite el JWT de SuperAdmin.
- **Seguridad:** Pública (requiere `temp_token` firmado de 5 min TTL).
- **Payload:**
```json
{
  "temp_token": "eyJhbGciOiJIUzI1NiIsIn...",
  "codigo": "729401"
}
```
- **Respuesta 200 OK:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsIn...",
  "admin": {
    "id_admin": 1,
    "usuario": "admin",
    "nombre": "Administrador Principal",
    "email": "admin@stockapp.com"
  }
}
```
- **Errores:** `401 Unauthorized` ("Código de autenticación inválido o expirado").

---

### 5.4 Módulo 4: Introspección Interna (`stock-api`)

#### `GET /internal/session`
- **Descripción:** Endpoint interno consumido exclusivamente por `stock-operations` para validar la firma de los JWT entrantes.
- **Seguridad:** Solo accesible dentro de la red interna de contenedores.
- **Headers:** `Authorization: Bearer <token>`
- **Respuesta 200 OK:**
```json
{
  "sub": 14,
  "negocio": 3,
  "nombre": "Carlos Gómez",
  "rol": "dueño",
  "permisos": ["vender", "stock", "usuarios"]
}
```

---

### 5.5 Módulo 5: Productos y Catálogo Maestro (`stock-operations`)

#### `GET /api/productos?buscar={termino}&limite={n}`
- **Descripción:** Búsqueda combinada en el Catálogo Maestro de 21.800+ productos y el stock interno del negocio. Soporta búsqueda por código de barras exacto o coincidencia por texto en descripción y marca.
- **Seguridad:** Requiere Bearer Token (Cajero o Dueño).
- **Parámetros Query:**
  - `buscar` (string, opcional): Ej: `"coca"`, `"7790895000997"`.
  - `limite` (int, default: 20, max: 200).
- **Respuesta 200 OK:**
```json
[
  {
    "id_producto": "7790895000997",
    "descripcion": "Gaseosa Coca Cola Sabor Original 500ml",
    "cantidad_presentacion": "500",
    "unidad_medida": "ml",
    "marca": "Coca Cola",
    "en_stock": true,
    "precio_venta": 1500.0,
    "cantidad_disponible": 45.0
  }
]
```

---

#### `POST /api/productos`
- **Descripción:** Crea un nuevo producto personalizado que no existe en el catálogo maestro y lo incorpora al stock del negocio en una única transacción.
- **Seguridad:** Requiere Bearer Token (Rol: `dueño`).
- **Payload:**
```json
{
  "descripcion": "Alfajor Artesanal Chocolate",
  "marca": "El Paraíso",
  "cantidad_presentacion": "70",
  "unidad_medida": "gr",
  "precio_venta": 850.0,
  "precio_costo": 500.0,
  "stock": 30,
  "stock_minimo": 5,
  "id_proveedor": 2
}
```
- **Respuesta 201 Created:** Retorna el producto con su `id_producto` autogenerado (UUID o código manual).

---

### 5.6 Módulo 6: Stock e Inventario (`stock-operations`)

#### `GET /api/stock`
- **Descripción:** Devuelve la lista completa de artículos cargados en el inventario del negocio actual.
- **Seguridad:** Requiere Bearer Token (Cajero o Dueño).
- **Respuesta 200 OK:**
```json
[
  {
    "id_producto": "7790895000997",
    "descripcion": "Gaseosa Coca Cola Sabor Original 500ml",
    "marca": "Coca Cola",
    "cantidad_presentacion": "500",
    "unidad_medida": "ml",
    "cantidad_disponible": 24.0,
    "precio_venta": 1500.0,
    "stock_minimo": 10.0,
    "id_proveedor": 1,
    "proveedor": "Distribuidora Bebidas del Sur"
  }
]
```

---

#### `PUT /api/stock/{id_producto}`
- **Descripción:** Modifica el stock disponible, precio de venta, costo o proveedor de un artículo del negocio.
- **Seguridad:** Requiere Bearer Token (Rol: `dueño`).
- **Payload:**
```json
{
  "cantidad_disponible": 50.0,
  "precio_venta": 1650.0,
  "precio_costo": 1100.0,
  "stock_minimo": 12.0,
  "id_proveedor": 1
}
```
- **Respuesta 200 OK:** `{ "mensaje": "stock actualizado correctamente" }`

---

#### `GET /api/stock/frecuentes`
- **Descripción:** Devuelve los 12 productos con mayor rotación de ventas en el negocio para la grilla de accesos rápidos del POS.
- **Seguridad:** Requiere Bearer Token (Cajero o Dueño).
- **Respuesta 200 OK:** Array de productos con precio y stock vigente.

---

### 5.7 Módulo 7: Ventas, Tickets y Facturación (`stock-operations`)

#### `POST /api/ventas`
- **Descripción:** Registra una venta atómica en PostgreSQL. Calcula los importes en el servidor (previniendo adulteraciones), valida existencias y descuenta el stock de inmediato.
- **Seguridad:** Requiere Bearer Token (Cajero o Dueño).
- **Payload:**
```json
{
  "metodo_pago": "efectivo",
  "items": [
    { "id_producto": "7790895000997", "cantidad": 2 },
    { "id_producto": "7791234567890", "cantidad": 1 }
  ]
}
```
- **Respuesta 201 Created:**
```json
{
  "id_venta": 482,
  "fecha_hora": "2026-09-08 14:25:00",
  "metodo_pago": "efectivo",
  "total": 3850.0,
  "items": [
    {
      "id_producto": "7790895000997",
      "descripcion": "Gaseosa Coca Cola 500ml",
      "cantidad": 2,
      "precio_unitario": 1500.0,
      "subtotal": 3000.0
    },
    {
      "id_producto": "7791234567890",
      "descripcion": "Turrón de Maní 25g",
      "cantidad": 1,
      "precio_unitario": 850.0,
      "subtotal": 850.0
    }
  ]
}
```
- **Errores:** `400 Bad Request` ("Stock insuficiente para el producto X").

---

#### `GET /api/ventas?fecha={YYYY-MM-DD}&limite={n}`
- **Descripción:** Historial de ventas del negocio actual con desglose de ítems, totales y filtros por fecha.
- **Seguridad:** Requiere Bearer Token (Cajero o Dueño).
- **Respuesta 200 OK:** Array de ventas ordenadas de más reciente a más antigua.

---

#### `POST /api/ventas/{id_venta}/factura`
- **Descripción:** Asigna un número de comprobante fiscal o interno a la venta (`"C 0001-00000142"`).
- **Seguridad:** Requiere Bearer Token (Rol: `dueño`).
- **Respuesta 200 OK:**
```json
{
  "id_venta": 482,
  "factura": "C 0001-00000142",
  "tipo_comprobante": 11
}
```

---

### 5.8 Módulo 8: Proveedores y Aumentos Masivos (`stock-operations`)

#### `GET /api/proveedores`
- **Descripción:** Lista los proveedores registrados por el comercio.
- **Seguridad:** Requiere Bearer Token (Rol: `dueño`).

#### `POST /api/proveedores`
- **Payload:** `{ "nombre": "Distribuidora Golosinas S.A.", "telefono": "1144332211", "email": "ventas@golosinas.com" }`
- **Respuesta 201 Created:** `{ "id_proveedor": 4, "nombre": "Distribuidora Golosinas S.A." }`

#### `POST /api/proveedores/{id_proveedor}/aumento`
- **Descripción:** Aplica un incremento porcentual masivo sobre el precio de venta de **todos** los productos asociados a dicho proveedor.
- **Seguridad:** Requiere Bearer Token (Rol: `dueño`).
- **Payload:**
```json
{
  "porcentaje": 12.5
}
```
- **Respuesta 200 OK:**
```json
{
  "mensaje": "se actualizaron los precios de 38 productos",
  "productos_afectados": 38,
  "porcentaje_aplicado": 12.5
}
```

---

### 5.9 Módulo 9: Resumen de Caja y Negocio (`stock-operations`)

#### `GET /api/resumen`
- **Descripción:** Métrica financiera en tiempo real del comercio para la fecha en curso en horario argentino (`America/Argentina/Buenos_Aires`).
- **Seguridad:** Requiere Bearer Token (Rol: `dueño`).
- **Respuesta 200 OK:**
```json
{
  "total_ventas": 124500.0,
  "cantidad_operaciones": 42,
  "desglose_pagos": {
    "efectivo": 78000.0,
    "transferencia": 32500.0,
    "tarjeta": 14000.0
  },
  "ticket_promedio": 2964.28
}
```

---

#### `GET /api/negocio` y `PUT /api/negocio`
- **Descripción:** Consulta y actualización de datos de la empresa (nombre, CUIT, punto de venta, dirección fiscal).
- **Seguridad:** Requiere Bearer Token (Rol: `dueño`).

---

### 5.10 Módulo 10: Usuarios y Empleados (`stock-operations`)

#### `GET /api/usuarios`
- **Descripción:** Lista los usuarios registrados pertenecientes al negocio actual.
- **Seguridad:** Requiere Bearer Token (Rol: `dueño`).

#### `POST /api/usuarios`
- **Descripción:** Da de alta un nuevo empleado cajero.
- **Seguridad:** Requiere Bearer Token (Rol: `dueño`).
- **Payload:**
```json
{
  "nombre": "Lucía Martínez",
  "usuario": "lucia",
  "password": "claveCajero123",
  "rol": "cajero"
}
```
- **Respuesta 201 Created:** `{ "id_usuario": 18, "usuario": "lucia", "rol": "cajero" }`

#### `PUT /api/usuarios/{id_usuario}/password`
- **Descripción:** Modifica la contraseña de un empleado.
- **Seguridad:** Requiere Bearer Token (Rol: `dueño`).

---

### 5.11 Módulo 11: Administración Global de Inquilinos (`stock-operations`)

*Todas las rutas de este módulo exigen estrictamente JWT de SuperAdmin con `rol: 'superadmin'`.*

#### `GET /api/admin/negocios`
- **Descripción:** Lista todos los comercios del sistema con sus datos de contacto, fecha de alta y datos de su dueño.
- **Respuesta 200 OK:** Array con la totalidad de los negocios registrados.

#### `POST /api/admin/negocios`
- **Descripción:** Da de alta un nuevo comercio en la plataforma y genera simultáneamente la cuenta del dueño.
- **Payload:**
```json
{
  "nombre_negocio": "Maxikiosco Los Hermanos",
  "cuit": "20334455667",
  "direccion": "Av. San Martín 450",
  "telefono": "02944112233",
  "email_negocio": "contacto@loshermanos.com",
  "nombre_dueno": "Martín Rodríguez",
  "usuario": "mrodriguez",
  "password": "claveInicialDueño123"
}
```
- **Respuesta 201 Created:**
```json
{
  "id_negocio": 8,
  "id_usuario": 22,
  "nombre_negocio": "Maxikiosco Los Hermanos",
  "usuario": "mrodriguez",
  "nombre_dueno": "Martín Rodríguez"
}
```

#### `PUT /api/admin/negocios/{id_negocio}`
- **Descripción:** Edita cualquier información comercial o credencial de un negocio cliente.

#### `PUT /api/admin/usuarios/{id_usuario}/password`
- **Descripción:** Restablece la contraseña del usuario dueño de cualquier negocio en caso de olvido o pérdida de acceso.

---

## 6. Mecanismos de Seguridad y Políticas Criptográficas

1. **Almacenamiento de Contraseñas:** Algoritmo **bcrypt** con factor de coste (*work factor*) 10 en todas las tablas (`super_admins` y `usuarios`).
2. **Tokens de Acceso (Access Tokens):** Formato JSON Web Token (JWT) firmado con algoritmo simétrico **HMAC-SHA256 (HS256)** mediante secreto de 64 caracteres (`JWT_SECRET`). Expiración fijada en 15 minutos para clientes y 2 horas para SuperAdmin.
3. **Tokens de Refresco (Refresh Tokens):** Entropía criptográfica de 256 bits generada con `crypto.randomBytes(32)`. El token plano viaja únicamente en la respuesta HTTP; en la base de datos se persiste exclusivamente su digest **SHA-256**. Rotación forzosa en cada consumo.
4. **Segundo Factor (TOTP):** Estándar RFC 6238 con ventana de tolerancia de 1 paso (±30 segundos) y algoritmo HMAC-SHA-1 con clave Base32.
5. **Cifrado de Secretos Internos:** Las claves fiscales de ARCA o certificados se almacenan cifrados con algoritmo simétrico **AES-256-GCM** mediante la clave maestra `CONFIG_SECRET_KEY`.

---

## 7. Guía de Despliegue y Configuración de Entorno

### 7.1 Variables de Entorno de Producción (`.env`)
```bash
# PostgreSQL
DB_USER=stock_admin
DB_PASSWORD=<CLAVE_ALEATORIA_64_HEX>
DB_NAME=stock_db

# Seguridad
JWT_SECRET=<SECRETO_ALEATORIO_64_HEX>
CONFIG_SECRET_KEY=<SECRETO_ALEATORIO_64_HEX>

# Puertos
PORT_API=3000
PORT_ADMIN=5173
```

### 7.2 Comandos de Inicialización
```bash
# Compilar e iniciar contenedores
podman-compose -f docker-compose.prod.yml up -d --build

# Crear Administrador inicial
podman exec -it stockapp_gateway npm run crear-admin:prod

# Precargar catálogo de 21.800 productos
cd ~/stockapp/backend/services/stock-operations
podman run --rm --network stockapp_network \
  -v $(pwd):/app:ro -w /app \
  -e DB_HOST=postgres -e DB_USER=stock_admin -e DB_PASSWORD=... -e DB_NAME=stock_db \
  docker.io/library/golang:1.25-alpine go run pkg/importar_db.go
```
