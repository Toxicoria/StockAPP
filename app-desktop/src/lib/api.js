// pedirApi: fetch con Bearer y auto-refresh. Ante un 401 intenta refrescar
// el token una sola vez y reintenta; si el refresh falla, la sesión murió y
// el layout redirige al login.
import { goto } from '$app/navigation';
import { BASE_API } from '$lib/config.js';
import { sesion, refrescar } from '$lib/sesion.svelte.js';

export class ErrorApi extends Error {
  /**
   * @param {number} codigo
   * @param {string} mensaje
   */
  constructor(codigo, mensaje) {
    super(mensaje);
    this.codigo = codigo;
  }
}

/** @typedef {Omit<RequestInit, 'body'> & { body?: any }} OpcionesApi */

/**
 * @param {string} ruta
 * @param {OpcionesApi} opciones
 */
async function hacerPedido(ruta, opciones) {
  return fetch(`${BASE_API}${ruta}`, {
    ...opciones,
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${sesion.accessToken}`,
      ...opciones.headers,
    },
  });
}

/**
 * @param {string} ruta
 * @param {OpcionesApi} [opciones]
 * @returns {Promise<any>}
 */
export async function pedirApi(ruta, opciones = {}) {
  if (opciones.body && typeof opciones.body !== 'string') {
    opciones = { ...opciones, body: JSON.stringify(opciones.body) };
  }

  let resp = await hacerPedido(ruta, opciones);
  if (resp.status === 401) {
    if (await refrescar()) {
      resp = await hacerPedido(ruta, opciones);
    } else {
      goto('/login');
      throw new ErrorApi(401, 'sesión vencida');
    }
  }

  if (resp.status === 204) return null;
  const datos = await resp.json().catch(() => ({}));
  if (!resp.ok) throw new ErrorApi(resp.status, datos.error ?? `error ${resp.status}`);
  return datos;
}
