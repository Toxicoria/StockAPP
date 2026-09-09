<script setup lang="ts">
import { ref, computed } from 'vue';
import type { VersionDesktop } from '../types/version';
import { actualizarVersion, eliminarVersion } from '../services/api';

const props = defineProps<{
  versiones: VersionDesktop[];
  cargando: boolean;
}>();

const emit = defineEmits<{
  (e: 'abrir-modal-crear'): void;
  (e: 'abrir-modal-editar', v: VersionDesktop): void;
  (e: 'recargar'): void;
}>();

const filtroEstado = ref<string>('todos');
const busqueda = ref('');
const procesandoId = ref<number | null>(null);

// Versión activa: la más reciente con estado === 'publicada'
const versionActiva = computed(() => {
  return props.versiones.find((v) => v.estado === 'publicada') || null;
});

const totalPublicadas = computed(() => props.versiones.filter((v) => v.estado === 'publicada').length);
const totalBorradores = computed(() => props.versiones.filter((v) => v.estado === 'borrador').length);
const totalDesactivadas = computed(() => props.versiones.filter((v) => v.estado === 'desactivada').length);

const versionesFiltradas = computed(() => {
  return props.versiones.filter((v) => {
    if (filtroEstado.value !== 'todos' && v.estado !== filtroEstado.value) {
      return false;
    }
    if (busqueda.value.trim()) {
      const q = busqueda.value.toLowerCase().trim();
      const coincideVer = v.version.toLowerCase().includes(q);
      const coincideNotas = (v.notas_version || '').toLowerCase().includes(q);
      return coincideVer || coincideNotas;
    }
    return true;
  });
});

async function cambiarEstado(v: VersionDesktop, nuevoEstado: 'publicada' | 'desactivada') {
  if (procesandoId.value) return;
  procesandoId.value = v.id;
  try {
    await actualizarVersion(v.id, { estado: nuevoEstado });
    emit('recargar');
  } catch (e: any) {
    alert(e.message || 'Error al cambiar estado');
  } finally {
    procesandoId.value = null;
  }
}

async function alternarObligatoria(v: VersionDesktop) {
  if (procesandoId.value) return;
  procesandoId.value = v.id;
  try {
    await actualizarVersion(v.id, { es_obligatoria: !v.es_obligatoria });
    emit('recargar');
  } catch (e: any) {
    alert(e.message || 'Error al modificar obligatoriedad');
  } finally {
    procesandoId.value = null;
  }
}

async function confirmarEliminar(v: VersionDesktop) {
  if (!confirm(`¿Estás seguro de eliminar la versión v${v.version}? Esta acción no se puede deshacer.`)) {
    return;
  }
  procesandoId.value = v.id;
  try {
    await eliminarVersion(v.id);
    emit('recargar');
  } catch (e: any) {
    alert(e.message || 'Error al eliminar versión');
  } finally {
    procesandoId.value = null;
  }
}

