<script setup lang="ts">
import { ref } from 'vue';
import { loginSuperAdmin } from '../services/api';
import { guardarSesionAdmin, type SuperAdminSesion } from '../services/auth';

const emit = defineEmits<{
  (e: 'loginExitoso', sesion: SuperAdminSesion): void;
}>();

const usuario = ref('');
const password = ref('');
const cargando = ref(false);
const error = ref('');

async function iniciarSesion() {
  error.value = '';
  if (!usuario.value.trim() || !password.value.trim()) {
    error.value = 'Ingresá tu usuario y contraseña de administrador';
    return;
  }

  cargando.value = true;
  try {
    const sesion = await loginSuperAdmin(usuario.value.trim(), password.value.trim());
    guardarSesionAdmin(sesion);
    emit('loginExitoso', sesion);
  } catch (e: any) {
    error.value = e.message || 'Usuario o contraseña de administrador incorrectos';
  } finally {
    cargando.value = false;
  }
}
</script>

<template>
  <div class="login-wrapper">
    <div class="card login-card">
      <div class="login-header">
        <div class="logo-icon">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path></svg>
        </div>
        <h2>StockAPP Admin</h2>
        <p class="subtitulo">Acceso restringido para el Administrador del Sistema</p>
      </div>

      <form @submit.prevent="iniciarSesion" class="login-form">
        <div class="field">
          <label>Usuario Administrador</label>
          <input
            v-model="usuario"
            class="input"
            :class="{ 'input-error': error }"
            type="text"
            placeholder="Ej: admin"
            required
            autocomplete="username"
            @input="error = ''"
          />
        </div>

        <div class="field">
          <label>Contraseña</label>
          <input
            v-model="password"
            class="input"
            :class="{ 'input-error': error }"
            type="password"
            placeholder="••••••••"
            required
            autocomplete="current-password"
            @input="error = ''"
          />
        </div>

        <p v-if="error" class="error-text">{{ error }}</p>

        <button type="submit" class="btn btn-primary btn-bloque" :disabled="cargando || !usuario || !password">
          <svg v-if="cargando" class="spinner" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12a9 9 0 1 1-6.219-8.56"></path></svg>
          <span>{{ cargando ? 'Verificando...' : 'Iniciar Sesión' }}</span>
        </button>
      </form>
    </div>
  </div>
</template>

<style scoped>
.login-wrapper {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: radial-gradient(circle at center, #16242d 0%, #0b1317 100%);
  padding: 20px;
}

.login-card {
  width: min(400px, 100%);
  padding: 32px;
  display: flex;
  flex-direction: column;
  gap: 24px;
  border-radius: var(--radius-lg);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.4);
}

.login-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 8px;
}

.logo-icon {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  background: var(--color-primary);
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 6px;
  box-shadow: 0 4px 16px rgba(46, 110, 115, 0.4);
}

.login-header h2 {
  font-size: 22px;
  font-weight: 700;
}

.subtitulo {
  font-size: 13px;
  color: var(--text-muted);
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.btn-bloque {
  width: 100%;
  height: 44px;
  font-size: 14px;
  margin-top: 6px;
}

.input-error {
  border-color: var(--color-danger) !important;
  box-shadow: 0 0 0 2px rgba(239, 68, 68, 0.25) !important;
}

.error-text {
  font-size: 13px;
  color: var(--color-danger);
  text-align: center;
}

.spinner {
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  100% {
    transform: rotate(360deg);
  }
}
</style>
