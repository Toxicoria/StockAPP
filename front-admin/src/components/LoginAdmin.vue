<script setup lang="ts">
import { ref, nextTick } from 'vue';
import { loginSuperAdmin, verificarTotpAdmin } from '../services/api';
import { guardarSesionAdmin, type SuperAdminSesion } from '../services/auth';

const emit = defineEmits<{
  (e: 'loginExitoso', sesion: SuperAdminSesion): void;
}>();

const paso = ref<'credenciales' | 'totp'>('credenciales');
const usuario = ref('');
const password = ref('');

// Estado del paso 2 (2FA)
const tempToken = ref('');
const esSetup = ref(false);
const qrCode = ref('');
const secretClave = ref('');
const codigoTotp = ref('');
const inputTotpRef = ref<HTMLInputElement | null>(null);

const cargando = ref(false);
const error = ref('');
const copiado = ref(false);

async function iniciarSesion() {
  error.value = '';
  if (!usuario.value.trim() || !password.value.trim()) {
    error.value = 'Ingresá tu usuario y contraseña de administrador';
    return;
  }

  cargando.value = true;
  try {
    const res = await loginSuperAdmin(usuario.value.trim(), password.value.trim());
    tempToken.value = res.temp_token;
    esSetup.value = res.setup_totp;
    qrCode.value = res.qr_code || '';
    secretClave.value = res.secret || '';
    codigoTotp.value = '';
    paso.value = 'totp';

    await nextTick();
    inputTotpRef.value?.focus();
  } catch (e: any) {
    error.value = e.message || 'Usuario o contraseña de administrador incorrectos';
  } finally {
    cargando.value = false;
  }
}

async function verificarCodigo() {
  error.value = '';
  const cod = codigoTotp.value.trim();
  if (cod.length !== 6 || !/^\d{6}$/.test(cod)) {
    error.value = 'El código debe tener exactamente 6 dígitos numéricos';
    return;
  }

  cargando.value = true;
  try {
    const res = await verificarTotpAdmin(tempToken.value, cod);
    const sesion: SuperAdminSesion = {
      id_admin: res.admin.id_admin,
      usuario: res.admin.usuario,
      nombre: res.admin.nombre,
      email: res.admin.email,
      token: res.token,
    };
    guardarSesionAdmin(sesion);
    emit('loginExitoso', sesion);
  } catch (e: any) {
    error.value = e.message || 'Código de verificación incorrecto o expirado';
  } finally {
    cargando.value = false;
  }
}

function volverACredenciales() {
  paso.value = 'credenciales';
  tempToken.value = '';
  codigoTotp.value = '';
  error.value = '';
}

function copiarSecreto() {
  if (!secretClave.value) return;
  navigator.clipboard.writeText(secretClave.value);
  copiado.value = true;
  setTimeout(() => {
    copiado.value = false;
  }, 2000);
}

function onInputCodigo(e: Event) {
  error.value = '';
  const input = e.target as HTMLInputElement;
  // Solo permitir números y máximo 6 caracteres
  const soloNumeros = input.value.replace(/\D/g, '').slice(0, 6);
  codigoTotp.value = soloNumeros;
  if (soloNumeros.length === 6) {
    verificarCodigo();
  }
}
</script>

