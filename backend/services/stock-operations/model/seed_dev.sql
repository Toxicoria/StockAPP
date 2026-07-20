-- ==============================================================================
-- 🌱 DATOS DE DESARROLLO — NO usar en producción
--    Crea un negocio demo y dos usuarios para probar el login:
--    admin@dev.local  / admin123   (rol admin: ve todo)
--    cajero@dev.local / cajero123  (rol cajero: solo caja y stock)
--    Además: 3 proveedores y stock mínimo variado para ver las alertas.
--
--    Aplicar con: task db:seed
-- ==============================================================================

INSERT INTO negocios (nombre_negocio, direccion, cuit)
SELECT 'Negocio Demo', 'Calle Falsa 123', '20-00000000-0'
WHERE NOT EXISTS (SELECT 1 FROM negocios WHERE nombre_negocio = 'Negocio Demo');

INSERT INTO usuarios (id_negocio, nombre, email, password_hash, rol)
SELECT n.id_negocio, 'Admin Demo', 'admin@dev.local',
       '$2a$10$g0roVPPbLD7UV60qVSnJnue2ICq5mbHxd638t84En5Au6a5cz/WSC', -- bcrypt("admin123")
       'admin'
  FROM negocios n
 WHERE n.nombre_negocio = 'Negocio Demo'
ON CONFLICT (email) DO NOTHING;

INSERT INTO usuarios (id_negocio, nombre, email, password_hash, rol)
SELECT n.id_negocio, 'Cajero Demo', 'cajero@dev.local',
       '$2b$10$5RCG5LBdtfZ1HAas76b6o..iYFl9kQx/P/LdmBtaCE540OMJXoyGS', -- bcrypt("cajero123")
       'cajero'
  FROM negocios n
 WHERE n.nombre_negocio = 'Negocio Demo'
ON CONFLICT (email) DO NOTHING;

-- Proveedores demo (idempotente por el UNIQUE (id_negocio, nombre))
INSERT INTO proveedores (id_negocio, nombre)
SELECT n.id_negocio, proveedor
  FROM negocios n,
       (VALUES ('Cervecería del Bolsón'), ('Dulces del Valle'), ('Distribuidora Austral')) AS p(proveedor)
 WHERE n.nombre_negocio = 'Negocio Demo'
ON CONFLICT (id_negocio, nombre) DO NOTHING;

-- Stock mínimo variado en el inventario existente, para ver las alertas
-- "para reponer" sin cargar datos a mano (solo donde aún no se configuró).
UPDATE stock_interno s
   SET stock_minimo = CASE WHEN s.cantidad_disponible <= 5 THEN 6 ELSE 4 END
  FROM negocios n
 WHERE s.id_negocio = n.id_negocio
   AND n.nombre_negocio = 'Negocio Demo'
   AND COALESCE(s.stock_minimo, 0) = 0;
