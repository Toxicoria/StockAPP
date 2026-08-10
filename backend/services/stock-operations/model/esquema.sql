-- ==============================================================================
-- 🗃️ 0. PRODUCTOS (Catálogo maestro compartido — lo carga importar_db.go)
--    Se define acá porque stock_interno y detalles_venta le apuntan con FK.
-- ==============================================================================
CREATE TABLE IF NOT EXISTS productos (
    id_producto VARCHAR(100) PRIMARY KEY,
    productos_ean VARCHAR(10),
    productos_descripcion VARCHAR(255),
    productos_cantidad_presentacion VARCHAR(50),
    productos_unidad_medida_presentacion VARCHAR(50),
    productos_marca VARCHAR(100)
);

-- ==============================================================================
-- 🏢 1. NEGOCIOS (Las sucursales o clientes que usarán el sistema)
-- ==============================================================================
CREATE TABLE IF NOT EXISTS negocios (
    id_negocio SERIAL PRIMARY KEY,
    nombre_negocio VARCHAR(150) NOT NULL,
    direccion VARCHAR(255),
    cuit VARCHAR(20),
    fecha_alta TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ==============================================================================
-- 👥 2. USUARIOS (Los empleados/cajeros/admin de cada negocio)
-- ==============================================================================
CREATE TABLE IF NOT EXISTS usuarios (
    id_usuario SERIAL PRIMARY KEY,
    id_negocio INT NOT NULL,
    nombre VARCHAR(100) NOT NULL,
    email VARCHAR(150) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    rol VARCHAR(50) DEFAULT 'cajero', -- 'dueño' o 'cajero'
    fecha_alta TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (id_negocio) REFERENCES negocios(id_negocio) ON DELETE CASCADE
);

-- ==============================================================================
-- 📦 3. STOCK INTERNO (El inventario real de cada negocio)
-- ==============================================================================
CREATE TABLE IF NOT EXISTS stock_interno (
    id_stock SERIAL PRIMARY KEY,
    id_negocio INT NOT NULL,
    id_producto VARCHAR(100) NOT NULL,
    cantidad_disponible NUMERIC(10, 2) DEFAULT 0,
    precio_venta NUMERIC(10, 2) DEFAULT 0,
    ultima_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (id_negocio) REFERENCES negocios(id_negocio) ON DELETE CASCADE,
    FOREIGN KEY (id_producto) REFERENCES productos(id_producto) ON DELETE RESTRICT,
    -- Un negocio no puede tener el mismo producto dos veces en su inventario
    UNIQUE (id_negocio, id_producto) 
);

-- ==============================================================================
-- 🧾 4. VENTAS (El "Ticket" general o cabecera)
-- ==============================================================================
CREATE TABLE IF NOT EXISTS ventas (
    id_venta SERIAL PRIMARY KEY,
    id_negocio INT NOT NULL,
    id_usuario INT NOT NULL,
    total_venta NUMERIC(12, 2) NOT NULL,
    metodo_pago VARCHAR(50) DEFAULT 'efectivo',
    fecha_hora TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (id_negocio) REFERENCES negocios(id_negocio) ON DELETE CASCADE,
    FOREIGN KEY (id_usuario) REFERENCES usuarios(id_usuario) ON DELETE RESTRICT
);

-- ==============================================================================
-- 🛒 5. DETALLES DE VENTA (Los productos dentro del ticket)
-- ==============================================================================
CREATE TABLE IF NOT EXISTS detalles_venta (
    id_detalle SERIAL PRIMARY KEY,
    id_venta INT NOT NULL,
    id_producto VARCHAR(100) NOT NULL,
    cantidad_llevada NUMERIC(10, 2) NOT NULL,
    precio_unitario_cobrado NUMERIC(10, 2) NOT NULL,
    subtotal NUMERIC(12, 2) NOT NULL,
    FOREIGN KEY (id_venta) REFERENCES ventas(id_venta) ON DELETE CASCADE,
    FOREIGN KEY (id_producto) REFERENCES productos(id_producto) ON DELETE RESTRICT
);

-- ==============================================================================
-- 🔑 6. REFRESH TOKENS (Sesiones activas por dispositivo)
-- ==============================================================================
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id             SERIAL PRIMARY KEY,
    id_usuario     INT NOT NULL,
    token_hash     VARCHAR(255) UNIQUE NOT NULL,  -- SHA-256 del token, nunca el token en crudo
    dispositivo    VARCHAR(100),                  -- Ej: "PC-Caja1", "Notebook-Admin"
    expira_en      TIMESTAMP NOT NULL,            -- 30 días desde creación
    fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (id_usuario) REFERENCES usuarios(id_usuario) ON DELETE CASCADE
);

