import type { Request, Response } from 'express';
import { ApiError } from '../errors/ApiError.js';
import { verifyAccessToken } from '../services/tokenService.js';
import * as proxyService from '../services/proxyService.js';

export async function forwardToStockOperations(req: Request, res: Response) {
  const esRutaPublica =
    req.originalUrl.startsWith('/api/registro') ||
    req.originalUrl.startsWith('/api/ping');

  if (!esRutaPublica) {
    const token = req.headers.authorization?.replace(/^Bearer /, '');
    if (!token) throw new ApiError(401, 'missing Authorization: Bearer <token> header');
    let claims: any;
    try {
      claims = verifyAccessToken(token);
    } catch {
      throw new ApiError(401, 'invalid or expired token');
    }

    if (req.originalUrl.startsWith('/api/admin')) {
      if (claims?.rol !== 'superadmin') {
        throw new ApiError(403, 'acceso restringido a administradores');
      }
    }
  }

  let upstream: globalThis.Response;
  try {
    upstream = await proxyService.forward(req.method, req.originalUrl, req.headers.authorization ?? '', req.body);
  } catch (e) {
    console.error(`[proxy] stock-operations unreachable: ${e}`);
    throw new ApiError(502, 'upstream service unavailable');
  }

  res.status(upstream.status);
  if (upstream.status === 204) {
    res.end();
    return;
  }
  res.type(upstream.headers.get('content-type') ?? 'application/json');
  res.send(await upstream.text());
}
