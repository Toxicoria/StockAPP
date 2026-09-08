import { Router } from 'express';
import * as adminAuthController from '../controllers/adminAuthController.js';

export const adminAuthRoutes = Router();

adminAuthRoutes.post('/admin/login', adminAuthController.login);
adminAuthRoutes.post('/admin/totp/verificar', adminAuthController.verificarTotp);
