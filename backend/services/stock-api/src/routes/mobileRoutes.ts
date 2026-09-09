import { Router, type Request, type Response } from 'express';
import { ApiError } from '../errors/ApiError.js';
import { verifyAccessToken } from '../services/tokenService.js';
import {
  generarQRVinculacion,
  canjearTokenMovil,
  obtenerResumenEnVivoMovil,
} from '../services/qrMobileService.js';
import { renderMobileHtml } from '../views/mobileHtml.js';

export const mobileRoutes = Router();

// 1. Servir la interfaz web móvil de supervisión
mobileRoutes.get('/movil', (_req: Request, res: Response) => {
  res.setHeader('Content-Type', 'text/html; charset=utf-8');
  res.send(renderMobileHtml());
});

interface ClaimsPayload {
  sub: number;
  negocio: number;
  nombre: string;
  rol: string;
  permisos: string[];
}

// Middleware auxiliar para verificar token de acceso en rutas móviles
function extraerClaims(req: Request): ClaimsPayload {
  const auth = req.headers.authorization;
  if (!auth || !auth.startsWith('Bearer ')) {
    throw new ApiError(401, 'Falta cabecera de autenticación Bearer');
  }
  const token = auth.slice(7);
  try {
    return verifyAccessToken(token) as unknown as ClaimsPayload;
  } catch {
    throw new ApiError(401, 'Token de sesión inválido o expirado');
  }
}

// 2. Generar el código QR de vinculación móvil (solo para el dueño en la PC)
mobileRoutes.post('/api/movil/generar-qr', async (req: Request, res: Response) => {
  const claims = extraerClaims(req);
  if (claims.rol !== 'dueño') {
    throw new ApiError(403, 'Solo el dueño puede generar el código QR de supervisión móvil');
  }

  const host = (req.headers['x-forwarded-host'] as string) || req.get('host') || 'localhost:3000';
  const resultado = await generarQRVinculacion(claims.sub, claims.negocio, claims.rol, host);
  res.json(resultado);
});

// 3. Canjear el token del QR desde el celular del dueño
mobileRoutes.post('/api/movil/canjear', async (req: Request, res: Response) => {
  const { token, device_id, nombre_dispositivo, tipo_dispositivo } = req.body ?? {};
  if (!token) {
    throw new ApiError(400, 'El token de vinculación es requerido');
  }

  const sesionMovil = await canjearTokenMovil(token, {
    device_id,
    nombre_dispositivo,
    tipo_dispositivo: tipo_dispositivo || 'mobile',
  });

  res.json(sesionMovil);
});

// 4. Obtener las métricas y balance en tiempo real para el celular
mobileRoutes.get('/api/movil/resumen', async (req: Request, res: Response) => {
  const claims = extraerClaims(req);
  const resumen = await obtenerResumenEnVivoMovil(claims.negocio);
  res.json(resumen);
});
