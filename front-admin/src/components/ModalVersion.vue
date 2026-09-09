<script setup lang="ts">
import { ref, watch } from 'vue';
import type { VersionDesktop } from '../types/version';
import { crearVersion, actualizarVersion } from '../services/api';

const props = defineProps<{
  abierto: boolean;
  versionEditar: VersionDesktop | null;
}>();

const emit = defineEmits<{
  (e: 'cerrar'): void;
  (e: 'guardado'): void;
}>();

const cargando = ref(false);
const error = ref('');

// Form fields
const version = ref('');
const versionMinima = ref('0.1.0');
const esObligatoria = ref(false);
const canal = ref<'produccion' | 'beta'>('produccion');
const estado = ref<'borrador' | 'publicada' | 'desactivada'>('borrador');
const urlWindows = ref('');
const firmaWindows = ref('');
const urlLinux = ref('');
const firmaLinux = ref('');
const notasVersion = ref('');
const motivoObligatoria = ref('');

watch(
  () => props.versionEditar,
  (v) => {
    error.value = '';
    if (v) {
      version.value = v.version;
      versionMinima.value = v.version_minima;
      esObligatoria.value = v.es_obligatoria;
      canal.value = v.canal;
      estado.value = v.estado;
      urlWindows.value = v.url_windows || '';
      firmaWindows.value = v.firma_windows || '';
      urlLinux.value = v.url_linux || '';
      firmaLinux.value = v.firma_linux || '';
      notasVersion.value = v.notas_version || '';
      motivoObligatoria.value = v.motivo_obligatoria || '';
    } else {
      version.value = '';
      versionMinima.value = '0.1.0';
      esObligatoria.value = false;
      canal.value = 'produccion';
      estado.value = 'borrador';
      urlWindows.value = '';
      firmaWindows.value = '';
      urlLinux.value = '';
      firmaLinux.value = '';
      notasVersion.value = '';
      motivoObligatoria.value = '';
    }
  },
  { immediate: true },
);

async function guardar() {
  error.value = '';

  if (!props.versionEditar && !version.value.trim()) {
    error.value = 'El número de versión es obligatorio (ej: 1.1.0)';
    return;
  }
  if (!versionMinima.value.trim()) {
    error.value = 'La versión mínima soportada es obligatoria (ej: 1.0.0)';
    return;
  }

  cargando.value = true;
  try {
    if (props.versionEditar) {
      await actualizarVersion(props.versionEditar.id, {
        version_minima: versionMinima.value.trim(),
        es_obligatoria: esObligatoria.value,
        canal: canal.value,
        estado: estado.value,
        url_windows: urlWindows.value.trim(),
        firma_windows: firmaWindows.value.trim(),
        url_linux: urlLinux.value.trim(),
        firma_linux: firmaLinux.value.trim(),
        notas_version: notasVersion.value.trim(),
        motivo_obligatoria: motivoObligatoria.value.trim(),
      });
    } else {
      await crearVersion({
        version: version.value.trim(),
        version_minima: versionMinima.value.trim(),
        es_obligatoria: esObligatoria.value,
        canal: canal.value,
        estado: estado.value,
        url_windows: urlWindows.value.trim(),
        firma_windows: firmaWindows.value.trim(),
        url_linux: urlLinux.value.trim(),
        firma_linux: firmaLinux.value.trim(),
        notas_version: notasVersion.value.trim(),
        motivo_obligatoria: motivoObligatoria.value.trim(),
      });
    }

    emit('guardado');
    emit('cerrar');
  } catch (e: any) {
    error.value = e.message || 'Error al guardar la versión';
  } finally {
    cargando.value = false;
  }
}
</script>

