-- ==============================================================================
-- 🌱 DATOS DE DESARROLLO — NO usar en producción
--    Crea un negocio demo, dos usuarios de prueba, catálogo maestro de ejemplo:
--    duenio@dev.local / duenio123  (rol dueño)
--    cajero@dev.local / cajero123  (rol cajero)
--    Además: productos del catálogo maestro (EAN), proveedores y stock inicial.
--
--    Aplicar con: task db:seed
-- ==============================================================================

INSERT INTO negocios (nombre_negocio, direccion, cuit)
SELECT 'Negocio Demo', 'Calle Falsa 123', '20-00000000-0'
WHERE NOT EXISTS (SELECT 1 FROM negocios WHERE nombre_negocio = 'Negocio Demo');

-- SuperAdmin del Sistema (acceso exclusivo a front-admin: admin / admin123)
INSERT INTO super_admins (usuario, email, password_hash, nombre)
SELECT 'admin', 'admin@stockapp.local',
       '$2a$10$g0roVPPbLD7UV60qVSnJnue2ICq5mbHxd638t84En5Au6a5cz/WSC', -- bcrypt("admin123")
       'Administrador del Sistema'
WHERE NOT EXISTS (SELECT 1 FROM super_admins WHERE usuario = 'admin');

-- Dueño del negocio (cliente que usa la app, tiene todos los permisos)
INSERT INTO usuarios (id_negocio, nombre, email, password_hash, rol)
SELECT n.id_negocio, 'Dueño Demo', 'duenio@dev.local',
       '$2a$10$g0roVPPbLD7UV60qVSnJnue2ICq5mbHxd638t84En5Au6a5cz/WSC', -- bcrypt("admin123")
       'dueño'
  FROM negocios n
 WHERE n.nombre_negocio = 'Negocio Demo'
   AND NOT EXISTS (SELECT 1 FROM usuarios u WHERE u.id_negocio = n.id_negocio AND LOWER(u.nombre) = LOWER('Dueño Demo'));

-- Cajero (empleado del dueño, solo caja y consulta)
INSERT INTO usuarios (id_negocio, nombre, email, password_hash, rol)
SELECT n.id_negocio, 'Cajero Demo', 'cajero@dev.local',
       '$2b$10$5RCG5LBdtfZ1HAas76b6o..iYFl9kQx/P/LdmBtaCE540OMJXoyGS', -- bcrypt("cajero123")
       'cajero'
  FROM negocios n
 WHERE n.nombre_negocio = 'Negocio Demo'
   AND NOT EXISTS (SELECT 1 FROM usuarios u WHERE u.id_negocio = n.id_negocio AND LOWER(u.nombre) = LOWER('Cajero Demo'));

-- Catálogo Maestro de Productos (EAN) para probar el autocompletado y el escáner
INSERT INTO productos (id_producto, productos_ean, productos_descripcion, productos_cantidad_presentacion, productos_unidad_medida_presentacion, productos_marca)
VALUES
  ('7791234567890', '7791234567', 'Coca Cola Original', '1.5', 'L', 'Coca Cola'),
  ('7791234567891', '7791234568', 'Coca Cola Zero', '1.5', 'L', 'Coca Cola'),
  ('7790001001010', '7790001001', 'Galletitas Chocolinas', '250', 'g', 'Bagley'),
  ('7790002002020', '7790002002', 'Leche Entera Tetra', '1', 'L', 'La Serenísima'),
  ('7790003003030', '7790003003', 'Cerveza Clasica Botella', '1', 'L', 'Quilmes'),
  ('7790004004040', '7790004004', 'Agua Mineral Sin Gas', '500', 'ml', 'Villavicencio'),
  ('7790005005050', '7790005005', 'Yerba Mate Playadito', '1', 'kg', 'Playadito')
ON CONFLICT (id_producto) DO NOTHING;

-- Proveedores demo (idempotente por el UNIQUE (id_negocio, nombre))
INSERT INTO proveedores (id_negocio, nombre)
SELECT n.id_negocio, proveedor
  FROM negocios n,
       (VALUES ('Cervecería del Bolsón'), ('Dulces del Valle'), ('Distribuidora Austral')) AS p(proveedor)
 WHERE n.nombre_negocio = 'Negocio Demo'
ON CONFLICT (id_negocio, nombre) DO NOTHING;

-- Stock inicial demo en inventario para el negocio demo
INSERT INTO stock_interno (id_negocio, id_producto, cantidad_disponible, precio_venta, stock_minimo)
SELECT n.id_negocio, '7791234567890', 24.00, 1800.00, 5.00
  FROM negocios n WHERE n.nombre_negocio = 'Negocio Demo'
ON CONFLICT (id_negocio, id_producto) DO NOTHING;

INSERT INTO stock_interno (id_negocio, id_producto, cantidad_disponible, precio_venta, stock_minimo)
SELECT n.id_negocio, '7790001001010', 3.00, 1200.00, 6.00
  FROM negocios n WHERE n.nombre_negocio = 'Negocio Demo'
ON CONFLICT (id_negocio, id_producto) DO NOTHING;

-- Key por defecto para entorno de desarrollo
UPDATE negocios
   SET ts_auth_key = 'tskey-auth-dev-local-demo-key'
 WHERE nombre_negocio = 'Negocio Demo' AND (ts_auth_key IS NULL OR ts_auth_key = '');

