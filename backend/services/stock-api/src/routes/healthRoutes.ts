import { Router } from 'express';

export const healthRoutes = Router();

healthRoutes.get('/ping', (_req, res) => {
  res.json({ status: 'online', message: 'stock-api is up' });
});
