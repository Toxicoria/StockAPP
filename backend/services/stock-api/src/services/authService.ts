import bcrypt from 'bcryptjs';
import type { Pool, PoolClient } from 'pg';
import { pool } from '../db/pool.js';
import { ApiError } from '../errors/ApiError.js';
import { hashToken, newRefreshToken, signAccessToken } from './tokenService.js';

interface UserRow {
  id_usuario: number;
  id_negocio: number;
  nombre_negocio: string;
  nombre: string;
  password_hash: string;
  rol: string;
}

type QueryExecutor = Pick<Pool | PoolClient, 'query'>;

async function issueTokens(db: QueryExecutor, user: Omit<UserRow, 'password_hash'>, device: string) {
  const access_token = signAccessToken(user);
  const { token: refresh_token, expiresAt } = newRefreshToken();

  await db.query(
    `INSERT INTO refresh_tokens (id_usuario, token_hash, dispositivo, expira_en)
     VALUES ($1, $2, $3, $4)`,
    [user.id_usuario, hashToken(refresh_token), device, expiresAt],
  );

  return {
    access_token,
    refresh_token,
    nombre: user.nombre,
    rol: user.rol,
    negocio: user.nombre_negocio,
  };
}

export async function login(email: string, password: string, device: string) {
  const { rows } = await pool.query<UserRow>(
    `SELECT u.id_usuario, u.id_negocio, n.nombre_negocio, u.nombre, u.password_hash, u.rol
       FROM usuarios u
       JOIN negocios n ON n.id_negocio = u.id_negocio
      WHERE u.email = $1`,
    [email],
  );

  // Same error for "unknown email" and "wrong password" — don't leak which
  // emails are registered.
  const user = rows[0];
  if (!user || !(await bcrypt.compare(password, user.password_hash))) {
    throw new ApiError(401, 'invalid credentials');
  }
  return issueTokens(pool, user, device);
}

export async function refresh(refreshToken: string) {
  const client = await pool.connect();
  let transactionOpen = false;
  try {
    await client.query('BEGIN');
    transactionOpen = true;

    // DELETE ... RETURNING consume el token en un único paso. Dos refresh
    // concurrentes no pueden obtener dos pares nuevos a partir del mismo token.
    const { rows } = await client.query(
      `DELETE FROM refresh_tokens rt
       USING usuarios u, negocios n
       WHERE rt.token_hash = $1
         AND u.id_usuario = rt.id_usuario
         AND n.id_negocio = u.id_negocio
       RETURNING rt.dispositivo, rt.expira_en, u.id_usuario, u.id_negocio,
                 n.nombre_negocio, u.nombre, u.rol`,
      [hashToken(refreshToken)],
    );
    const row = rows[0];
    if (!row) throw new ApiError(401, 'unknown refresh token');

    if (new Date(row.expira_en).getTime() < Date.now()) {
      await client.query('COMMIT');
      transactionOpen = false;
      throw new ApiError(401, 'session expired, please log in again');
    }

    const tokens = await issueTokens(client, row, row.dispositivo ?? '');
    await client.query('COMMIT');
    transactionOpen = false;
    return tokens;
  } catch (error) {
    if (transactionOpen) await client.query('ROLLBACK');
    throw error;
  } finally {
    client.release();
  }
}

export async function logout(refreshToken: string) {
  await pool.query(`DELETE FROM refresh_tokens WHERE token_hash = $1`, [hashToken(refreshToken)]);
}
