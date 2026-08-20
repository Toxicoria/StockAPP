<script setup lang="ts">
import { ref } from 'vue';
import { crearNegocioCliente } from '../services/api';

const props = defineProps<{
  abierto: boolean;
}>();

const emit = defineEmits<{
  (e: 'cerrar'): void;
  (e: 'creado'): void;
}>();

const nombreNegocio = ref('');
const usuario = ref('');
const password = ref('');
const nombreDueno = ref('');

const cargando = ref(false);
const error = ref('');
const copiado = ref(false);
const negocioCreadoExito = ref<{ usuario: string; password: string; nombre_negocio: string } | null>(null);

function generarPassword() {
  const chars = 'abcdefghjkmnpqrstuvwxyz23456789ABCDEFGHJKLMNPQRSTUVWXYZ';
  let pass = '';
  for (let i = 0; i < 8; i++) {
    pass += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  password.value = pass;
}

function sugerirUsuario() {
  if (nombreNegocio.value && !usuario.value) {
    const slug = nombreNegocio.value
      .toLowerCase()
      .normalize('NFD')
      .replace(/[\u0300-\u036f]/g, '')
      .replace(/[^a-z0-9]/g, '_')
      .replace(/_+/g, '_')
      .replace(/^_+|_+$/g, '');
    usuario.value = slug;
  }
}

async function guardar() {
  error.value = '';
  if (!usuario.value.trim() || !password.value.trim()) {
    error.value = 'El usuario y la contraseña son obligatorios';
    return;
  }

  cargando.value = true;
  try {
    const res = await crearNegocioCliente({
      nombre_negocio: nombreNegocio.value.trim(),
      usuario: usuario.value.trim(),
      password: password.value.trim(),
      nombre_dueno: nombreDueno.value.trim(),
    });

    negocioCreadoExito.value = {
      nombre_negocio: res.nombre_negocio,
      usuario: res.usuario,
      password: password.value.trim(),
    };
    emit('creado');
  } catch (e: any) {
    error.value = e.message || 'No se pudo crear la cuenta del cliente';
  } finally {
    cargando.value = false;
  }
}

function copiarCredenciales() {
  if (!negocioCreadoExito.value) return;
  const texto = `Hola! Aquí están tus credenciales de acceso a StockAPP:

• Negocio: ${negocioCreadoExito.value.nombre_negocio}
• Usuario: ${negocioCreadoExito.value.usuario}
• Contraseña: ${negocioCreadoExito.value.password}

Descargá la app e ingresá para completar tu perfil.`;

  navigator.clipboard.writeText(texto);
  copiado.value = true;
  setTimeout(() => (copiado.value = false), 2500);
}

function cerrarModal() {
  nombreNegocio.value = '';
  usuario.value = '';
  password.value = '';
  nombreDueno.value = '';
  error.value = '';
  negocioCreadoExito.value = null;
  emit('cerrar');
}
</script>

<template>
  <div v-if="abierto" class="modal-backdrop">
    <div class="modal">
      <div class="modal-header">
        <h3>{{ negocioCreadoExito ? '¡Cliente Creado con Éxito!' : 'Crear Nuevo Cliente / Negocio' }}</h3>
        <button class="btn-cerrar" @click="cerrarModal">✕</button>
      </div>

      <!-- VISTA 1: FORMULARIO DE CREACIÓN -->
      <div v-if="!negocioCreadoExito" class="modal-body">
        <p class="desc-muted">
          Creá el usuario y contraseña inicial para tu cliente. El cliente luego completará sus datos comerciales desde la app.
        </p>

        <form @submit.prevent="guardar" class="form-grid">
          <div class="field">
            <label>Nombre del Negocio</label>
            <input
              v-model="nombreNegocio"
              class="input"
              type="text"
              placeholder="Ej: Kiosco Central"
              @blur="sugerirUsuario"
            />
          </div>

          <div class="field">
            <label>Nombre del Dueño (Opcional)</label>
            <input v-model="nombreDueno" class="input" type="text" placeholder="Ej: Juan Pérez" />
          </div>

          <div class="field">
            <label>Nombre de Usuario <span class="req">*</span></label>
            <input v-model="usuario" class="input" type="text" placeholder="Ej: juan_kiosco" required />
          </div>

          <div class="field">
            <div class="label-fila">
              <label>Contraseña <span class="req">*</span></label>
              <button type="button" class="btn-generar" @click="generarPassword">🎲 Generar clave</button>
            </div>
            <input v-model="password" class="input" type="text" placeholder="Mínimo 4 caracteres" required />
          </div>

          <p v-if="error" class="error-msg">{{ error }}</p>

          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="cerrarModal">Cancelar</button>
            <button type="submit" class="btn btn-primary" :disabled="cargando || !usuario || !password">
              {{ cargando ? 'Creando...' : 'Crear Cliente' }}
            </button>
          </div>
        </form>
      </div>

      <!-- VISTA 2: CREDENCIALES CREADAS LISTAS PARA COPIAR -->
      <div v-else class="modal-body exito-box">
        <div class="tarjeta-credenciales">
          <div class="fila-cred">
            <span class="lbl">Negocio:</span>
            <strong>{{ negocioCreadoExito.nombre_negocio }}</strong>
          </div>
          <div class="fila-cred">
            <span class="lbl">Usuario:</span>
            <code class="val-code">{{ negocioCreadoExito.usuario }}</code>
          </div>
          <div class="fila-cred">
            <span class="lbl">Contraseña:</span>
            <code class="val-code">{{ negocioCreadoExito.password }}</code>
          </div>
        </div>

        <p class="tip-text">Copiá estas credenciales para enviárselas a tu cliente.</p>

        <div class="modal-footer">
          <button class="btn btn-primary btn-copiar" @click="copiarCredenciales">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
            <span>{{ copiado ? '¡Copiado al Portapapeles!' : 'Copiar Credenciales' }}</span>
          </button>
          <button class="btn btn-secondary" @click="cerrarModal">Listo</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.desc-muted {
  font-size: 13px;
  color: var(--text-muted);
}

.form-grid {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.label-fila {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.btn-generar {
  background: none;
  border: none;
  color: var(--color-accent);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
}
.btn-generar:hover {
  text-decoration: underline;
}

.req {
  color: var(--color-danger);
}

.error-msg {
  color: var(--color-danger);
  font-size: 13px;
}

.btn-cerrar {
  background: none;
  border: none;
  color: var(--text-muted);
  font-size: 18px;
  cursor: pointer;
}
.btn-cerrar:hover {
  color: var(--text-main);
}

.tarjeta-credenciales {
  background: var(--bg-surface);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.fila-cred {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.lbl {
  font-size: 13px;
  color: var(--text-muted);
}
.val-code {
  font-family: monospace;
  font-size: 14px;
  font-weight: 700;
  background: rgba(56, 189, 248, 0.15);
  color: var(--color-accent);
  padding: 2px 8px;
  border-radius: 4px;
}

.tip-text {
  font-size: 13px;
  color: var(--text-muted);
  text-align: center;
}

.btn-copiar {
  flex: 1;
}
</style>
