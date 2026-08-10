// Estado de sesión con runas. El access token vive solo en memoria (dura 15
// minutos); el refresh token se persiste en localStorage para que la app
// recuerde la sesión entre reinicios.
import { BASE_API } from '$lib/config.js';

const CLAVE_REFRESH = 'refresh_token';

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
 */
export async function iniciarSesion(email, password) {
  const resp = await fetch(`${BASE_API}/api/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password, dispositivo: 'app-escritorio' }),
  });
  const datos = await resp.json();
  if (!resp.ok) throw new Error(datos.error ?? 'no se pudo iniciar sesión');
  guardarTokens(datos);
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
