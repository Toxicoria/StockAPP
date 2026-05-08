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
    rol VARCHAR(50) DEFAULT 'cajero', -- Puede ser 'admin' o 'cajero'
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