<template>
  <div v-if="abierto" class="modal-overlay" @click.self="emit('cerrar')">
    <div class="modal-dialog">
      <div class="modal-header">
        <div class="header-title">
          <div class="icon-box">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect width="20" height="14" x="2" y="3" rx="2"></rect>
              <line x1="8" x2="16" y1="21" y2="21"></line>
              <line x1="12" x2="12" y1="17" y2="21"></line>
            </svg>
          </div>
          <div>
            <h2>{{ versionEditar ? `Editar Versión v${versionEditar.version}` : 'Nueva Versión Desktop' }}</h2>
            <p class="modal-sub">Configuración de distribución y políticas de actualización</p>
          </div>
        </div>
        <button class="btn-close" @click="emit('cerrar')">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
        </button>
      </div>

      <div class="modal-body">
        <div v-if="error" class="alert-error">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
          <span>{{ error }}</span>
        </div>

        <form @submit.prevent="guardar" class="form-grid">
          <!-- FILA 1: VERSION Y ESTADO -->
          <div class="form-row cols-3">
            <div class="form-group">
              <label>Número de Versión *</label>
              <input
                v-model="version"
                type="text"
                class="input"
                placeholder="1.1.0"
                :disabled="Boolean(versionEditar)"
                required
              />
              <small v-if="versionEditar" class="hint">La versión semántica no se puede editar una vez creada.</small>
            </div>

            <div class="form-group">
              <label>Canal de Despliegue</label>
              <select v-model="canal" class="input">
                <option value="produccion">Producción (General)</option>
                <option value="beta">Beta (Pruebas)</option>
              </select>
            </div>

            <div class="form-group">
              <label>Estado</label>
              <select v-model="estado" class="input">
                <option value="borrador">🟡 Borrador</option>
                <option value="publicada">🟢 Publicada (Activa)</option>
                <option value="desactivada">🔴 Desactivada (Pausada)</option>
              </select>
            </div>
          </div>

          <!-- SECCIÓN: POLÍTICA DE ACTUALIZACIÓN (OBLIGATORIA VS OPCIONAL) -->
          <div class="card-policy">
            <div class="policy-header">
              <div class="switch-container">
                <input
                  id="chk-obligatoria"
                  type="checkbox"
                  v-model="esObligatoria"
                  class="switch-input"
                />
                <label for="chk-obligatoria" class="switch-label"></label>
              </div>
              <div>
                <strong :class="{ 'text-warning': esObligatoria }">
                  {{ esObligatoria ? '⚠️ Actualización Esencial / Forzada' : '🟢 Actualización Opcional / Sugerida' }}
                </strong>
                <p class="policy-desc">
                  {{ esObligatoria
                    ? 'Bloquea la operación de la caja en versiones viejas hasta que el usuario descargue y aplique la actualización.'
                    : 'Avisa discretamente en la barra de título sin interrumpir las ventas del cajero.'
                  }}
                </p>
              </div>
            </div>

            <div class="policy-fields">
              <div class="form-group">
                <label>Versión Mínima Soportada *</label>
                <input
                  v-model="versionMinima"
                  type="text"
                  class="input"
                  placeholder="1.0.0"
                  required
                />
                <small class="hint">Cualquier app cliente por debajo de esta versión será forzada a actualizar aunque el switch esté en opcional.</small>
              </div>

              <div v-if="esObligatoria" class="form-group">
                <label>Motivo de la Actualización Forzada</label>
                <input
                  v-model="motivoObligatoria"
                  type="text"
                  class="input"
                  placeholder="Ej: Cambio en cálculo fiscal y migración de esquema"
                />
              </div>
            </div>
          </div>

          <!-- SECCIÓN: NOTAS DE LA VERSIÓN / CHANGELOG -->
          <div class="form-group">
            <label>Novedades y Notas de la Versión (Changelog)</label>
            <textarea
              v-model="notasVersion"
              class="input textarea"
              rows="3"
              placeholder="Describe los cambios y mejoras visibles para el cliente..."
            ></textarea>
          </div>

          <!-- SECCIÓN: DESCARGAS WINDOWS -->
          <div class="form-section">
            <div class="section-title">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
                <path d="M0 3.449L9.75 2.1v9.451H0m10.949-9.602L24 0v11.4H10.949M0 12.6h9.75v9.451L0 20.699M10.949 12.6H24V24l-12.95-1.8"></path>
              </svg>
              <span>Paquete Windows (x86_64)</span>
            </div>
            <div class="form-row cols-2">
              <div class="form-group">
                <label>URL de Descarga (.msi.zip / .exe)</label>
                <input
                  v-model="urlWindows"
                  type="url"
                  class="input"
                  placeholder="https://github.com/.../app.msi.zip"
                />
              </div>
              <div class="form-group">
                <label>Firma Criptográfica (Minisign)</label>
                <input
                  v-model="firmaWindows"
                  type="text"
                  class="input code-input"
                  placeholder="dW50cnVzdGVkIGNvbW1lbnQ6..."
                />
              </div>
            </div>
          </div>

          <!-- SECCIÓN: DESCARGAS LINUX -->
          <div class="form-section">
            <div class="section-title">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M12 2a4 4 0 0 0-4 4v5a4 4 0 0 0 8 0V6a4 4 0 0 0-4-4Z"></path>
                <path d="M6 13a6 6 0 0 0 12 0"></path>
                <line x1="8" x2="8" y1="19" y2="21"></line>
                <line x1="16" x2="16" y1="19" y2="21"></line>
                <line x1="12" x2="12" y1="19" y2="22"></line>
              </svg>
              <span>Paquete Linux (x86_64)</span>
            </div>
            <div class="form-row cols-2">
              <div class="form-group">
                <label>URL de Descarga (.AppImage / .deb)</label>
                <input
                  v-model="urlLinux"
                  type="url"
                  class="input"
                  placeholder="https://github.com/.../app.AppImage.tar.gz"
                />
              </div>
              <div class="form-group">
                <label>Firma Criptográfica (Minisign)</label>
                <input
                  v-model="firmaLinux"
                  type="text"
                  class="input code-input"
                  placeholder="dW50cnVzdGVkIGNvbW1lbnQ6..."
                />
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="emit('cerrar')">
              Cancelar
            </button>
            <button type="submit" class="btn btn-primary" :disabled="cargando">
              <svg v-if="cargando" class="spinner" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"></circle>
                <path d="M12 6v6l4 2"></path>
              </svg>
              <span>{{ cargando ? 'Guardando...' : (versionEditar ? 'Guardar Cambios' : 'Registrar Versión') }}</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 999;
  padding: 20px;
}

