import { env } from '../config/env.js';

export function forward(method: string, path: string, authHeader: string, body: unknown): Promise<globalThis.Response> {
  const hasBody = !['GET', 'HEAD'].includes(method);
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (authHeader) headers.Authorization = authHeader;

  return fetch(`${env('GO_INTERNO', 'http://localhost:8080')}${path}`, {
    method,
    headers,
    body: hasBody ? JSON.stringify(body ?? {}) : undefined,
  });
}
