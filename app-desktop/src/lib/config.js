// Base de la API. En desarrollo (.env.development) apunta directo al gateway Express
// en :3000; en producción se usa el sidecar de Tailscale en :9090.
export const BASE_API = import.meta.env.VITE_API_URL ?? 'http://localhost:3000';

/**
 * Obtiene o genera un UUID único y persistente para este dispositivo.
 */
export function obtenerDeviceId() {
  if (typeof window === 'undefined') return 'device-server-side';
  let deviceId = localStorage.getItem('stockapp_device_id');
  if (!deviceId) {
    deviceId = 'dev-' + crypto.randomUUID();
    localStorage.setItem('stockapp_device_id', deviceId);
  }
  return deviceId;
}

/**
 * Solicita al servidor público la Auth Key de Tailscale previa verificación de email y clave,
 * respetando el límite máximo de 4 dispositivos por cliente.
 */
export async function solicitarKeyTailscale(identificador, password, nombreDispositivo = '') {
  const deviceId = obtenerDeviceId();
  const nombreDev = nombreDispositivo || `PC Desktop (${navigator.platform || 'Local'})`;

  const res = await fetch(`${BASE_API}/api/auth/obtener-tailscale-key`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      email: identificador,
      password: password,
      device_id: deviceId,
      nombre_dispositivo: nombreDev,
      tipo_dispositivo: 'desktop',
    }),
  });

  const datos = await res.json();
  if (!res.ok) {
    throw new Error(datos.mensaje || datos.error || 'Error verificando la cuenta del cliente');
  }

  // Guardar la ts_auth_key entregada localmente
  if (datos.ts_auth_key) {
    localStorage.setItem('stockapp_ts_auth_key', datos.ts_auth_key);
  }

  return datos;
}
