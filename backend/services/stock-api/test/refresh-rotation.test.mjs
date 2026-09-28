import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { test } from 'node:test';

test('a refresh token can be consumed only once under concurrency', async (t) => {
  const databaseUrl = process.env.TEST_DATABASE_URL;
  if (!databaseUrl) {
    t.skip('TEST_DATABASE_URL no está configurada');
    return;
  }

  process.env.DATABASE_URL = databaseUrl;
  process.env.JWT_SECRET = 'integration-test-jwt-secret-with-enough-entropy';

  const [{ pool }, { refresh }, { hashToken }] = await Promise.all([
    import('../dist/db/pool.js'),
    import('../dist/services/authService.js'),
    import('../dist/services/tokenService.js'),
  ]);

  const schemaUrl = new URL('../../stock-operations/model/esquema.sql', import.meta.url);
  await pool.query(await readFile(schemaUrl, 'utf8'));

  const suffix = `${process.pid}-${Date.now()}`;
  const email = `refresh-${suffix}@example.com`;
  const businessName = `Refresh test ${suffix}`;
  const oldToken = `old-refresh-token-${suffix}`;

  const business = await pool.query(
    `INSERT INTO negocios (nombre_negocio) VALUES ($1) RETURNING id_negocio`,
    [businessName],
  );
  const businessId = business.rows[0].id_negocio;
  t.after(async () => {
    await pool.query(`DELETE FROM negocios WHERE id_negocio = $1`, [businessId]);
    await pool.end();
  });

  const user = await pool.query(
    `INSERT INTO usuarios (id_negocio, nombre, email, password_hash, rol)
     VALUES ($1, 'Refresh Test', $2, 'unused-in-this-test', 'dueño')
     RETURNING id_usuario`,
    [businessId, email],
  );
  const userId = user.rows[0].id_usuario;

  await pool.query(
    `INSERT INTO refresh_tokens (id_usuario, token_hash, dispositivo, expira_en)
     VALUES ($1, $2, 'integration-test', NOW() + INTERVAL '1 hour')`,
    [userId, hashToken(oldToken)],
  );

  const results = await Promise.allSettled([refresh(oldToken), refresh(oldToken)]);
  const successful = results.filter((result) => result.status === 'fulfilled');
  const rejected = results.filter((result) => result.status === 'rejected');

  assert.equal(successful.length, 1);
  assert.equal(rejected.length, 1);
  assert.equal(rejected[0].reason?.status, 401);

  const tokens = await pool.query(
    `SELECT token_hash FROM refresh_tokens WHERE id_usuario = $1`,
    [userId],
  );
  assert.equal(tokens.rowCount, 1);
  assert.notEqual(tokens.rows[0].token_hash, hashToken(oldToken));
});
