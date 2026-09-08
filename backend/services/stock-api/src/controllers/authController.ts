import type { Request, Response } from 'express';
import { ApiError } from '../errors/ApiError.js';
import * as authService from '../services/authService.js';
import { verifyAccessToken } from '../services/tokenService.js';

export async function login(req: Request, res: Response) {
  const { usuario, email, identificador, password, dispositivo = '' } = req.body ?? {};
  const loginHandle = usuario || identificador || email;
  if (!loginHandle || !password) throw new ApiError(400, 'usuario y contraseña son obligatorios');
  res.json(await authService.login(loginHandle, password, dispositivo));
}

export async function refresh(req: Request, res: Response) {
  const refreshToken = req.body?.refresh_token;
  if (!refreshToken) throw new ApiError(400, 'refresh_token is required');
  res.json(await authService.refresh(refreshToken));
}

export async function logout(req: Request, res: Response) {
  const refreshToken = req.body?.refresh_token;
  if (!refreshToken) throw new ApiError(400, 'refresh_token is required');
  await authService.logout(refreshToken);
  res.status(204).end();
}

export async function solicitarTailscaleToken(req: Request, res: Response) {
  const { usuario, email, identificador, password, device_id, nombre_dispositivo, tipo_dispositivo } = req.body ?? {};
  const loginHandle = usuario || email || identificador;
  if (!loginHandle || !password) {
    throw new ApiError(400, 'El usuario/email y la contraseña son obligatorios');
  }
  const resultado = await authService.obtenerTailscaleKey(
    loginHandle,
    password,
    device_id,
    nombre_dispositivo,
    tipo_dispositivo,
  );
  res.json(resultado);
}

function autorizarGestionDispositivos(req: Request, idNegocio: number) {
  const token = req.headers.authorization?.replace(/^Bearer /, '');
  if (!token) throw new ApiError(401, 'missing Authorization: Bearer <token> header');
  let claims: any;
  try {
    claims = verifyAccessToken(token);
  } catch {
    throw new ApiError(401, 'invalid or expired token');
  }
  if (claims?.rol !== 'superadmin' && claims?.negocio !== idNegocio) {
    throw new ApiError(403, 'acceso no autorizado para este negocio');
  }
}

export async function listarDispositivos(req: Request, res: Response) {
  const rawId = Array.isArray(req.params.id_negocio) ? req.params.id_negocio[0] : req.params.id_negocio;
  const idNegocio = parseInt(rawId, 10);
  if (isNaN(idNegocio)) throw new ApiError(400, 'ID de negocio inválido');
  autorizarGestionDispositivos(req, idNegocio);
  res.json(await authService.obtenerDispositivosNegocio(idNegocio));
}

export async function eliminarDispositivo(req: Request, res: Response) {
  const rawNegocio = Array.isArray(req.params.id_negocio) ? req.params.id_negocio[0] : req.params.id_negocio;
  const rawDev = Array.isArray(req.params.id_dispositivo) ? req.params.id_dispositivo[0] : req.params.id_dispositivo;
  const idNegocio = parseInt(rawNegocio, 10);
  const idDispositivo = parseInt(rawDev, 10);
  if (isNaN(idNegocio) || isNaN(idDispositivo)) throw new ApiError(400, 'Parámetros inválidos');
  autorizarGestionDispositivos(req, idNegocio);
  res.json(await authService.desvincularDispositivo(idNegocio, idDispositivo));
}
