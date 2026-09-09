export interface VersionDesktop {
  id: number;
  version: string;
  version_minima: string;
  es_obligatoria: boolean;
  canal: 'produccion' | 'beta';
  estado: 'borrador' | 'publicada' | 'desactivada';
  url_windows: string;
  firma_windows: string;
  url_linux: string;
  firma_linux: string;
  notas_version: string;
  motivo_obligatoria: string;
  publicada_en?: string | null;
  creada_en: string;
  actualizada_en: string;
}

export interface CrearVersionPayload {
  version: string;
  version_minima?: string;
  es_obligatoria?: boolean;
  canal?: string;
  estado?: string;
  url_windows?: string;
  firma_windows?: string;
  url_linux?: string;
  firma_linux?: string;
  notas_version?: string;
  motivo_obligatoria?: string;
}

export interface EditarVersionPayload {
  version_minima?: string;
  es_obligatoria?: boolean;
  canal?: string;
  estado?: string;
  url_windows?: string;
  firma_windows?: string;
  url_linux?: string;
  firma_linux?: string;
  notas_version?: string;
  motivo_obligatoria?: string;
}
