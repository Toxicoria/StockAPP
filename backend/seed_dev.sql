-- ==============================================================================
-- 🌱 DATOS DE DESARROLLO — NO usar en producción
--    Crea un negocio demo y un usuario admin para probar el login:
--    email: admin@dev.local · contraseña: admin123
--
--    Aplicar con:
--    docker exec -i db_sistema_stock psql -U admin_dev -d stock_db < seed_dev.sql
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
