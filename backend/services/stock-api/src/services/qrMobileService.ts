import jwt from 'jsonwebtoken';
import QRCode from 'qrcode';
import { pool } from '../db/pool.js';
import { ApiError } from '../errors/ApiError.js';
import { jwtSecret } from './tokenService.js';
import { registrarOVerificarDispositivo, issueTokens, type DeviceInfo } from './authService.js';

const ZONA_HORARIA = 'America/Argentina/Buenos_Aires';

interface TokenPayload {
  action: string;
  id_usuario: number;
  id_negocio: number;
  rol: string;
}

export async function generarQRVinculacion(
  idUsuario: number,
  idNegocio: number,
  rol: string,
  host: string,
) {
  if (rol !== 'dueño') {
    throw new ApiError(403, 'Solo el dueño puede generar el código QR de supervisión móvil');
  }

  // Generar token efímero de vinculación móvil (15 min de vigencia)
  const pairingToken = jwt.sign(
    {
      action: 'mobile_supervisor',
      id_usuario: idUsuario,
      id_negocio: idNegocio,
      rol,
    },
    jwtSecret(),
    { expiresIn: '15m' },
  );

  // Construir la URL completa a la que apuntará el QR
  const hostLimpio = host.replace(/^https?:\/\//, '');
  const proto = host.startsWith('https') ? 'https' : 'http';
  const baseUrl = process.env.PUBLIC_URL || `${proto}://${hostLimpio}`;
  const mobileUrl = `${baseUrl}/movil?token=${pairingToken}`;

  // Generar QR en formato Data URL con los colores patagónicos de StockAPP
  const qrDataUrl = await QRCode.toDataURL(mobileUrl, {
    width: 320,
    margin: 2,
    color: {
      dark: '#2e6e73',
      light: '#ffffff',
    },
  });

  // Consultar cantidad de dispositivos actualmente activos
  const { rows: countRows } = await pool.query<{ count: string; max_dispositivos: number }>(
    `SELECT COUNT(dc.id_dispositivo) as count, COALESCE(n.max_dispositivos, 4) as max_dispositivos
       FROM negocios n
       LEFT JOIN dispositivos_cliente dc ON dc.id_negocio = n.id_negocio AND dc.activo = true
      WHERE n.id_negocio = $1
      GROUP BY n.max_dispositivos`,
    [idNegocio],
  );

  const activos = parseInt(countRows[0]?.count || '0', 10);
  const maximo = countRows[0]?.max_dispositivos || 4;

  return {
    qr_data_url: qrDataUrl,
    mobile_url: mobileUrl,
    dispositivos_activos: activos,
    max_dispositivos: maximo,
  };
}

export async function canjearTokenMovil(pairingToken: string, deviceInfo: DeviceInfo) {
  if (!pairingToken) {
    throw new ApiError(400, 'Falta el token de vinculación');
  }

  let payload: TokenPayload;
  try {
    payload = jwt.verify(pairingToken, jwtSecret()) as TokenPayload;
  } catch {
    throw new ApiError(401, 'El código QR ha expirado o no es válido. Generá uno nuevo desde la computadora.');
  }

  if (payload.action !== 'mobile_supervisor') {
    throw new ApiError(400, 'Acción no permitida');
  }

  // Obtener datos del dueño y negocio
  const { rows } = await pool.query(
    `SELECT u.id_usuario, u.id_negocio, u.nombre, u.rol,
            COALESCE(n.nombre_negocio, '') as nombre_negocio,
            COALESCE(n.direccion, '') as direccion,
            COALESCE(n.telefono, '') as telefono,
            COALESCE(n.nombre_dueno, '') as nombre_dueno,
            COALESCE(n.max_dispositivos, 4) as max_dispositivos,
            COALESCE(n.ts_auth_key, '') as ts_auth_key
       FROM usuarios u
       JOIN negocios n ON n.id_negocio = u.id_negocio
      WHERE u.id_usuario = $1 AND u.id_negocio = $2`,
    [payload.id_usuario, payload.id_negocio],
  );

  const user = rows[0];
  if (!user) {
    throw new ApiError(404, 'Usuario o negocio no encontrado');
  }

  // Asignar metadatos de terminal móvil
  const devId = deviceInfo.device_id?.trim() || `dev-mob-${user.id_usuario}-${Date.now()}`;
  const nombreMovil = deviceInfo.nombre_dispositivo?.trim() || 'Celular Dueño';

  // Registrar el celular en dispositivos_cliente (validando límite de 4 simultáneos)
  await registrarOVerificarDispositivo(user.id_negocio, user.max_dispositivos || 4, {
    device_id: devId,
    nombre_dispositivo: nombreMovil,
    tipo_dispositivo: 'mobile',
  });

  // Emitir sesión con permisos estrictamente de supervisión/lectura (NUNCA 'vender')
  const userMobile = {
    id_usuario: user.id_usuario,
    id_negocio: user.id_negocio,
    nombre: user.nombre,
    rol: 'dueño',
    permisos: ['resumen', 'stock', 'ventas_lectura'],
    nombre_negocio: user.nombre_negocio,
    direccion: user.direccion,
    telefono: user.telefono,
    nombre_dueno: user.nombre_dueno,
    max_dispositivos: user.max_dispositivos,
    ts_auth_key: user.ts_auth_key,
  };

  const tokens = await issueTokens(
    userMobile,
    `celular-${devId}`,
    { device_id: devId, nombre_dispositivo: nombreMovil, tipo_dispositivo: 'mobile' },
    user.ts_auth_key,
  );

  return {
    ...tokens,
    device_id: devId,
    nombre_dispositivo: nombreMovil,
  };
}

export async function obtenerResumenEnVivoMovil(idNegocio: number) {
  // 1. Total facturado hoy, cantidad de tickets y promedio
  const { rows: hoyRows } = await pool.query<{ total: string; cantidad: string }>(
    `SELECT COALESCE(SUM(total_venta), 0) as total, COUNT(*) as cantidad
       FROM ventas
      WHERE id_negocio = $1
        AND (fecha_hora AT TIME ZONE 'UTC' AT TIME ZONE '${ZONA_HORARIA}')::date =
            (now() AT TIME ZONE '${ZONA_HORARIA}')::date`,
    [idNegocio],
  );

  const hoyTotal = parseFloat(hoyRows[0]?.total || '0');
  const hoyCantidad = parseInt(hoyRows[0]?.cantidad || '0', 10);
  const hoyPromedio = hoyCantidad > 0 ? hoyTotal / hoyCantidad : 0;

  // 2. Desglose de pagos hoy
  const { rows: pagosRows } = await pool.query<{ metodo_pago: string; total: string; cantidad: string }>(
    `SELECT metodo_pago, COALESCE(SUM(total_venta), 0) as total, COUNT(*) as cantidad
       FROM ventas
      WHERE id_negocio = $1
        AND (fecha_hora AT TIME ZONE 'UTC' AT TIME ZONE '${ZONA_HORARIA}')::date =
            (now() AT TIME ZONE '${ZONA_HORARIA}')::date
      GROUP BY metodo_pago
      ORDER BY total DESC`,
    [idNegocio],
  );

  // 3. Últimas 10 ventas y quién está atendiendo ahora en el mostrador
  const { rows: ultimasVentas } = await pool.query<{
    id_venta: number;
    total_venta: string;
    metodo_pago: string;
    fecha_hora: string;
    cajero: string;
  }>(
    `SELECT v.id_venta, v.total_venta, v.metodo_pago, v.fecha_hora, u.nombre as cajero
       FROM ventas v
       JOIN usuarios u ON u.id_usuario = v.id_usuario
      WHERE v.id_negocio = $1
      ORDER BY v.fecha_hora DESC
      LIMIT 10`,
    [idNegocio],
  );

  const cajeroActivo = ultimasVentas[0]
    ? {
        nombre: ultimasVentas[0].cajero,
        ultima_venta_hora: ultimasVentas[0].fecha_hora,
        ultimo_monto: parseFloat(ultimasVentas[0].total_venta),
        ultimo_metodo: ultimasVentas[0].metodo_pago,
      }
    : null;

  // 4. Productos con stock crítico (en o por debajo del stock mínimo)
  const { rows: stockCritico } = await pool.query<{
    id_producto: string;
    descripcion: string;
    cantidad_disponible: string;
    stock_minimo: string;
  }>(
    `SELECT s.id_producto, COALESCE(p.productos_descripcion, s.id_producto) as descripcion,
            s.cantidad_disponible, s.stock_minimo
       FROM stock_interno s
       LEFT JOIN productos p ON p.id_producto = s.id_producto
      WHERE s.id_negocio = $1
        AND s.stock_minimo > 0
        AND s.cantidad_disponible <= s.stock_minimo
      ORDER BY s.cantidad_disponible ASC
      LIMIT 12`,
    [idNegocio],
  );

  // 5. Equipos actualmente conectados simultáneamente
  const { rows: equiposConectados } = await pool.query<{
    id_dispositivo: number;
    device_id: string;
    nombre_dispositivo: string;
    tipo_dispositivo: string;
    ultima_conexion: string;
  }>(
    `SELECT id_dispositivo, device_id, nombre_dispositivo, tipo_dispositivo, ultima_conexion
       FROM dispositivos_cliente
      WHERE id_negocio = $1 AND activo = true
      ORDER BY ultima_conexion DESC`,
    [idNegocio],
  );

  return {
    hoy: {
      total: hoyTotal,
      cantidad: hoyCantidad,
      promedio: hoyPromedio,
    },
    pagos: pagosRows.map((p) => ({
      metodo_pago: p.metodo_pago,
      total: parseFloat(p.total),
      cantidad: parseInt(p.cantidad, 10),
    })),
    cajero_activo: cajeroActivo,
    ultimas_ventas: ultimasVentas.map((v) => ({
      id_venta: v.id_venta,
      total_venta: parseFloat(v.total_venta),
      metodo_pago: v.metodo_pago,
      fecha_hora: v.fecha_hora,
      cajero: v.cajero,
    })),
    stock_critico: stockCritico.map((s) => ({
      id_producto: s.id_producto,
      descripcion: s.descripcion,
      cantidad_disponible: parseFloat(s.cantidad_disponible),
      stock_minimo: parseFloat(s.stock_minimo),
    })),
    dispositivos_conectados: equiposConectados,
    total_equipos_activos: equiposConectados.length,
    fecha_consulta: new Date().toISOString(),
  };
}
