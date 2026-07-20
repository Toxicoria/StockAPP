import type { Request, Response } from 'express';
import { ApiError } from '../errors/ApiError.js';
import * as authService from '../services/authService.js';

export async function login(req: Request, res: Response) {
  const { email, password, dispositivo = '' } = req.body ?? {};
  if (!email || !password) throw new ApiError(400, 'email and password are required');
  res.json(await authService.login(email, password, dispositivo));
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
