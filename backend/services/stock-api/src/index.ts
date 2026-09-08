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
import { proxyRoutes } from './routes/proxyRoutes.js';

const app = express();
app.use(express.json());
app.use(cors);
app.use(securityHeaders);

app.use('/api', healthRoutes);
app.use('/api', authRoutes);
app.use('/api', adminAuthRoutes);
app.use(sessionRoutes);
app.use('/api', proxyRoutes);

app.use(notFound);
app.use(errorHandler);

const port = Number(env('PUERTO', '3000'));

try {
  await pool.query('SELECT 1');
} catch (e) {
  console.error(`[fatal] could not connect to PostgreSQL: ${e}`);
  process.exit(1);
}

app.listen(port, () => {
  console.log(`[info] stock-api listening on port ${port} (stock-operations: ${env('GO_INTERNO', 'http://localhost:8080')})`);
});
