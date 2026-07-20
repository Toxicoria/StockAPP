import type { NextFunction, Request, Response } from 'express';

// The API only ever returns JSON, but these headers are cheap insurance
// against a browser (or webview) misinterpreting a misconfigured response —
// relevant now that this service also holds an encrypted fiscal credential.
export function securityHeaders(_req: Request, res: Response, next: NextFunction) {
  res.set('X-Content-Type-Options', 'nosniff');
  res.set('X-Frame-Options', 'DENY');
  res.set('Referrer-Policy', 'no-referrer');
  next();
}
