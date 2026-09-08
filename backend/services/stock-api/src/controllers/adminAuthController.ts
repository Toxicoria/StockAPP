import type { Request, Response } from 'express';
import * as adminAuthService from '../services/adminAuthService.js';

export async function login(req: Request, res: Response) {
  const { usuario, password } = req.body;
  const resultado = await adminAuthService.iniciarLoginAdmin(usuario, password);
  res.json(resultado);
}

export async function verificarTotp(req: Request, res: Response) {
  const { temp_token, codigo } = req.body;
  const resultado = await adminAuthService.verificarTotpAdmin(temp_token, codigo);
  res.json(resultado);
}
