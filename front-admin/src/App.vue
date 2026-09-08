<script setup lang="ts">
import { ref, onMounted } from 'vue';
import type { NegocioAdmin } from './types/negocio';
import { obtenerNegocios } from './services/api';
import { obtenerSesionAdmin, cerrarSesionAdmin, type SuperAdminSesion } from './services/auth';

import LoginAdmin from './components/LoginAdmin.vue';
import StatsHeader from './components/StatsHeader.vue';
import TablaNegocios from './components/TablaNegocios.vue';
import ModalCrearNegocio from './components/ModalCrearNegocio.vue';
import ModalPassword from './components/ModalPassword.vue';
import ModalDetalleNegocio from './components/ModalDetalleNegocio.vue';

const sesionAdmin = ref<SuperAdminSesion | null>(null);

const negocios = ref<NegocioAdmin[]>([]);
const cargando = ref(true);
const error = ref('');

const modalCrearAbierto = ref(false);
const modalPasswordAbierto = ref(false);
const modalDetalleAbierto = ref(false);

const negocioSeleccionadoPassword = ref<NegocioAdmin | null>(null);
const negocioSeleccionadoDetalle = ref<NegocioAdmin | null>(null);

function alAutenticar(sesion: SuperAdminSesion) {
  sesionAdmin.value = sesion;
  cargarDatos();
}

function salir() {
  cerrarSesionAdmin();
  sesionAdmin.value = null;
  negocios.value = [];
}

async function cargarDatos() {
  if (!sesionAdmin.value) return;
  cargando.value = true;
  error.value = '';
  try {
    negocios.value = await obtenerNegocios();
  } catch (e: any) {
    error.value = e.message || 'No se pudieron cargar los negocios del servidor';
  } finally {
    cargando.value = false;
  }
}

function abrirModalPassword(negocio: NegocioAdmin) {
  negocioSeleccionadoPassword.value = negocio;
  modalPasswordAbierto.value = true;
}

function abrirModalDetalle(negocio: NegocioAdmin) {
  negocioSeleccionadoDetalle.value = negocio;
  modalDetalleAbierto.value = true;
}

onMounted(() => {
  const sesionExistente = obtenerSesionAdmin();
  if (sesionExistente && sesionExistente.token) {
    sesionAdmin.value = sesionExistente;
    cargarDatos();
  } else {
    cerrarSesionAdmin();
    sesionAdmin.value = null;
  }
});
</script>

<template>
  <!-- SI NO HAY SESIÓN: MOSTRAR PANTALLA DE LOGIN -->
  <LoginAdmin v-if="!sesionAdmin" @login-exitoso="alAutenticar" />

  <!-- SI HAY SESIÓN: MOSTRAR DASHBOARD ADMINISTRADOR -->
  <div v-else class="admin-app">
    <!-- BARRA SUPERIOR -->
    <header class="admin-header">
      <div class="header-contenido">
        <div class="brand">
          <div class="logo">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z"></path><path d="m3.3 7 8.7 5 8.7-5"></path><path d="M12 22V12"></path></svg>
          </div>
          <div>
            <h1>StockAPP Admin</h1>
            <p class="subtitulo">Panel de Administración de Clientes y Negocios</p>
          </div>
        </div>

        <div class="acciones-header">
          <div class="perfil-tag">
            <span class="lbl-admin">{{ sesionAdmin.nombre }}</span>
            <code class="tag-usr">@{{ sesionAdmin.usuario }}</code>
          </div>

          <button class="btn btn-secondary" title="Actualizar datos" @click="cargarDatos">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"></polyline><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"></path></svg>
            <span>Refrescar</span>
          </button>
          <button class="btn btn-primary" @click="modalCrearAbierto = true">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
            <span>Nuevo Cliente / Negocio</span>
          </button>
          <button class="btn btn-ghost" title="Cerrar sesión" @click="salir">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path><polyline points="16 17 21 12 16 7"></polyline><line x1="21" y1="12" x2="9" y2="12"></line></svg>
          </button>
        </div>
      </div>
    </header>

    <!-- CONTENIDO PRINCIPAL -->
    <main class="admin-main">
      <div v-if="error" class="banner-error">
        <p>{{ error }}</p>
        <button class="btn btn-secondary btn-sm" @click="cargarDatos">Reintentar</button>
      </div>

      <StatsHeader :negocios="negocios" />

      <TablaNegocios
        :negocios="negocios"
        :cargando="cargando"
        @abrir-modal-password="abrirModalPassword"
        @abrir-modal-detalle="abrirModalDetalle"
      />
    </main>

    <!-- MODALES -->
    <ModalCrearNegocio
      :abierto="modalCrearAbierto"
      @cerrar="modalCrearAbierto = false"
      @creado="cargarDatos"
    />

    <ModalPassword
      :abierto="modalPasswordAbierto"
      :negocio="negocioSeleccionadoPassword"
      @cerrar="modalPasswordAbierto = false"
      @actualizado="cargarDatos"
    />

    <ModalDetalleNegocio
      :abierto="modalDetalleAbierto"
      :negocio="negocioSeleccionadoDetalle"
      @cerrar="modalDetalleAbierto = false"
      @actualizado="cargarDatos"
    />
  </div>
</template>

<style scoped>
.admin-app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.admin-header {
  background: var(--bg-surface);
  border-bottom: 1px solid var(--border-color);
  padding: 16px 32px;
}

.header-contenido {
  max-width: 1280px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.brand {
  display: flex;
  align-items: center;
  gap: 14px;
}

.logo {
  width: 42px;
  height: 42px;
  border-radius: 10px;
  background: var(--color-primary);
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(46, 110, 115, 0.3);
}

.brand h1 {
  font-size: 20px;
  font-weight: 700;
  line-height: 1.2;
}

.subtitulo {
  font-size: 12.5px;
  color: var(--text-muted);
}

.acciones-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.perfil-tag {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  margin-right: 8px;
}

.lbl-admin {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-main);
}

.tag-usr {
  font-size: 11.5px;
  color: var(--color-accent);
  font-family: monospace;
}

.admin-main {
  max-width: 1280px;
  width: 100%;
  margin: 0 auto;
  padding: 28px 32px;
  flex: 1;

  display: flex;
  flex-direction: column;
  gap: 16px;
}

.banner-error {
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #fca5a5;
  padding: 12px 18px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: space-between;
}
</style>
