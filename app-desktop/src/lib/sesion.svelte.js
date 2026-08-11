// Estado de sesión con runas. El access token vive solo en memoria (dura 15
// minutos); el refresh token se persiste en localStorage para que la app
// recuerde la sesión entre reinicios.
import { BASE_API } from '$lib/config.js';

const CLAVE_REFRESH = 'refresh_token';
const CLAVE_USUARIOS_RECIENTES = 'usuarios_recientes';

export const sesion = $state({
  nombre: '',
  rol: '',
  accessToken: '',
  negocio: '', // nombre del negocio, para la barra de título
});

export function esDueno() {
  return sesion.rol === 'dueño';
}

export function haySesion() {
  return sesion.accessToken !== '';
}

/**
 * @returns {Array<{ email: string, nombre: string, rol: string, password?: string, recordarPassword?: boolean, ultimaSesion: number }>}
 */
export function obtenerUsuariosRecientes() {
  try {
    const raw = localStorage.getItem(CLAVE_USUARIOS_RECIENTES);
    if (!raw) return [];
    const lista = JSON.parse(raw);
    return Array.isArray(lista) ? lista : [];
  } catch {
    return [];
  }
}

/**
 * @param {{ email: string, nombre: string, rol: string, password?: string, recordarPassword?: boolean }} u
 */
export function guardarUsuarioReciente(u) {
  if (!u || !u.email) return;
  const lista = obtenerUsuariosRecientes().filter((item) => item.email !== u.email);
  lista.unshift({
    email: u.email,
    nombre: u.nombre || u.email.split('@')[0],
    rol: u.rol || 'cajero',
    password: u.recordarPassword ? (u.password || '') : '',
    recordarPassword: Boolean(u.recordarPassword),
    ultimaSesion: Date.now(),
  });
  // Mantener como máximo 6 usuarios recientes
  const recortada = lista.slice(0, 6);
  localStorage.setItem(CLAVE_USUARIOS_RECIENTES, JSON.stringify(recortada));
}

/**
 * @param {string} email
 */
export function eliminarUsuarioReciente(email) {
  const lista = obtenerUsuariosRecientes().filter((item) => item.email !== email);
  localStorage.setItem(CLAVE_USUARIOS_RECIENTES, JSON.stringify(lista));
}

/** @param {{ access_token: string, refresh_token: string, nombre: string, rol: string }} datos */
function guardarTokens(datos) {
  sesion.accessToken = datos.access_token;
  sesion.nombre = datos.nombre;
  sesion.rol = datos.rol;
  localStorage.setItem(CLAVE_REFRESH, datos.refresh_token);
}

function limpiar() {
  sesion.accessToken = '';
  sesion.nombre = '';
  sesion.rol = '';
  localStorage.removeItem(CLAVE_REFRESH);
}

/**
 * @param {string} email
 * @param {string} password
 * @param {boolean} [recordarPassword=true]
 */
export async function iniciarSesion(email, password, recordarPassword = true) {
  const resp = await fetch(`${BASE_API}/api/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password, dispositivo: 'app-escritorio' }),
  });
  const datos = await resp.json();
  if (!resp.ok) throw new Error(datos.error ?? 'no se pudo iniciar sesión');
  guardarTokens(datos);
  guardarUsuarioReciente({
    email,
    nombre: datos.nombre,
    rol: datos.rol,
    password,
    recordarPassword,
  });
}

// refrescar canjea el refresh token guardado por un par nuevo (rotación).
// Devuelve true si la sesión quedó viva. Single-flight: si dos pedidos
// reciben 401 a la vez, comparten el mismo refresh en curso.
let refrescando = null;

export function refrescar() {
  refrescando ??= (async () => {
    try {
      const refreshToken = localStorage.getItem(CLAVE_REFRESH);
      if (!refreshToken) return false;
      const resp = await fetch(`${BASE_API}/api/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });
      if (!resp.ok) {
        limpiar();
        return false;
      }
      guardarTokens(await resp.json());
      return true;
    } finally {
      refrescando = null;
    }
  })();
  return refrescando;
}

export async function cerrarSesion() {
  const refreshToken = localStorage.getItem(CLAVE_REFRESH);
  if (refreshToken) {
    // Best-effort: si el backend no responde, la sesión local se cierra igual.
    fetch(`${BASE_API}/api/logout`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    }).catch(() => {});
  }
  limpiar();
}
