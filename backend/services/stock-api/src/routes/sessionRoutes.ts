import { Router } from 'express';
import * as sessionController from '../controllers/sessionController.js';

export const sessionRoutes = Router();

sessionRoutes.get('/internal/session', sessionController.introspect);