.modal-dialog {
  background: var(--bg-surface);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  width: 100%;
  max-width: 820px;
  max-height: 92vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
  overflow: hidden;
}

.modal-header {
  padding: 18px 24px;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.icon-box {
  width: 38px;
  height: 38px;
  border-radius: 8px;
  background: rgba(46, 110, 115, 0.2);
  color: var(--color-accent);
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-header h2 {
  font-size: 17px;
  font-weight: 700;
}

.modal-sub {
  font-size: 12px;
  color: var(--text-muted);
}

.btn-close {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 6px;
  border-radius: var(--radius-sm);
  transition: all 0.15s ease;
}

.btn-close:hover {
  color: var(--text-main);
  background: rgba(255, 255, 255, 0.08);
}

.modal-body {
  padding: 22px 24px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.alert-error {
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #fca5a5;
  padding: 10px 14px;
  border-radius: var(--radius-md);
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.form-grid {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-row {
  display: grid;
  gap: 14px;
}

.cols-2 {
  grid-template-columns: 1fr 1fr;
}

.cols-3 {
  grid-template-columns: 1.2fr 1fr 1fr;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  font-size: 12.5px;
  font-weight: 600;
  color: var(--text-main);
}

.hint {
  font-size: 11px;
  color: var(--text-dim);
  line-height: 1.3;
}

.input {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  color: var(--text-main);
  padding: 9px 12px;
  border-radius: var(--radius-md);
  font-family: inherit;
  font-size: 13.5px;
  outline: none;
  transition: border-color 0.15s ease;
}

.input:focus {
  border-color: var(--border-focus);
}

.input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.textarea {
  resize: vertical;
  min-height: 70px;
}

.code-input {
  font-family: monospace;
  font-size: 12px;
}

/* POLICY CARD */
.card-policy {
  background: rgba(24, 38, 47, 0.6);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.policy-header {
  display: flex;
  align-items: center;
  gap: 14px;
}

.policy-desc {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 2px;
}

.text-warning {
  color: var(--color-warning);
}

.policy-fields {
  display: grid;
  grid-template-columns: 1fr 1.5fr;
  gap: 14px;
  padding-top: 10px;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
}

/* SWITCH */
.switch-container {
  position: relative;
  width: 44px;
  height: 24px;
  flex-shrink: 0;
}

.switch-input {
  opacity: 0;
  width: 0;
  height: 0;
}

.switch-label {
  position: absolute;
  cursor: pointer;
  inset: 0;
  background-color: var(--border-color);
  border-radius: 24px;
  transition: 0.2s;
}

.switch-label:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background-color: #ffffff;
  border-radius: 50%;
  transition: 0.2s;
}

.switch-input:checked + .switch-label {
  background-color: var(--color-warning);
}

.switch-input:checked + .switch-label:before {
  transform: translateX(20px);
}

/* SECTION */
.form-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: var(--radius-md);
  padding: 14px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12.5px;
  font-weight: 700;
  color: var(--color-accent);
}

.modal-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--border-color);
}

.spinner {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
