import express from 'express';
import { cors } from './middleware/cors.js';
import { securityHeaders } from './middleware/security.js';
import { errorHandler, notFound } from './middleware/errorHandler.js';
import { healthRoutes } from './routes/healthRoutes.js';
import { authRoutes } from './routes/authRoutes.js';
import { sessionRoutes } from './routes/sessionRoutes.js';
import { proxyRoutes } from './routes/proxyRoutes.js';

export function createApp() {
  const app = express();
  app.use(express.json());
  app.use(cors);
  app.use(securityHeaders);

  app.use('/api', healthRoutes);
  app.use('/api', authRoutes);
  app.use(sessionRoutes);
  app.use('/api', proxyRoutes);

  app.use(notFound);
  app.use(errorHandler);
  return app;
}
