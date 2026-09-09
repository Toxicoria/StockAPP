import type { NegocioAdmin, CrearNegocioPayload, DispositivoCliente } from '../types/negocio';
import type { VersionDesktop, CrearVersionPayload, EditarVersionPayload } from '../types/version';
import { obtenerTokenAdmin, cerrarSesionAdmin } from './auth';

const API_BASE = import.meta.env.VITE_API_URL || '';

async function pedirApi<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const token = obtenerTokenAdmin();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers,
  });

  if (res.status === 401 || res.status === 403) {
    if (token && !endpoint.includes('/api/admin/login') && !endpoint.includes('/api/admin/totp')) {
      cerrarSesionAdmin();
      window.location.reload();
    }
  }

  const datos = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(datos.error || datos.mensaje || 'Error en la solicitud a la API');
  }

  return datos as T;
}

export interface RespuestaInicioLoginAdmin {
  requiere_2fa: true;
  setup_totp: boolean;
  qr_code?: string;
  secret?: string;
  temp_token: string;
}

export interface RespuestaVerificacionAdmin {
  token: string;
  admin: {
    id_admin: number;
    usuario: string;
    nombre: string;
    email: string;
  };
}

export async function loginSuperAdmin(usuario: string, password: string): Promise<RespuestaInicioLoginAdmin> {
  return pedirApi<RespuestaInicioLoginAdmin>('/api/admin/login', {
    method: 'POST',
    body: JSON.stringify({ usuario, password }),
  });
}

export async function verificarTotpAdmin(tempToken: string, codigo: string): Promise<RespuestaVerificacionAdmin> {
  return pedirApi<RespuestaVerificacionAdmin>('/api/admin/totp/verificar', {
    method: 'POST',
    body: JSON.stringify({ temp_token: tempToken, codigo }),
  });
}

export async function obtenerNegocios(): Promise<NegocioAdmin[]> {
  return pedirApi<NegocioAdmin[]>('/api/admin/negocios');
}

export async function crearNegocioCliente(payload: CrearNegocioPayload): Promise<{
  id_negocio: number;
  id_usuario: number;
  nombre_negocio: string;
  usuario: string;
  nombre_dueno: string;
}> {
  return pedirApi('/api/admin/negocios', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export async function cambiarPasswordUsuario(idUsuario: number, password: string): Promise<{ mensaje: string }> {
  return pedirApi(`/api/admin/usuarios/${idUsuario}/password`, {
    method: 'PUT',
    body: JSON.stringify({ password }),
  });
}

export async function actualizarNegocio(
  idNegocio: number,
  payload: {
    nombre_negocio?: string;
    direccion?: string;
    cuit?: string;
    telefono?: string;
    email_negocio?: string;
    nombre_dueno?: string;
    usuario?: string;
    email_usuario?: string;
    password?: string;
  },
): Promise<{ mensaje: string }> {
  return pedirApi(`/api/admin/negocios/${idNegocio}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  });
}

export async function obtenerDispositivosNegocio(idNegocio: number): Promise<DispositivoCliente[]> {
  return pedirApi<DispositivoCliente[]>(`/api/negocios/${idNegocio}/dispositivos`);
}

export async function desvincularDispositivo(idNegocio: number, idDispositivo: number): Promise<{ ok: boolean }> {
  return pedirApi<{ ok: boolean }>(`/api/negocios/${idNegocio}/dispositivos/${idDispositivo}`, {
    method: 'DELETE',
  });
}

// ==============================================================================
// 💻 CONTROL DE VERSIONES DESKTOP
// ==============================================================================
export async function obtenerVersiones(): Promise<VersionDesktop[]> {
  const data = await pedirApi<{ versiones: VersionDesktop[] }>('/api/admin/versiones');
  return data.versiones || [];
}

export async function crearVersion(payload: CrearVersionPayload): Promise<{
  id: number;
  version: string;
  mensaje: string;
}> {
  return pedirApi('/api/admin/versiones', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export async function actualizarVersion(
  id: number,
  payload: EditarVersionPayload,
): Promise<{ mensaje: string }> {
  return pedirApi(`/api/admin/versiones/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  });
}

export async function eliminarVersion(id: number): Promise<{ mensaje: string }> {
  return pedirApi(`/api/admin/versiones/${id}`, {
    method: 'DELETE',
  });
}

