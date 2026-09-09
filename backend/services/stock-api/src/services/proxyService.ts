import { env } from '../config/env.js';

const STOCK_OPERATIONS_URL = env('GO_INTERNO', 'http://localhost:8080');

export function forward(
  method: string,
  path: string,
  authHeader: string,
  body: unknown,
  extraHeaders?: Record<string, string>
): Promise<globalThis.Response> {
  const hasBody = !['GET', 'HEAD'].includes(method);
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    Authorization: authHeader,
    ...extraHeaders,
  };
  return fetch(`${STOCK_OPERATIONS_URL}${path}`, {
    method,
    headers,
    body: hasBody ? JSON.stringify(body ?? {}) : undefined,
  });
}
