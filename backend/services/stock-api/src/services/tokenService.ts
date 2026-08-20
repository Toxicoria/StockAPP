import { createHash, randomBytes } from 'node:crypto';
import jwt from 'jsonwebtoken';
import { env } from '../config/env.js';

const ACCESS_TOKEN_TTL_SEC = 15 * 60;
const REFRESH_TOKEN_TTL_MS = 30 * 24 * 60 * 60 * 1000; // 30 days

interface TokenClaims {
  id_usuario: number;
  id_negocio: number;
  nombre: string;
  rol: string;
  permisos?: any;
}

export function jwtSecret(): string {
  return env('JWT_SECRET', 'secreto_solo_para_desarrollo');
}

export function hashToken(token: string): string {
  return createHash('sha256').update(token).digest('hex');
}

export function signAccessToken(claims: TokenClaims): string {
  return jwt.sign(
    {
      sub: claims.id_usuario,
      negocio: claims.id_negocio,
      nombre: claims.nombre,
      rol: claims.rol,
      permisos: claims.permisos ?? ['vender', 'stock'],
    },
    jwtSecret(),
    { algorithm: 'HS256', expiresIn: ACCESS_TOKEN_TTL_SEC },
  );
}

export function verifyAccessToken(token: string) {
  return jwt.verify(token, jwtSecret(), { algorithms: ['HS256'] });
}

export function newRefreshToken(): { token: string; expiresAt: Date } {
  return {
    token: randomBytes(32).toString('hex'),
    expiresAt: new Date(Date.now() + REFRESH_TOKEN_TTL_MS),
  };
}
