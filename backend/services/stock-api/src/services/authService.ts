import bcrypt from 'bcryptjs';
import { pool } from '../db/pool.js';
import { ApiError } from '../errors/ApiError.js';
import { hashToken, newRefreshToken, signAccessToken } from './tokenService.js';

interface UserRow {
  id_usuario: number;
  id_negocio: number;
  nombre: string;
  password_hash: string;
  rol: string;
}

async function issueTokens(user: Omit<UserRow, 'password_hash'>, device: string) {
  const access_token = signAccessToken(user);
  const { token: refresh_token, expiresAt } = newRefreshToken();

  await pool.query(
    `INSERT INTO refresh_tokens (id_usuario, token_hash, dispositivo, expira_en)
     VALUES ($1, $2, $3, $4)`,
    [user.id_usuario, hashToken(refresh_token), device, expiresAt],
  );

  return { access_token, refresh_token, nombre: user.nombre, rol: user.rol };
}

export async function login(email: string, password: string, device: string) {
  const { rows } = await pool.query<UserRow>(
    `SELECT id_usuario, id_negocio, nombre, password_hash, rol
       FROM usuarios
      WHERE email = $1`,
    [email],
  );

  // Same error for "unknown email" and "wrong password" — don't leak which
  // emails are registered.
  const user = rows[0];
  if (!user || !(await bcrypt.compare(password, user.password_hash))) {
    throw new ApiError(401, 'invalid credentials');
  }
  return issueTokens(user, device);
}

export async function refresh(refreshToken: string) {
  const { rows } = await pool.query(
    `SELECT rt.id, rt.dispositivo, rt.expira_en, u.id_usuario, u.id_negocio, u.nombre, u.rol
       FROM refresh_tokens rt
       JOIN usuarios u ON u.id_usuario = rt.id_usuario
      WHERE rt.token_hash = $1`,
    [hashToken(refreshToken)],
  );
  const row = rows[0];
  if (!row) throw new ApiError(401, 'unknown refresh token');

  // Rotation: the presented token is always deleted, even if expired.
  await pool.query(`DELETE FROM refresh_tokens WHERE id = $1`, [row.id]);

  if (new Date(row.expira_en).getTime() < Date.now()) {
    throw new ApiError(401, 'session expired, please log in again');
  }
  return issueTokens(row, row.dispositivo ?? '');
}

export async function logout(refreshToken: string) {
  await pool.query(`DELETE FROM refresh_tokens WHERE token_hash = $1`, [hashToken(refreshToken)]);
}