-- ==============================================================================
-- 🚚 7. PROVEEDORES (Cada negocio maneja su propia lista de proveedores)
-- ==============================================================================
CREATE TABLE IF NOT EXISTS proveedores (
    id_proveedor SERIAL PRIMARY KEY,
    id_negocio   INT NOT NULL,
    nombre       VARCHAR(150) NOT NULL,
    ultima_actualizacion_precios TIMESTAMP,      -- última vez que se aplicó un aumento
    fecha_alta   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (id_negocio) REFERENCES negocios(id_negocio) ON DELETE CASCADE,
    UNIQUE (id_negocio, nombre)
);

-- ==============================================================================
-- 🔧 8. MIGRACIONES (Idempotentes: corren también sobre bases existentes.
--    El init de compose solo ejecuta este archivo con el volumen vacío;
--    `task db:esquema` lo re-aplica sobre bases vivas sin destruir datos.)
-- ==============================================================================
ALTER TABLE stock_interno ADD COLUMN IF NOT EXISTS stock_minimo NUMERIC(10, 2) DEFAULT 0;
ALTER TABLE stock_interno ADD COLUMN IF NOT EXISTS id_proveedor INT REFERENCES proveedores(id_proveedor) ON DELETE SET NULL;

-- Facturación: numeración local hoy, esquema listo para ARCA (WSFEv1) mañana.
-- factura = string de display "C 0001-00000214"; cae/cae_vencimiento quedan
-- NULL hasta integrar el web service de ARCA.
ALTER TABLE ventas ADD COLUMN IF NOT EXISTS factura VARCHAR(30);
ALTER TABLE ventas ADD COLUMN IF NOT EXISTS tipo_comprobante SMALLINT DEFAULT 11;  -- 11 = Factura C (monotributo)
ALTER TABLE ventas ADD COLUMN IF NOT EXISTS nro_comprobante INT;
ALTER TABLE ventas ADD COLUMN IF NOT EXISTS cae VARCHAR(14);
ALTER TABLE ventas ADD COLUMN IF NOT EXISTS cae_vencimiento DATE;

ALTER TABLE negocios ADD COLUMN IF NOT EXISTS punto_venta VARCHAR(4) DEFAULT '0001';
ALTER TABLE negocios ADD COLUMN IF NOT EXISTS proximo_numero_factura INT NOT NULL DEFAULT 1;

-- Clave fiscal de ARCA (u otro secreto del negocio): se guarda cifrada
-- (AES-256-GCM, ver pkg/api/secreto.go) y nunca se devuelve por la API.
ALTER TABLE negocios ADD COLUMN IF NOT EXISTS clave_fiscal_cifrada BYTEA;

-- Datos de contacto del negocio (recabados en el onboarding inicial).
ALTER TABLE negocios ADD COLUMN IF NOT EXISTS nombre_dueno   VARCHAR(150);
ALTER TABLE negocios ADD COLUMN IF NOT EXISTS telefono       VARCHAR(50);
ALTER TABLE negocios ADD COLUMN IF NOT EXISTS email_negocio  VARCHAR(150);

CREATE INDEX IF NOT EXISTS idx_ventas_negocio_fecha ON ventas (id_negocio, fecha_hora);