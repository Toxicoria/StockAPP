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
}

export interface CrearNegocioPayload {
  nombre_negocio: string;
  usuario: string;
  password: string;
  nombre_dueno?: string;
}
