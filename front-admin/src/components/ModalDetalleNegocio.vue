<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import type { NegocioAdmin } from '../types/negocio';
import { actualizarNegocio } from '../services/api';

const props = defineProps<{
  abierto: boolean;
  negocio: NegocioAdmin | null;
}>();

const emit = defineEmits<{
  (e: 'cerrar'): void;
  (e: 'actualizado'): void;
}>();

// Formulario reactivo
const form = ref({
  nombre_negocio: '',
  cuit: '',
  direccion: '',
  telefono: '',
  email_negocio: '',
  nombre_dueno: '',
  usuario: '',
  email_usuario: '',
  password: '',
});

const mostrandoConfirmacion = ref(false);
const cargando = ref(false);
const error = ref('');
const exito = ref('');

// Cargar valores iniciales cada vez que cambia el negocio o abre el modal
watch(
  () => [props.abierto, props.negocio],
  () => {
    if (props.abierto && props.negocio) {
      form.value = {
        nombre_negocio: props.negocio.nombre_negocio || '',
        cuit: props.negocio.cuit || '',
        direccion: props.negocio.direccion || '',
        telefono: props.negocio.telefono || '',
        email_negocio: props.negocio.email_negocio || '',
        nombre_dueno: props.negocio.nombre_dueno || '',
        usuario: props.negocio.usuario || '',
        email_usuario: props.negocio.email_usuario || '',
        password: '',
      };
      mostrandoConfirmacion.value = false;
      error.value = '';
      exito.value = '';
    }
  },
  { immediate: true }
);

