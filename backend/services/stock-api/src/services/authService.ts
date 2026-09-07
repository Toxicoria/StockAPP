import bcrypt from 'bcryptjs';
import { pool } from '../db/pool.js';
import { ApiError } from '../errors/ApiError.js';
import { hashToken, newRefreshToken, signAccessToken } from './tokenService.js';
import { generarAuthKeyTailscale } from './tailscaleService.js';

interface UserRow {
  id_usuario: number;
  id_negocio: number;
  nombre: string;
  password_hash: string;
  rol: string;
  permisos: any;
}

async function issueTokens(user: Omit<UserRow, 'password_hash'>, device: string) {
  const access_token = signAccessToken(user);
  const { token: refresh_token, expiresAt } = newRefreshToken();

  await pool.query(
    `INSERT INTO refresh_tokens (id_usuario, token_hash, dispositivo, expira_en)
     VALUES ($1, $2, $3, $4)`,
    [user.id_usuario, hashToken(refresh_token), device, expiresAt],
  );

  return { access_token, refresh_token, nombre: user.nombre, rol: user.rol, permisos: user.permisos };
}

export async function login(identificador: string, password: string, device: string) {
  const { rows } = await pool.query<UserRow>(
    `SELECT id_usuario, id_negocio, nombre, password_hash, rol, COALESCE(permisos, '["vender", "stock"]'::jsonb) as permisos
       FROM usuarios
      WHERE LOWER(usuario) = LOWER($1) OR LOWER(email) = LOWER($1)`,
    [identificador],
  );

  const user = rows[0];
  if (!user || !(await bcrypt.compare(password, user.password_hash))) {
    throw new ApiError(401, 'usuario o contraseña incorrectos');
  }
  return issueTokens(user, device);
}

export async function refresh(refreshToken: string) {
  const { rows } = await pool.query(
    `SELECT rt.id, rt.dispositivo, rt.expira_en, u.id_usuario, u.id_negocio, u.nombre, u.rol, COALESCE(u.permisos, '["vender", "stock"]'::jsonb) as permisos
       FROM refresh_tokens rt
       JOIN usuarios u ON u.id_usuario = rt.id_usuario
      WHERE rt.token_hash = $1`,
    [hashToken(refreshToken)],
  );
  const row = rows[0];
  if (!row) throw new ApiError(401, 'unknown refresh token');

  await pool.query(`DELETE FROM refresh_tokens WHERE id = $1`, [row.id]);

  if (new Date(row.expira_en).getTime() < Date.now()) {
    throw new ApiError(401, 'session expired, please log in again');
  }
  return issueTokens(row, row.dispositivo ?? '');
}

export async function logout(refreshToken: string) {
  await pool.query(`DELETE FROM refresh_tokens WHERE token_hash = $1`, [hashToken(refreshToken)]);
}

export async function obtenerTailscaleKey(
  identificador: string,
  password: string,
  deviceId: string,
  nombreDispositivo: string = 'Dispositivo Desktop',
  tipoDispositivo: string = 'desktop',
) {
  const { rows } = await pool.query<
    UserRow & { nombre_negocio: string; max_dispositivos: number; ts_auth_key: string }
  >(
    `SELECT u.id_usuario, u.id_negocio, u.nombre, u.password_hash, u.rol,
            n.nombre_negocio, COALESCE(n.max_dispositivos, 4) as max_dispositivos, n.ts_auth_key
       FROM usuarios u
       JOIN negocios n ON n.id_negocio = u.id_negocio
      WHERE LOWER(u.usuario) = LOWER($1) OR LOWER(u.email) = LOWER($1)`,
    [identificador],
  );

  const user = rows[0];
  if (!user || !(await bcrypt.compare(password, user.password_hash))) {
    throw new ApiError(401, 'Usuario o contraseña incorrectos');
  }

  // Si el negocio aún no tiene un ts_auth_key generado, se genera dinámicamente
  let tsAuthKey = user.ts_auth_key;
  if (!tsAuthKey) {
    tsAuthKey = await generarAuthKeyTailscale(user.nombre_negocio);
    await pool.query(`UPDATE negocios SET ts_auth_key = $1 WHERE id_negocio = $2`, [tsAuthKey, user.id_negocio]);
  }

  // Verificar el dispositivo en la tabla dispositivos_cliente
  const devId = deviceId || `dev-${user.id_usuario}-${Date.now()}`;

  const { rows: devRows } = await pool.query(
    `SELECT id_dispositivo, activo FROM dispositivos_cliente WHERE id_negocio = $1 AND device_id = $2`,
    [user.id_negocio, devId],
  );

  if (devRows.length > 0) {
    const dev = devRows[0];
    if (!dev.activo) {
      throw new ApiError(403, 'Este dispositivo ha sido desvinculado por el administrador');
    }
    // Actualizar última conexión
    await pool.query(
      `UPDATE dispositivos_cliente SET ultima_conexion = CURRENT_TIMESTAMP, nombre_dispositivo = $1 WHERE id_dispositivo = $2`,
      [nombreDispositivo, dev.id_dispositivo],
    );
  } else {
    // Es un dispositivo nuevo: verificar cuántos hay activos
    const { rows: countRows } = await pool.query<{ count: string }>(
      `SELECT COUNT(*) as count FROM dispositivos_cliente WHERE id_negocio = $1 AND activo = true`,
      [user.id_negocio],
    );

    const activeCount = parseInt(countRows[0]?.count || '0', 10);
    if (activeCount >= user.max_dispositivos) {
      throw new ApiError(
        403,
        `Límite de ${user.max_dispositivos} dispositivos alcanzado para este negocio. Desvincula un equipo desde el panel de administración para ingresar con uno nuevo.`,
      );
    }

    // Registrar nuevo dispositivo
    await pool.query(
      `INSERT INTO dispositivos_cliente (id_negocio, device_id, nombre_dispositivo, tipo_dispositivo)
       VALUES ($1, $2, $3, $4)`,
      [user.id_negocio, devId, nombreDispositivo, tipoDispositivo],
    );
  }

  return {
    ok: true,
    ts_auth_key: tsAuthKey,
    destino_api: process.env.DESTINO_API || 'http://stock-server-api:3000',
    negocio: user.nombre_negocio,
    id_negocio: user.id_negocio,
  };
}

export async function obtenerDispositivosNegocio(idNegocio: number) {
  const { rows } = await pool.query(
    `SELECT id_dispositivo, device_id, nombre_dispositivo, tipo_dispositivo, fecha_registro, ultima_conexion, activo
       FROM dispositivos_cliente
      WHERE id_negocio = $1
      ORDER BY fecha_registro DESC`,
    [idNegocio],
  );
  return rows;
}

export async function desvincularDispositivo(idNegocio: number, idDispositivo: number) {
  await pool.query(`DELETE FROM dispositivos_cliente WHERE id_negocio = $1 AND id_dispositivo = $2`, [
    idNegocio,
    idDispositivo,
  ]);
  return { ok: true };
}
