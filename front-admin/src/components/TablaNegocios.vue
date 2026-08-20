<script setup lang="ts">
import { ref, computed } from 'vue';
import type { NegocioAdmin } from '../types/negocio';

const props = defineProps<{
  negocios: NegocioAdmin[];
  cargando: boolean;
}>();

const emit = defineEmits<{
  (e: 'abrirModalPassword', negocio: NegocioAdmin): void;
  (e: 'abrirModalDetalle', negocio: NegocioAdmin): void;
}>();

const busqueda = ref('');
const copiadoId = ref<number | null>(null);

const negociosFiltrados = computed(() => {
  const q = busqueda.value.trim().toLowerCase();
  if (!q) return props.negocios;
  return props.negocios.filter(
    (n) =>
      n.nombre_negocio.toLowerCase().includes(q) ||
      n.usuario.toLowerCase().includes(q) ||
      n.nombre_dueno.toLowerCase().includes(q) ||
      n.direccion.toLowerCase().includes(q)
  );
});

function copiarInformacion(n: NegocioAdmin) {
  const texto = `• Negocio: ${n.nombre_negocio}
• Ubicación: ${n.direccion || 'Sin especificar'}
• Dueño: ${n.nombre_dueno || n.usuario}
• Usuario: ${n.usuario}
• Alta: ${n.fecha_alta}`;

  navigator.clipboard.writeText(texto);
  copiadoId.value = n.id_negocio;
  setTimeout(() => (copiadoId.value = null), 2000);
}
</script>

<template>
  <div class="card contenedor-tabla">
    <div class="tabla-cabecera">
      <div class="buscador-box">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--text-dim)" stroke-width="2"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
        <input
          v-model="busqueda"
          class="input entrada-buscar"
          type="text"
          placeholder="Buscar negocio por nombre, usuario o ubicación..."
        />
      </div>
      <span class="total-badge">{{ negociosFiltrados.length }} negocios encontrados</span>
    </div>

    <div class="tabla-scroll">
      <table class="table">
        <thead>
          <tr>
            <th class="th-id">ID</th>
            <th>Negocio / Ubicación</th>
            <th>Dueño / Usuario</th>
            <th>Estado</th>
            <th>Fecha de Alta</th>
            <th class="th-acciones">Acciones</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="cargando">
            <td colspan="6" class="vacio">Cargando negocios...</td>
          </tr>
          <tr v-else-if="negociosFiltrados.length === 0">
            <td colspan="6" class="vacio">No se encontraron negocios registrados.</td>
          </tr>
          <tr
            v-for="n in negociosFiltrados"
            :key="n.id_negocio"
            class="fila-negocio"
            title="Hacer clic para ver detalle y editar"
            @click="emit('abrirModalDetalle', n)"
          >
            <td class="col-id">#{{ n.id_negocio }}</td>
            <td class="col-negocio">
              <div class="celda-stack">
                <span class="tit-negocio">{{ n.nombre_negocio }}</span>
                <span class="sub-meta">{{ n.direccion || 'Ubicación pendiente' }}</span>
              </div>
            </td>
            <td class="col-dueno">
              <div class="celda-stack">
                <span class="tit-dueno">{{ n.nombre_dueno || 'Sin nombre registrado' }}</span>
                <code class="tag-usuario">@{{ n.usuario }}</code>
              </div>
            </td>
            <td>
              <span v-if="n.perfil_completo" class="badge badge-success">
                ● Completo
              </span>
              <span v-else class="badge badge-warning">
                ● Pendiente datos
              </span>
            </td>
            <td class="col-fecha">{{ n.fecha_alta }}</td>
            <td class="col-acciones" @click.stop>
              <div class="acciones-group">
                <button
                  class="btn btn-primary btn-sm"
                  title="Ver Ficha y Editar"
                  @click.stop="emit('abrirModalDetalle', n)"
                >
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>
                  <span>Ficha</span>
                </button>
                <button
                  class="btn btn-secondary btn-sm"
                  title="Cambiar Contraseña"
                  @click.stop="emit('abrirModalPassword', n)"
                >
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg>
                  <span>Clave</span>
                </button>
                <button
                  class="btn btn-ghost btn-sm"
                  title="Copiar información del negocio"
                  @click.stop="copiarInformacion(n)"
                >
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
                  <span>{{ copiadoId === n.id_negocio ? '¡Copiado!' : 'Copiar' }}</span>
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
.contenedor-tabla {
  padding: 0;
  overflow: hidden;
}

.tabla-cabecera {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}

.buscador-box {
  position: relative;
  flex: 1;
  max-width: 440px;
  display: flex;
  align-items: center;
}
.buscador-box svg {
  position: absolute;
  left: 14px;
}
.entrada-buscar {
  padding-left: 40px;
  background: var(--bg-surface);
}

.total-badge {
  font-size: 12.5px;
  color: var(--text-muted);
  font-weight: 500;
}

.tabla-scroll {
  overflow-x: auto;
}

.table {
  width: 100%;
  border-collapse: collapse;
  table-layout: auto;
}

.table th {
  padding: 14px 18px;
  text-align: left;
  font-size: 11.5px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-muted);
  background: var(--bg-surface);
  border-bottom: 1px solid var(--border-color);
  white-space: nowrap;
}

.th-id {
  width: 60px;
}

.th-acciones {
  text-align: right;
  width: 240px;
}

.table td {
  padding: 14px 18px;
  border-bottom: 1px solid var(--border-color);
  vertical-align: middle;
}

.fila-negocio {
  cursor: pointer;
  transition: background-color 0.15s ease;
}
.fila-negocio:hover {
  background: var(--bg-card-hover);
}

.col-id {
  font-weight: 700;
  color: var(--text-dim);
  font-family: monospace;
  white-space: nowrap;
}

.celda-stack {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.tit-negocio, .tit-dueno {
  font-weight: 600;
  color: var(--text-main);
  font-size: 14px;
}

.sub-meta {
  font-size: 12px;
  color: var(--text-muted);
}

.tag-usuario {
  font-family: monospace;
  font-size: 12px;
  color: var(--color-accent);
}

.col-fecha {
  font-size: 12.5px;
  color: var(--text-muted);
  white-space: nowrap;
}

.col-acciones {
  text-align: right;
}

.acciones-group {
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
}

.vacio {
  text-align: center;
  padding: 40px;
  color: var(--text-muted);
}
</style>
