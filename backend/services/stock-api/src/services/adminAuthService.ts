import bcrypt from 'bcryptjs';
import QRCode from 'qrcode';
import { generateSecret, generateURI, verifySync } from 'otplib';
import { pool } from '../db/pool.js';
import { ApiError } from '../errors/ApiError.js';
import {
  signAdminAccessToken,
  signAdminMfaToken,
  verifyAdminMfaToken,
} from './tokenService.js';

export interface InicioLoginAdminResultado {
  requiere_2fa: true;
  setup_totp: boolean;
  qr_code?: string;
  secret?: string;
  temp_token: string;
}

export interface LoginAdminCompletoResultado {
  token: string;
  admin: {
    id_admin: number;
    usuario: string;
    nombre: string;
    email: string;
  };
}

export async function iniciarLoginAdmin(usuario: string, contrasena: string): Promise<InicioLoginAdminResultado> {
  const usr = usuario.trim();
  const pass = contrasena.trim();

  if (!usr || !pass) {
    throw new ApiError(400, 'El usuario y la contraseña son obligatorios');
  }

  const res = await pool.query(
    `SELECT id_admin, usuario, COALESCE(email, '') AS email, password_hash, nombre,
            totp_secret, COALESCE(totp_activado, FALSE) AS totp_activado
       FROM super_admins
      WHERE LOWER(usuario) = LOWER($1)`,
    [usr],
  );

  if (res.rowCount === 0) {
    throw new ApiError(401, 'Usuario o contraseña de administrador incorrectos');
  }

  const fila = res.rows[0];
  const passValido = await bcrypt.compare(pass, fila.password_hash);
  if (!passValido) {
    throw new ApiError(401, 'Usuario o contraseña de administrador incorrectos');
  }

  const totpActivo = fila.totp_activado && fila.totp_secret;

  if (!totpActivo) {
    // Primer inicio de sesión o configuración inicial de 2FA
    const nuevoSecret = generateSecret();
    const uri = generateURI({
      issuer: 'StockAPP Admin',
      label: fila.usuario,
      secret: nuevoSecret,
    });

    const qrCodeDataUrl = await QRCode.toDataURL(uri, {
      margin: 2,
      width: 240,
      color: {
        dark: '#16242d',
        light: '#ffffff',
      },
    });

    const tempToken = signAdminMfaToken({
      id_admin: fila.id_admin,
      setup: true,
      secret: nuevoSecret,
    });

    return {
      requiere_2fa: true,
      setup_totp: true,
      qr_code: qrCodeDataUrl,
      secret: nuevoSecret,
      temp_token: tempToken,
    };
  }

  // Ya tiene 2FA configurado y activo
  const tempToken = signAdminMfaToken({
    id_admin: fila.id_admin,
    setup: false,
  });

  return {
    requiere_2fa: true,
    setup_totp: false,
    temp_token: tempToken,
  };
}

export async function verificarTotpAdmin(tempToken: string, codigo: string): Promise<LoginAdminCompletoResultado> {
  const cod = (codigo || '').trim().replace(/\s+/g, '');
  if (!cod || cod.length !== 6 || !/^\d{6}$/.test(cod)) {
    throw new ApiError(400, 'El código debe tener 6 dígitos numéricos');
  }

  let datosMfa: { sub: number; setup?: boolean; secret?: string };
  try {
    datosMfa = verifyAdminMfaToken(tempToken);
  } catch {
    throw new ApiError(401, 'La sesión de verificación ha expirado. Por favor, vuelve a iniciar sesión.');
  }

  const idAdmin = datosMfa.sub;

  const res = await pool.query(
    `SELECT id_admin, usuario, COALESCE(email, '') AS email, nombre,
            totp_secret, COALESCE(totp_activado, FALSE) AS totp_activado
       FROM super_admins
      WHERE id_admin = $1`,
    [idAdmin],
  );

  if (res.rowCount === 0) {
    throw new ApiError(401, 'Administrador no encontrado');
  }

  const admin = res.rows[0];
  let secretAValidar = '';

  if (datosMfa.setup) {
    if (!datosMfa.secret) {
      throw new ApiError(400, 'Secreto de configuración no encontrado');
    }
    secretAValidar = datosMfa.secret;
  } else {
    if (!admin.totp_secret) {
      throw new ApiError(400, 'El administrador no tiene configurado 2FA');
    }
    secretAValidar = admin.totp_secret;
  }

  // Verificar código TOTP con tolerancia de 30s antes/después
  const verificado = verifySync({
    token: cod,
    secret: secretAValidar,
    epochTolerance: 30,
  });

  if (!verificado.valid) {
    throw new ApiError(401, 'Código de autenticación inválido o expirado. Verificá la hora en tu dispositivo.');
  }

  // Si estaba configurando por primera vez, persistir en base de datos
  if (datosMfa.setup) {
    await pool.query(
      `UPDATE super_admins
          SET totp_secret = $1,
              totp_activado = TRUE
        WHERE id_admin = $2`,
      [datosMfa.secret, idAdmin],
    );
  }

  // Generar JWT definitivo de SuperAdmin
  const token = signAdminAccessToken({
    id_admin: admin.id_admin,
    usuario: admin.usuario,
    nombre: admin.nombre,
    email: admin.email,
  });

  return {
    token,
    admin: {
      id_admin: admin.id_admin,
      usuario: admin.usuario,
      nombre: admin.nombre,
      email: admin.email,
    },
  };
}
