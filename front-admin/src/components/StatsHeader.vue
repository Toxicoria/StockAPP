<script setup lang="ts">
import { computed } from 'vue';
import type { NegocioAdmin } from '../types/negocio';

const props = defineProps<{
  negocios: NegocioAdmin[];
}>();

const totalNegocios = computed(() => props.negocios.length);
const perfilesCompletos = computed(() => props.negocios.filter((n) => n.perfil_completo).length);
const perfilesPendientes = computed(() => props.negocios.filter((n) => !n.perfil_completo).length);
</script>

<template>
  <div class="stats-grid">
    <div class="card stat-card">
      <div class="stat-icon">
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="var(--color-accent)" stroke-width="2"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path><polyline points="9 22 9 12 15 12 15 22"></polyline></svg>
      </div>
      <div class="stat-info">
        <span class="stat-label">Total de Negocios</span>
        <strong class="stat-value">{{ totalNegocios }}</strong>
      </div>
    </div>

    <div class="card stat-card">
      <div class="stat-icon success">
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="var(--color-success)" stroke-width="2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>
      </div>
      <div class="stat-info">
        <span class="stat-label">Perfiles Completos</span>
        <strong class="stat-value">{{ perfilesCompletos }}</strong>
      </div>
    </div>

    <div class="card stat-card">
      <div class="stat-icon warning">
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="var(--color-warning)" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
      </div>
      <div class="stat-info">
        <span class="stat-label">Pendientes de Datos</span>
        <strong class="stat-value">{{ perfilesPendientes }}</strong>
      </div>
    </div>
  </div>
</template>

<style scoped>
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 18px 22px;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: rgba(56, 189, 248, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.stat-icon.success {
  background: rgba(34, 197, 94, 0.1);
}
.stat-icon.warning {
  background: rgba(245, 158, 11, 0.1);
}

.stat-info {
  display: flex;
  flex-direction: column;
}

.stat-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: var(--text-main);
  line-height: 1.2;
}
</style>
