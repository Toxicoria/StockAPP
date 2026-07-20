import type { NextFunction, Request, Response } from 'express';
import { ApiError } from '../errors/ApiError.js';

export function notFound(_req: Request, res: Response) {
  res.status(404).json({ error: 'route not found' });
}

// Mounted last: Express 5 forwards both sync throws and rejected promises
// from route handlers here automatically, no per-route try/catch needed.
export function errorHandler(err: unknown, req: Request, res: Response, _next: NextFunction) {
  if (err instanceof ApiError) {
    res.status(err.status).json({ error: err.message });
    return;
  }
  console.error(`[error] ${req.method} ${req.originalUrl}:`, err);
  res.status(500).json({ error: 'internal server error' });
}