function generarPassword() {
  const chars = 'abcdefghjkmnpqrstuvwxyz23456789ABCDEFGHJKLMNPQRSTUVWXYZ';
  let pass = '';
  for (let i = 0; i < 8; i++) {
    pass += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  form.value.password = pass;
}

interface CambioDiff {
  clave: keyof typeof form.value;
  etiqueta: string;
  anterior: string;
  nuevo: string;
}

const camposConfig: { clave: keyof typeof form.value; etiqueta: string }[] = [
  { clave: 'nombre_negocio', etiqueta: 'Nombre del Negocio' },
  { clave: 'cuit', etiqueta: 'CUIT' },
  { clave: 'direccion', etiqueta: 'Dirección' },
  { clave: 'telefono', etiqueta: 'Teléfono de Contacto' },
  { clave: 'email_negocio', etiqueta: 'Email del Negocio' },
  { clave: 'nombre_dueno', etiqueta: 'Nombre del Dueño' },
  { clave: 'usuario', etiqueta: 'Nombre de Usuario' },
  { clave: 'email_usuario', etiqueta: 'Email del Usuario' },
  { clave: 'password', etiqueta: 'Nueva Contraseña' },
];

// Calcular las diferencias entre el estado inicial y el formulario actual
const cambios = computed<CambioDiff[]>(() => {
  if (!props.negocio) return [];
  const lista: CambioDiff[] = [];

  const original: Record<string, string> = {
    nombre_negocio: props.negocio.nombre_negocio || '',
    cuit: props.negocio.cuit || '',
    direccion: props.negocio.direccion || '',
    telefono: props.negocio.telefono || '',
    email_negocio: props.negocio.email_negocio || '',
    nombre_dueno: props.negocio.nombre_dueno || '',
    usuario: props.negocio.usuario || '',
    email_usuario: props.negocio.email_usuario || '',
    password: '',
  };

  for (const item of camposConfig) {
    const valNuevo = form.value[item.clave].trim();
    const valAnt = (original[item.clave] || '').trim();

    if (item.clave === 'password') {
      if (valNuevo.length > 0) {
        lista.push({
          clave: item.clave,
          etiqueta: item.etiqueta,
          anterior: '(Sin cambios)',
          nuevo: `Nueva Clave: ${valNuevo}`,
        });
      }
    } else if (valNuevo !== valAnt) {
      lista.push({
        clave: item.clave,
        etiqueta: item.etiqueta,
        anterior: valAnt || '(Vacio)',
        nuevo: valNuevo || '(Vacío)',
      });
    }
  }

  return lista;
});

function solicitarConfirmacion() {
  if (cambios.value.length === 0) return;
  mostrandoConfirmacion.value = true;
}

async function guardarCambios() {
  if (!props.negocio) return;
  error.value = '';
  exito.value = '';
  cargando.value = true;

  try {
    const payload: any = {};
    for (const c of cambios.value) {
      payload[c.clave] = form.value[c.clave].trim();
    }

    await actualizarNegocio(props.negocio.id_negocio, payload);
    exito.value = '¡Cambios guardados con éxito!';
    emit('actualizado');
    setTimeout(() => {
      mostrandoConfirmacion.value = false;
      emit('cerrar');
    }, 1500);
  } catch (e: any) {
    error.value = e.message || 'No se pudieron guardar los cambios';
  } finally {
    cargando.value = false;
  }
}
</script>

<template>
  <div v-if="abierto && negocio" class="modal-backdrop">
    <div class="modal modal-ancho">
      <!-- CABECERA -->
      <div class="modal-header">
        <div class="tit-box">
          <h3>Ficha Técnica del Negocio</h3>
          <span v-if="negocio.perfil_completo" class="badge badge-success">● Perfil Completo</span>
          <span v-else class="badge badge-warning">● Pendiente Datos</span>
        </div>
        <button class="btn-cerrar" @click="emit('cerrar')">✕</button>
      </div>

      <!-- VISTA 1: FORMULARIO COMPLETO DE DETALLE Y EDICIÓN -->
      <div v-if="!mostrandoConfirmacion" class="modal-body">
        <form @submit.prevent="solicitarConfirmacion" class="form-secciones">
          <!-- SECCIÓN 1: DATOS COMERCIALES -->
          <div class="seccion-box">
            <h4 class="seccion-tit">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="var(--color-accent)" stroke-width="2"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path></svg>
              Datos del Comercio
            </h4>
            <div class="grid-2col">
              <div class="field">
                <label>Nombre del Negocio</label>
                <input v-model="form.nombre_negocio" class="input" type="text" placeholder="Ej: Kiosco Central" required />
              </div>
              <div class="field">
                <label>CUIT / Identificación Fiscal</label>
                <input v-model="form.cuit" class="input" type="text" placeholder="20-12345678-9" />
              </div>
              <div class="field col-span-2">
                <label>Dirección Comercial</label>
                <input v-model="form.direccion" class="input" type="text" placeholder="Ej: Av. San Martín 450, El Bolsón" />
              </div>
              <div class="field">
                <label>Teléfono de Contacto</label>
                <input v-model="form.telefono" class="input" type="text" placeholder="Ej: +54 9 294 4123456" />
              </div>
              <div class="field">
                <label>Email Comercial</label>
                <input v-model="form.email_negocio" class="input" type="email" placeholder="contacto@negocio.com" />
              </div>
            </div>
          </div>

          <!-- SECCIÓN 2: DATOS DEL DUEÑO Y USUARIO DE ACCESO -->
          <div class="seccion-box">
            <h4 class="seccion-tit">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="var(--color-primary-hover)" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>
              Datos del Titular / Usuario
            </h4>
            <div class="grid-2col">
              <div class="field">
                <label>Nombre del Dueño</label>
                <input v-model="form.nombre_dueno" class="input" type="text" placeholder="Ej: Juan Pérez" />
              </div>
              <div class="field">
                <label>Nombre de Usuario (Login)</label>
                <input v-model="form.usuario" class="input" type="text" placeholder="Ej: juan_kiosco" required />
              </div>
              <div class="field">
                <label>Email Personal del Dueño</label>
                <input v-model="form.email_usuario" class="input" type="email" placeholder="juan@email.com" />
              </div>
              <div class="field">
                <div class="label-fila">
                  <label>Cambiar Contraseña</label>
                  <button type="button" class="btn-generar" @click="generarPassword">🎲 Generar clave</button>
                </div>
                <input v-model="form.password" class="input" type="text" placeholder="Dejar en blanco para mantener la actual" />
              </div>
            </div>
          </div>

          <!-- BANNER INDICADOR DE CAMBIOS -->
          <div v-if="cambios.length > 0" class="banner-cambios">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--color-warning)" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
            <span>Se detectaron <strong>{{ cambios.length }}</strong> cambios en los campos.</span>
          </div>

          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="emit('cerrar')">Cancelar</button>
            <button type="submit" class="btn btn-primary" :disabled="cambios.length === 0">
              Aplicar Cambios
            </button>
          </div>
        </form>
      </div>

      <!-- VISTA 2: MODAL DE CONFIRMACIÓN DE CAMBIOS (DIFF) -->
      <div v-else class="modal-body">
        <p class="desc-muted">
          Revisá el resumen de los cambios que estás por aplicar al negocio <strong>{{ negocio.nombre_negocio }}</strong>:
        </p>

        <div class="tabla-diff-box">
          <div v-for="c in cambios" :key="c.clave" class="fila-diff">
            <span class="diff-etiqueta">{{ c.etiqueta }}</span>
            <div class="diff-comparativo">
              <span class="val-antigo">{{ c.anterior }}</span>
              <span class="flecha">➔</span>
              <strong class="val-nuevo">{{ c.nuevo }}</strong>
            </div>
          </div>
        </div>

        <p v-if="error" class="error-msg">{{ error }}</p>
        <p v-if="exito" class="exito-msg">{{ exito }}</p>

        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" @click="mostrandoConfirmacion = false">Volver a Editar</button>
          <button type="button" class="btn btn-primary" :disabled="cargando" @click="guardarCambios">
            {{ cargando ? 'Guardando...' : 'Confirmar y Guardar' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-ancho {
  width: min(640px, 100%);
}

.tit-box {
  display: flex;
  align-items: center;
  gap: 12px;
}

.modal-body {
  display: flex;
  flex-direction: column;
  gap: 18px;
  max-height: 80vh;
  overflow-y: auto;
}

.form-secciones {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.seccion-box {
  background: var(--bg-surface);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.seccion-tit {
  font-size: 14px;
  font-weight: 700;
  color: var(--text-main);
  display: flex;
  align-items: center;
  gap: 8px;
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 8px;
}

.grid-2col {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}
.col-span-2 {
  grid-column: span 2;
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

.banner-cambios {
  background: rgba(245, 158, 11, 0.12);
  border: 1px solid rgba(245, 158, 11, 0.3);
  color: var(--color-warning);
  padding: 10px 14px;
  border-radius: var(--radius-md);
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 10px;
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

/* DIFF COMPONENT */
.tabla-diff-box {
  background: var(--bg-surface);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  display: flex;
  flex-direction: column;
  divide-y: 1px solid var(--border-color);
}

.fila-diff {
  padding: 12px 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  border-bottom: 1px solid var(--border-color);
}
.fila-diff:last-child {
  border-bottom: none;
}

.diff-etiqueta {
  font-size: 12.5px;
  font-weight: 600;
  color: var(--text-muted);
}

.diff-comparativo {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
}

.val-antigo {
  color: var(--color-danger);
  text-decoration: line-through;
  opacity: 0.8;
}

.flecha {
  color: var(--text-dim);
}

.val-nuevo {
  color: var(--color-success);
}

.error-msg {
  color: var(--color-danger);
  font-size: 13px;
}
.exito-msg {
  color: var(--color-success);
  font-size: 13px;
  font-weight: 600;
}
</style>
