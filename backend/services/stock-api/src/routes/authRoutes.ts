import { Router } from 'express';
import * as authController from '../controllers/authController.js';

export const authRoutes = Router();

authRoutes.post('/login', authController.login);
authRoutes.post('/refresh', authController.refresh);
authRoutes.post('/logout', authController.logout);

// Obtener Auth Key de Tailscale previa verificación de email/credenciales y control de 4 dispositivos
authRoutes.post('/obtener-tailscale-key', authController.solicitarTailscaleToken);

// Gestión de dispositivos vinculados (para front-admin o ajustes)
authRoutes.get('/negocios/:id_negocio/dispositivos', authController.listarDispositivos);
authRoutes.delete('/negocios/:id_negocio/dispositivos/:id_dispositivo', authController.eliminarDispositivo);
