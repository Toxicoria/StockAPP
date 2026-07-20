import { Router } from 'express';
import * as proxyController from '../controllers/proxyController.js';

export const proxyRoutes = Router();

// Catch-all: everything not handled above goes to stock-operations.
proxyRoutes.all('/{*resto}', proxyController.forwardToStockOperations);
