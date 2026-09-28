// stock-api: the only entry point for the desktop app.
// - Users and sessions (login/refresh/logout) are resolved here.
// - /internal/session is the introspection endpoint stock-operations uses
//   to validate each request's token.
// - Everything else under /api is forwarded to stock-operations (Go), which
//   is never exposed directly.
import { pool } from './db/pool.js';
import { env } from './config/env.js';
import { createApp } from './app.js';
import { jwtSecret } from './services/tokenService.js';

const app = createApp();

const port = Number(env('PUERTO', '3000'));

// Fallar al arrancar, no en el primer login, si falta el secreto de firma.
jwtSecret();

try {
  await pool.query('SELECT 1');
} catch (e) {
  console.error(`[fatal] could not connect to PostgreSQL: ${e}`);
  process.exit(1);
}

app.listen(port, () => {
  console.log(`[info] stock-api listening on port ${port} (stock-operations: ${env('GO_INTERNO', 'http://localhost:8080')})`);
});
