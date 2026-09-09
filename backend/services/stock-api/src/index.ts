// stock-api: the only entry point for the desktop app.
// - Users and sessions (login/refresh/logout) are resolved here.
// - /internal/session is the introspection endpoint stock-operations uses
//   to validate each request's token.
// - Everything else under /api is forwarded to stock-operations (Go), which
//   is never exposed directly.
import express from 'express';
import { pool } from './db/pool.js';
import { env } from './config/env.js';
import { cors } from './middleware/cors.js';
import { securityHeaders } from './middleware/security.js';
import { errorHandler, notFound } from './middleware/errorHandler.js';
import { healthRoutes } from './routes/healthRoutes.js';
import { authRoutes } from './routes/authRoutes.js';
import { adminAuthRoutes } from './routes/adminAuthRoutes.js';
import { sessionRoutes } from './routes/sessionRoutes.js';
import { desktopRoutes } from './routes/desktopRoutes.js';
import { mobileRoutes } from './routes/mobileRoutes.js';
import { proxyRoutes } from './routes/proxyRoutes.js';

const app = express();
app.use(express.json());
app.use(cors);
app.use(securityHeaders);

app.use(mobileRoutes);
app.use('/api', healthRoutes);
app.use('/api', desktopRoutes);
app.use('/api', authRoutes);
app.use('/api', adminAuthRoutes);
app.use(sessionRoutes);
app.use('/api', proxyRoutes);

app.use(notFound);
app.use(errorHandler);

const port = Number(env('PUERTO', '3000'));

let conectado = false;
for (let intento = 1; intento <= 15; intento++) {
  try {
    await pool.query('SELECT 1');
    conectado = true;
    break;
  } catch {
    console.warn(`[warn] esperando a PostgreSQL (intento ${intento}/15)...`);
    await new Promise((r) => setTimeout(r, 2000));
  }
}

if (!conectado) {
  console.error('[fatal] no se pudo conectar a PostgreSQL tras varios intentos');
  process.exit(1);
}

// Asegurar esquema de dispositivos cliente y claves de negocio
try {
  await pool.query(`
    ALTER TABLE negocios ADD COLUMN IF NOT EXISTS max_dispositivos INT NOT NULL DEFAULT 4;
    ALTER TABLE negocios ADD COLUMN IF NOT EXISTS ts_auth_key TEXT;
    CREATE TABLE IF NOT EXISTS dispositivos_cliente (
        id_dispositivo SERIAL PRIMARY KEY,
        id_negocio INT NOT NULL REFERENCES negocios(id_negocio) ON DELETE CASCADE,
        device_id VARCHAR(100) NOT NULL,
        nombre_dispositivo VARCHAR(100) NOT NULL,
        tipo_dispositivo VARCHAR(20) DEFAULT 'desktop',
        fecha_registro TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        ultima_conexion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        activo BOOLEAN DEFAULT TRUE,
        UNIQUE (id_negocio, device_id)
    );
    CREATE INDEX IF NOT EXISTS idx_dispositivos_negocio_activo ON dispositivos_cliente (id_negocio, activo);
  `);
  console.log('[info] esquema de dispositivos verificado y listo.');
} catch (e) {
  console.warn('[warn] error verificando esquema de dispositivos:', e);
}

app.listen(port, () => {
  console.log(`[info] stock-api listening on port ${port} (stock-operations: ${env('GO_INTERNO', 'http://localhost:8080')})`);
});
