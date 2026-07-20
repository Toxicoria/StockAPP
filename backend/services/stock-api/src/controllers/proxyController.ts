import type { Request, Response } from 'express';
import { ApiError } from '../errors/ApiError.js';
import { verifyAccessToken } from '../services/tokenService.js';
import * as proxyService from '../services/proxyService.js';

export async function forwardToStockOperations(req: Request, res: Response) {
  const token = req.headers.authorization?.replace(/^Bearer /, '');
  if (!token) throw new ApiError(401, 'missing Authorization: Bearer <token> header');
  try {
    verifyAccessToken(token);
  } catch {
    throw new ApiError(401, 'invalid or expired token');
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