function formatearFecha(f?: string | null): string {
  if (!f) return '—';
  try {
    const d = new Date(f);
    return d.toLocaleDateString('es-AR', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch {
    return f;
  }
}
</script>

<template>
  <div class="seccion-versiones">
    <!-- TARJETAS SUPERIORES DE RESUMEN DE DESPLIEGUE -->
    <div class="metrics-grid">
      <!-- TARJETA 1: VERSIÓN ACTIVA -->
      <div class="metric-card">
        <div class="metric-header">
          <span class="metric-title">Versión Activa en Producción</span>
          <div class="metric-icon active-icon">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"></path></svg>
          </div>
        </div>
        <div v-if="versionActiva" class="metric-content">
          <div class="metric-big">
            <span class="ver-tag">v{{ versionActiva.version }}</span>
            <span class="badge badge-success">Activa</span>
          </div>
          <p class="metric-desc">
            Publicada el {{ formatearFecha(versionActiva.publicada_en || versionActiva.creada_en) }}
          </p>
          <div class="plataformas-resumen">
            <span :class="['plat-badge', versionActiva.url_windows ? 'has-plat' : 'no-plat']">
              Windows {{ versionActiva.url_windows ? '✓' : '✗' }}
            </span>
            <span :class="['plat-badge', versionActiva.url_linux ? 'has-plat' : 'no-plat']">
              Linux {{ versionActiva.url_linux ? '✓' : '✗' }}
            </span>
          </div>
        </div>
        <div v-else class="metric-content">
          <div class="metric-big text-muted">Sin versión activa</div>
          <p class="metric-desc">Publica una versión para habilitar el auto-updater</p>
        </div>
      </div>

      <!-- TARJETA 2: POLÍTICA Y CONDICIONES DE DESPLIEGUE -->
      <div class="metric-card">
        <div class="metric-header">
          <span class="metric-title">Política de Despliegue Actual</span>
          <div class="metric-icon policy-icon">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path></svg>
          </div>
        </div>
        <div v-if="versionActiva" class="metric-content">
          <div class="policy-item">
            <span class="lbl-pol">Tipo de Actualización:</span>
            <span v-if="versionActiva.es_obligatoria" class="badge badge-warning">
              ⚠️ Esencial / Forzada
            </span>
            <span v-else class="badge badge-info">
              🟢 Opcional / Sugerida
            </span>
          </div>
          <div class="policy-item">
            <span class="lbl-pol">Versión Mínima Soportada:</span>
            <code class="code-min">v{{ versionActiva.version_minima }}</code>
          </div>
          <p v-if="versionActiva.es_obligatoria && versionActiva.motivo_obligatoria" class="policy-sub">
            Motivo: <em>"{{ versionActiva.motivo_obligatoria }}"</em>
          </p>
        </div>
        <div v-else class="metric-content">
          <p class="metric-desc">No hay versión activa configurada.</p>
        </div>
      </div>

      <!-- TARJETA 3: ESTADÍSTICAS TOTALES -->
      <div class="metric-card">
        <div class="metric-header">
          <span class="metric-title">Inventario de Versiones</span>
          <div class="metric-icon box-icon">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="m7.5 4.27 9 5.15"></path><path d="M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z"></path><path d="m3.3 7 8.7 5 8.7-5"></path><path d="M12 22V12"></path></svg>
          </div>
        </div>
        <div class="metric-content stats-column">
          <div class="stat-line">
            <span>Total registradas:</span>
            <strong>{{ versiones.length }}</strong>
          </div>
          <div class="stat-line text-success">
            <span>Publicadas:</span>
            <strong>{{ totalPublicadas }}</strong>
          </div>
          <div class="stat-line text-warning">
            <span>Borradores:</span>
            <strong>{{ totalBorradores }}</strong>
          </div>
          <div class="stat-line text-dim">
            <span>Desactivadas:</span>
            <strong>{{ totalDesactivadas }}</strong>
          </div>
        </div>
      </div>
    </div>

    <!-- BARRA DE HERRAMIENTAS Y FILTROS -->
    <div class="tabla-toolbar">
      <div class="toolbar-left">
        <div class="search-box">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
          <input
            v-model="busqueda"
            type="text"
            placeholder="Buscar por versión o notas..."
            class="search-input"
          />
        </div>

        <div class="tabs-filtro">
          <button
            :class="['tab-btn', { activo: filtroEstado === 'todos' }]"
            @click="filtroEstado = 'todos'"
          >
            Todas ({{ versiones.length }})
          </button>
          <button
            :class="['tab-btn', { activo: filtroEstado === 'publicada' }]"
            @click="filtroEstado = 'publicada'"
          >
            Publicadas ({{ totalPublicadas }})
          </button>
          <button
            :class="['tab-btn', { activo: filtroEstado === 'borrador' }]"
            @click="filtroEstado = 'borrador'"
          >
            Borradores ({{ totalBorradores }})
          </button>
          <button
            :class="['tab-btn', { activo: filtroEstado === 'desactivada' }]"
            @click="filtroEstado = 'desactivada'"
          >
            Desactivadas ({{ totalDesactivadas }})
          </button>
        </div>
      </div>

      <div class="toolbar-right">
        <button class="btn btn-secondary btn-sm" title="Refrescar tabla" @click="emit('recargar')">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"></polyline><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"></path></svg>
          <span>Refrescar</span>
        </button>
        <button class="btn btn-primary btn-sm" @click="emit('abrir-modal-crear')">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
          <span>Nueva Versión Manual</span>
        </button>
      </div>
    </div>

    <!-- TABLA DE VERSIONES -->
    <div class="tabla-container">
      <div v-if="cargando" class="estado-vacio">
        <div class="spinner-lg"></div>
        <p>Cargando versiones desde el servidor...</p>
      </div>

      <div v-else-if="versionesFiltradas.length === 0" class="estado-vacio">
        <div class="icon-vacio">
          <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect width="20" height="14" x="2" y="3" rx="2"></rect><line x1="8" x2="16" y1="21" y2="21"></line><line x1="12" x2="12" y1="17" y2="21"></line></svg>
        </div>
        <h3>No se encontraron versiones</h3>
        <p>No hay registros que coincidan con los filtros aplicados o aún no se ha registrado ninguna versión.</p>
        <button class="btn btn-secondary btn-sm" @click="emit('abrir-modal-crear')">
          Registrar primera versión
        </button>
      </div>

      <table v-else class="tabla">
        <thead>
          <tr>
            <th>Versión & Canal</th>
            <th>Tipo de Actualización</th>
            <th>Versión Mínima</th>
            <th>Plataformas</th>
            <th>Estado</th>
            <th>Fecha</th>
            <th class="th-acciones">Acciones</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="v in versionesFiltradas"
            :key="v.id"
            :class="{ 'fila-activa': versionActiva && versionActiva.id === v.id }"
          >
            <!-- VERSIÓN -->
            <td>
              <div class="celda-version">
                <div class="ver-numero">
                  <strong>v{{ v.version }}</strong>
                  <span v-if="versionActiva && versionActiva.id === v.id" class="badge-activa" title="Versión activa en producción">
                    Activa
                  </span>
                </div>
                <div class="sub-canal">
                  <span class="tag-canal">{{ v.canal }}</span>
                  <span v-if="v.notas_version" class="notas-resumen" :title="v.notas_version">
                    {{ v.notas_version.length > 40 ? v.notas_version.slice(0, 40) + '...' : v.notas_version }}
                  </span>
                </div>
              </div>
            </td>

            <!-- TIPO DE ACTUALIZACIÓN (ESENCIAL VS OPCIONAL) -->
            <td>
              <div class="celda-politica">
                <button
                  class="btn-toggle-politica"
                  :title="v.es_obligatoria ? 'Clic para cambiar a Opcional' : 'Clic para cambiar a Forzada / Esencial'"
                  :disabled="procesandoId === v.id"
                  @click="alternarObligatoria(v)"
                >
                  <span v-if="v.es_obligatoria" class="badge badge-warning badge-clickable">
                    ⚠️ Esencial (Forzada)
                  </span>
                  <span v-else class="badge badge-info badge-clickable">
                    🟢 Opcional (Sugerida)
                  </span>
                </button>
                <small v-if="v.motivo_obligatoria" class="motivo-sub" :title="v.motivo_obligatoria">
                  {{ v.motivo_obligatoria }}
                </small>
              </div>
            </td>

            <!-- VERSIÓN MÍNIMA -->
            <td>
              <code class="tag-version-min">v{{ v.version_minima }}</code>
            </td>

            <!-- PLATAFORMAS -->
            <td>
              <div class="iconos-plataformas">
                <!-- WINDOWS -->
                <span
                  :class="['plat-chip', v.url_windows ? 'plat-ok' : 'plat-missing']"
                  :title="v.url_windows ? 'Windows paquete disponible y firmado' : 'Sin paquete Windows'"
                >
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M0 3.449L9.75 2.1v9.451H0m10.949-9.602L24 0v11.4H10.949M0 12.6h9.75v9.451L0 20.699M10.949 12.6H24V24l-12.95-1.8"></path>
                  </svg>
                  <span>Win</span>
                </span>

                <!-- LINUX -->
                <span
                  :class="['plat-chip', v.url_linux ? 'plat-ok' : 'plat-missing']"
                  :title="v.url_linux ? 'Linux paquete disponible y firmado' : 'Sin paquete Linux'"
                >
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M12 2a4 4 0 0 0-4 4v5a4 4 0 0 0 8 0V6a4 4 0 0 0-4-4Z"></path>
                    <path d="M6 13a6 6 0 0 0 12 0"></path>
                    <line x1="8" x2="8" y1="19" y2="21"></line>
                    <line x1="16" x2="16" y1="19" y2="21"></line>
                  </svg>
                  <span>Linux</span>
                </span>
              </div>
            </td>

            <!-- ESTADO -->
            <td>
              <span v-if="v.estado === 'publicada'" class="badge badge-success">
                Publicada
              </span>
              <span v-else-if="v.estado === 'borrador'" class="badge badge-draft">
                Borrador
              </span>
              <span v-else class="badge badge-danger">
                Desactivada
              </span>
            </td>

            <!-- FECHA -->
            <td>
              <span class="fecha-txt">{{ formatearFecha(v.publicada_en || v.creada_en) }}</span>
            </td>

            <!-- ACCIONES -->
            <td class="td-acciones">
              <div class="acciones-grupo">
                <!-- BOTÓN PUBLICAR O PAUSAR -->
                <button
                  v-if="v.estado !== 'publicada'"
                  class="btn-icon btn-publicar"
                  title="Publicar para clientes de escritorio"
                  :disabled="procesandoId === v.id"
                  @click="cambiarEstado(v, 'publicada')"
                >
                  <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"></polygon></svg>
                </button>
                <button
                  v-else
                  class="btn-icon btn-pausar"
                  title="Pausar / Desactivar versión"
                  :disabled="procesandoId === v.id"
                  @click="cambiarEstado(v, 'desactivada')"
                >
                  <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="6" y="4" width="4" height="16"></rect><rect x="14" y="4" width="4" height="16"></rect></svg>
                </button>

                <!-- BOTÓN EDITAR -->
                <button
                  class="btn-icon"
                  title="Editar metadatos, firmas y notas"
                  @click="emit('abrir-modal-editar', v)"
                >
                  <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>
                </button>

                <!-- BOTÓN ELIMINAR -->
                <button
                  class="btn-icon btn-danger"
                  title="Eliminar versión"
                  :disabled="procesandoId === v.id"
                  @click="confirmarEliminar(v)"
                >
                  <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.seccion-versiones {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* METRICS */
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 16px;
}

.metric-card {
  background: var(--bg-surface);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 18px 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.metric-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.metric-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-muted);
}

.metric-icon {
  width: 34px;
  height: 34px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.active-icon {
  background: rgba(34, 197, 94, 0.15);
  color: var(--color-success);
}

.policy-icon {
  background: rgba(245, 158, 11, 0.15);
  color: var(--color-warning);
}

.box-icon {
  background: rgba(56, 189, 248, 0.15);
  color: var(--color-accent);
}

.metric-content {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.metric-big {
  display: flex;
  align-items: center;
  gap: 10px;
}

.ver-tag {
  font-size: 22px;
  font-weight: 700;
  color: var(--text-main);
  font-family: monospace;
}

.metric-desc {
  font-size: 12px;
  color: var(--text-dim);
}

.plataformas-resumen {
  display: flex;
  gap: 8px;
  margin-top: 4px;
}

.plat-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  font-weight: 600;
}

.has-plat {
  background: rgba(34, 197, 94, 0.15);
  color: var(--color-success);
  border: 1px solid rgba(34, 197, 94, 0.3);
}

.no-plat {
  background: rgba(255, 255, 255, 0.04);
  color: var(--text-dim);
  border: 1px solid var(--border-color);
}

.policy-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 13px;
}

.lbl-pol {
  color: var(--text-muted);
}

.code-min {
  background: var(--bg-card);
  padding: 3px 8px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  font-size: 12px;
  color: var(--color-accent);
}

.policy-sub {
  font-size: 11.5px;
  color: var(--text-dim);
  margin-top: 4px;
}

.stats-column {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat-line {
  display: flex;
  justify-content: space-between;
  font-size: 12.5px;
}

/* TOOLBAR */
.tabla-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 14px;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--bg-surface);
  border: 1px solid var(--border-color);
  padding: 6px 12px;
  border-radius: var(--radius-md);
  color: var(--text-muted);
  width: 260px;
}