<template>
  <div class="login-wrapper">
    <div class="card login-card" :class="{ 'card-setup': esSetup && paso === 'totp' }">
      
      <!-- PASO 1: CREDENCIALES (USUARIO Y CONTRASEÑA) -->
      <template v-if="paso === 'credenciales'">
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
            <span>{{ cargando ? 'Verificando...' : 'Siguiente' }}</span>
          </button>
        </form>
      </template>

      <!-- PASO 2: VERIFICACIÓN 2FA (TOTP CON QR O CÓDIGO) -->
      <template v-else>
        <!-- CASO 2A: SETUP INICIAL DE 2FA CON CÓDIGO QR -->
        <div v-if="esSetup" class="setup-totp-container">
          <div class="login-header">
            <div class="logo-icon icon-qr">
              <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7"></rect><rect x="14" y="3" width="7" height="7"></rect><rect x="14" y="14" width="7" height="7"></rect><rect x="3" y="14" width="7" height="7"></rect></svg>
            </div>
            <h2>Configurar Autenticación 2FA</h2>
            <p class="subtitulo">Escaneá el código con Google Authenticator, Authy o tu app favorita</p>
          </div>

          <div class="qr-box">
            <img v-if="qrCode" :src="qrCode" alt="Código QR 2FA" class="qr-image" />
            <div v-else class="qr-placeholder">Cargando QR...</div>
          </div>

          <div class="secreto-manual">
            <span class="lbl-secreto">O ingresá esta clave de respaldo:</span>
            <div class="clave-copiar">
              <code class="codigo-secreto">{{ secretClave }}</code>
              <button type="button" class="btn btn-secondary btn-sm" @click="copiarSecreto" title="Copiar al portapapeles">
                {{ copiado ? '¡Copiado!' : 'Copiar' }}
              </button>
            </div>
          </div>

          <form @submit.prevent="verificarCodigo" class="login-form">
            <div class="field">
              <label>Código de 6 dígitos generado por la App</label>
              <input
                ref="inputTotpRef"
                :value="codigoTotp"
                class="input input-totp"
                :class="{ 'input-error': error }"
                type="text"
                inputmode="numeric"
                pattern="[0-9]*"
                maxlength="6"
                placeholder="000000"
                required
                autocomplete="one-time-code"
                @input="onInputCodigo"
              />
            </div>

            <p v-if="error" class="error-text">{{ error }}</p>

            <div class="botones-totp">
              <button type="button" class="btn btn-secondary" @click="volverACredenciales" :disabled="cargando">
                Volver
              </button>
              <button type="submit" class="btn btn-primary btn-flex" :disabled="cargando || codigoTotp.length !== 6">
                <svg v-if="cargando" class="spinner" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12a9 9 0 1 1-6.219-8.56"></path></svg>
                <span>{{ cargando ? 'Activando...' : 'Activar y Entrar' }}</span>
              </button>
            </div>
          </form>
        </div>

        <!-- CASO 2B: LOGIN HABITUAL CON 2FA YA ACTIVO -->
        <div v-else class="login-totp-container">
          <div class="login-header">
            <div class="logo-icon icon-shield">
              <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="5" y="2" width="14" height="20" rx="2" ry="2"></rect><line x1="12" y1="18" x2="12.01" y2="18"></line></svg>
            </div>
            <h2>Doble Factor de Seguridad</h2>
            <p class="subtitulo">Ingresá el código de 6 dígitos de tu aplicación autenticadora</p>
          </div>

          <form @submit.prevent="verificarCodigo" class="login-form">
            <div class="field">
              <label>Código de Autenticación</label>
              <input
                ref="inputTotpRef"
                :value="codigoTotp"
                class="input input-totp"
                :class="{ 'input-error': error }"
                type="text"
                inputmode="numeric"
                pattern="[0-9]*"
                maxlength="6"
                placeholder="000 000"
                required
                autocomplete="one-time-code"
                @input="onInputCodigo"
              />
            </div>

            <p v-if="error" class="error-text">{{ error }}</p>

            <div class="botones-totp">
              <button type="button" class="btn btn-secondary" @click="volverACredenciales" :disabled="cargando">
                Volver
              </button>
              <button type="submit" class="btn btn-primary btn-flex" :disabled="cargando || codigoTotp.length !== 6">
                <svg v-if="cargando" class="spinner" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12a9 9 0 1 1-6.219-8.56"></path></svg>
                <span>{{ cargando ? 'Verificando...' : 'Iniciar Sesión' }}</span>
              </button>
            </div>
          </form>
        </div>
      </template>

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
  width: min(420px, 100%);
  padding: 32px;
  display: flex;
  flex-direction: column;
  gap: 24px;
  border-radius: var(--radius-lg);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.5);
  transition: width 0.3s ease;
}

.card-setup {
  width: min(480px, 100%);
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

.icon-qr {
  background: #0d9488;
  box-shadow: 0 4px 16px rgba(13, 148, 136, 0.4);
}

.icon-shield {
  background: #2563eb;
  box-shadow: 0 4px 16px rgba(37, 99, 235, 0.4);
}

.login-header h2 {
  font-size: 22px;
  font-weight: 700;
  color: var(--text-main);
}

.subtitulo {
  font-size: 13px;
  color: var(--text-muted);
  line-height: 1.4;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.qr-box {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 12px;
  background: #ffffff;
  border-radius: 14px;
  margin: 4px auto;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  width: fit-content;
}

.qr-image {
  width: 200px;
  height: 200px;
  display: block;
}

.secreto-manual {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  background: rgba(255, 255, 255, 0.04);
  padding: 10px 14px;
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.lbl-secreto {
  font-size: 12px;
  color: var(--text-muted);
}

.clave-copiar {
  display: flex;
  align-items: center;
  gap: 8px;
}

.codigo-secreto {
  font-family: monospace;
  font-size: 13px;
  letter-spacing: 1.5px;
  background: rgba(0, 0, 0, 0.3);
  padding: 4px 8px;
  border-radius: 6px;
  color: #38bdf8;
}

.input-totp {
  text-align: center;
  font-size: 24px;
  letter-spacing: 6px;
  font-weight: 700;
  height: 52px;
  font-family: monospace;
}

.botones-totp {
  display: flex;
  gap: 12px;
  margin-top: 6px;
}

.btn-flex {
  flex: 1;
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
