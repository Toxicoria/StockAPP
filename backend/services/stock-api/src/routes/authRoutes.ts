import { Router } from 'express';
import * as authController from '../controllers/authController.js';
import { loginRateLimit } from '../middleware/loginRateLimit.js';

export const authRoutes = Router();

authRoutes.post('/login', loginRateLimit, authController.login);
authRoutes.post('/refresh', authController.refresh);
authRoutes.post('/logout', authController.logout);
