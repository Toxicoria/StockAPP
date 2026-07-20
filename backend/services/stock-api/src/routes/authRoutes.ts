import { Router } from 'express';
import * as authController from '../controllers/authController.js';

export const authRoutes = Router();

authRoutes.post('/login', authController.login);
authRoutes.post('/refresh', authController.refresh);
authRoutes.post('/logout', authController.logout);
