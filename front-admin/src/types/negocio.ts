export interface NegocioAdmin {
  id_negocio: number;
  nombre_negocio: string;
  direccion: string;
  cuit: string;
  nombre_dueno: string;
  telefono: string;
  email_negocio: string;
  fecha_alta: string;
  id_dueno: number;
  nombre_usuario: string;
  usuario: string;
  email_usuario: string;
  perfil_completo: boolean;
  max_dispositivos?: number;
  ts_auth_key?: string;
}

export interface CrearNegocioPayload {
  nombre_negocio: string;
  usuario: string;
  email?: string;
  password: string;
  nombre_dueno?: string;
}

export interface DispositivoCliente {
  id_dispositivo: number;
  device_id: string;
  nombre_dispositivo: string;
  tipo_dispositivo: string;
  fecha_registro: string;
  ultima_conexion: string;
  activo: boolean;
}