.search-input {
  background: transparent;
  border: none;
  outline: none;
  color: var(--text-main);
  font-size: 13px;
  width: 100%;
}

.tabs-filtro {
  display: flex;
  background: var(--bg-surface);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 3px;
  gap: 2px;
}

.tab-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  padding: 5px 10px;
  font-size: 12px;
  font-weight: 500;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.15s ease;
}

.tab-btn:hover {
  color: var(--text-main);
}

.tab-btn.activo {
  background: var(--color-primary);
  color: #ffffff;
  font-weight: 600;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* TABLA */
.tabla-container {
  background: var(--bg-surface);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.tabla {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

.tabla th {
  background: rgba(0, 0, 0, 0.2);
  padding: 12px 18px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-bottom: 1px solid var(--border-color);
}

.tabla td {
  padding: 14px 18px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
  font-size: 13.5px;
  vertical-align: middle;
}

.tabla tr:hover {
  background: rgba(255, 255, 255, 0.02);
}

.fila-activa {
  background: rgba(34, 197, 94, 0.04) !important;
}

.celda-version {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.ver-numero {
  display: flex;
  align-items: center;
  gap: 8px;
}

.ver-numero strong {
  font-size: 14.5px;
  font-family: monospace;
  color: var(--text-main);
}

.badge-activa {
  background: var(--color-success);
  color: #0b1317;
  font-size: 10.5px;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
  text-transform: uppercase;
}

.sub-canal {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tag-canal {
  font-size: 11px;
  color: var(--text-dim);
  text-transform: capitalize;
}

.notas-resumen {
  font-size: 11px;
  color: var(--text-muted);
  max-width: 200px;
}

.celda-politica {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.btn-toggle-politica {
  background: transparent;
  border: none;
  padding: 0;
  cursor: pointer;
  text-align: left;
}

.badge-clickable:hover {
  filter: brightness(1.15);
  transform: translateY(-1px);
}

.motivo-sub {
  font-size: 11px;
  color: var(--text-dim);
  max-width: 180px;
}

.tag-version-min {
  background: var(--bg-card);
  padding: 3px 6px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  font-size: 12px;
}

.iconos-plataformas {
  display: flex;
  gap: 6px;
}

.plat-chip {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 3px 7px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}

.plat-ok {
  background: rgba(56, 189, 248, 0.15);
  color: var(--color-accent);
  border: 1px solid rgba(56, 189, 248, 0.3);
}

.plat-missing {
  background: rgba(255, 255, 255, 0.03);
  color: var(--text-dim);
  border: 1px solid var(--border-color);
  opacity: 0.5;
}

/* BADGES */
.badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px;
  border-radius: var(--radius-sm);
  font-size: 11.5px;
  font-weight: 600;
  width: fit-content;
}

.badge-success {
  background: rgba(34, 197, 94, 0.15);
  color: var(--color-success);
  border: 1px solid rgba(34, 197, 94, 0.3);
}

.badge-warning {
  background: rgba(245, 158, 11, 0.15);
  color: var(--color-warning);
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.badge-info {
  background: rgba(56, 189, 248, 0.15);
  color: var(--color-accent);
  border: 1px solid rgba(56, 189, 248, 0.3);
}

.badge-draft {
  background: rgba(148, 163, 184, 0.15);
  color: #94a3b8;
  border: 1px solid rgba(148, 163, 184, 0.3);
}

.badge-danger {
  background: rgba(239, 68, 68, 0.15);
  color: var(--color-danger);
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.fecha-txt {
  font-size: 12px;
  color: var(--text-muted);
}

.th-acciones, .td-acciones {
  text-align: right;
}

.acciones-grupo {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 6px;
}

.btn-icon {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
  background: var(--bg-card);
  color: var(--text-muted);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-icon:hover {
  background: var(--bg-card-hover);
  color: var(--text-main);
  border-color: var(--border-focus);
}

.btn-publicar {
  color: var(--color-success);
  border-color: rgba(34, 197, 94, 0.3);
}

.btn-publicar:hover {
  background: rgba(34, 197, 94, 0.15);
  color: #ffffff;
}

.btn-pausar {
  color: var(--color-warning);
  border-color: rgba(245, 158, 11, 0.3);
}

.btn-pausar:hover {
  background: rgba(245, 158, 11, 0.15);
  color: #ffffff;
}

.btn-danger:hover {
  background: rgba(239, 68, 68, 0.15);
  border-color: rgba(239, 68, 68, 0.4);
  color: var(--color-danger);
}

.estado-vacio {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  text-align: center;
  gap: 12px;
}

.icon-vacio {
  color: var(--text-dim);
}

.spinner-lg {
  width: 32px;
  height: 32px;
  border: 3px solid rgba(255, 255, 255, 0.1);
  border-top-color: var(--color-primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
