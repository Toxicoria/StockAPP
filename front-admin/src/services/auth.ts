export interface SuperAdminSesion {
  id_admin: number;
  usuario: string;
  nombre: string;
  email: string;
  token: string;
}

const STORAGE_KEY = 'stockapp_superadmin_sesion';

export function obtenerSesionAdmin(): SuperAdminSesion | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    return JSON.parse(raw) as SuperAdminSesion;
  } catch {
    return null;
  }
}

export function guardarSesionAdmin(sesion: SuperAdminSesion): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(sesion));
}

export function cerrarSesionAdmin(): void {
  localStorage.removeItem(STORAGE_KEY);
}
