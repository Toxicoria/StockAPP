import { Router } from 'express';
import { registrarClienteSSE, contarClientesConectados } from '../services/sseService.js';

export const desktopRoutes = Router();

/**
 * GET /api/desktop/stream
 * Canal SSE pasivo para recepción de eventos en tiempo real desde terminales de escritorio.
 */
desktopRoutes.get('/desktop/stream', (req, res) => {
  registrarClienteSSE(res);
});

/**
 * GET /api/desktop/stream/status
 * Métricas de terminales conectadas.
 */
desktopRoutes.get('/desktop/stream/status', (req, res) => {
  res.json({ conectados: contarClientesConectados() });
});
