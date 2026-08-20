<script setup lang="ts">
import { ref, watch } from 'vue';
import { cambiarPasswordUsuario } from '../services/api';
import type { NegocioAdmin } from '../types/negocio';

const props = defineProps<{
  abierto: boolean;
  negocio: NegocioAdmin | null;
}>();

const emit = defineEmits<{
  (e: 'cerrar'): void;
  (e: 'actualizado'): void;
}>();

const nuevaPassword = ref('');
const cargando = ref(false);
const error = ref('');
const exito = ref('');

watch(
  () => props.abierto,
  (val) => {
    if (val) {
      nuevaPassword.value = '';
      error.value = '';
      exito.value = '';
    }
  }
);

function generarPassword() {
  const chars = 'abcdefghjkmnpqrstuvwxyz23456789ABCDEFGHJKLMNPQRSTUVWXYZ';
  let pass = '';
  for (let i = 0; i < 8; i++) {
    pass += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  nuevaPassword.value = pass;
}

async function guardar() {
  if (!props.negocio || !props.negocio.id_dueno) return;
  error.value = '';
  exito.value = '';

  if (!nuevaPassword.value.trim() || nuevaPassword.value.length < 4) {
    error.value = 'La contraseña debe tener al menos 4 caracteres';
    return;
  }

  cargando.value = true;
  try {
    await cambiarPasswordUsuario(props.negocio.id_dueno, nuevaPassword.value.trim());
    exito.value = '¡Contraseña actualizada con éxito!';
    emit('actualizado');
    setTimeout(() => {
      emit('cerrar');
    }, 1500);
  } catch (e: any) {
    error.value = e.message || 'No se pudo actualizar la contraseña';
  } finally {
    cargando.value = false;
  }
}
</script>

<template>
  <div v-if="abierto && negocio" class="modal-backdrop">
    <div class="modal">
      <div class="modal-header">
        <h3>Cambiar Contraseña</h3>
        <button class="btn-cerrar" @click="emit('cerrar')">✕</button>
      </div>

      <div class="modal-body">
        <div class="info-negocio">
          <span class="negocio-tit">{{ negocio.nombre_negocio }}</span>
          <span class="negocio-sub">Usuario: <strong>{{ negocio.usuario }}</strong></span>
        </div>

        <form @submit.prevent="guardar" class="form-grid">
          <div class="field">
            <div class="label-fila">
              <label>Nueva Contraseña</label>
              <button type="button" class="btn-generar" @click="generarPassword">🎲 Generar clave</button>
            </div>
            <input v-model="nuevaPassword" class="input" type="text" placeholder="Ingresá la nueva clave" required />
          </div>

          <p v-if="error" class="error-msg">{{ error }}</p>
          <p v-if="exito" class="exito-msg">{{ exito }}</p>

          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="emit('cerrar')">Cancelar</button>
            <button type="submit" class="btn btn-primary" :disabled="cargando || !nuevaPassword">
              {{ cargando ? 'Guardando...' : 'Cambiar Clave' }}
            </button>
          </div>
        </form>
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

.info-negocio {
  background: var(--bg-surface);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 12px 16px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.negocio-tit {
  font-size: 15px;
  font-weight: 700;
}
.negocio-sub {
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

.error-msg {
  color: var(--color-danger);
  font-size: 13px;
}
.exito-msg {
  color: var(--color-success);
  font-size: 13px;
  font-weight: 600;
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
</style>
