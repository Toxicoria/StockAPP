import type { Request, Response } from 'express';
import { ApiError } from '../errors/ApiError.js';
import { verifyAccessToken } from '../services/tokenService.js';

// Session introspection for stock-operations: never exposed outside the
// internal network (the sidecar only forwards /api/*, and the K8s Service
// stays cluster-internal).
export function introspect(req: Request, res: Response) {
  const token = req.headers.authorization?.replace(/^Bearer /, '');
  if (!token) throw new ApiError(401, 'missing Authorization: Bearer <token> header');
  try {
    res.json(verifyAccessToken(token));
  } catch {
    throw new ApiError(401, 'invalid or expired token');
  }
}
