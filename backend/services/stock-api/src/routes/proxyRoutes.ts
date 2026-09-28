import { Router } from 'express';
import * as proxyController from '../controllers/proxyController.js';

export const proxyRoutes = Router();

// Primer arranque: todavía no existe un usuario capaz de obtener un JWT.
proxyRoutes.get('/registro', proxyController.forwardRegistration);
proxyRoutes.post('/registro', proxyController.forwardRegistration);

// Catch-all: everything not handled above goes to stock-operations.
proxyRoutes.all('/{*resto}', proxyController.forwardToStockOperations);
