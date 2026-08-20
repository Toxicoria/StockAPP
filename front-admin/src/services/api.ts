import { obtenerSesionAdmin } from './auth';
import type { NegocioAdmin, CrearNegocioPayload } from '../types/negocio';

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:3000';

async function pedirApi<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const sesion = obtenerSesionAdmin();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };

  if (sesion && sesion.token) {
    headers['Authorization'] = `Bearer ${sesion.token}`;
  }

  // Si hay headers en options, los mergeamos, asegurando que sean un objeto simple
  if (options.headers) {
    Object.assign(headers, options.headers);
  }

  const res = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers,
  });

  const datos = await res.json();
  if (!res.ok) {
    throw new Error(datos.mensaje || datos.error || 'Error en la solicitud a la API');
  }

  return datos as T;
}

export async function loginSuperAdmin(usuario: string, password: string): Promise<{
  id_admin: number;
  usuario: string;
  nombre: string;
  email: string;
  token: string;
}> {
  return pedirApi('/api/admin/login', {
    method: 'POST',
    body: JSON.stringify({ usuario, password }),
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
  }
): Promise<{ mensaje: string }> {
  return pedirApi(`/api/admin/negocios/${idNegocio}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  });
}
