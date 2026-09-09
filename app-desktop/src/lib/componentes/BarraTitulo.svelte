<script>
  import { sesion } from '$lib/sesion.svelte.js';
  import { solicitarCerrarApp } from '$lib/confirmacion.svelte.js';
  import {
    estadoUpdater,
    iniciarDescargaEInstalacion,
    reiniciarYAplicar,
  } from '$lib/updater.svelte.js';

  // Fuera de Tauri (vite dev en el navegador) los botones no hacen nada
  // pero tampoco rompen: la API de ventana solo se invoca si existe.
  const enTauri = typeof window !== 'undefined' && '__TAURI_INTERNALS__' in window;

  async function ventana() {
    const { getCurrentWindow } = await import('@tauri-apps/api/window');
    return getCurrentWindow();
  }

  async function minimizar() {
    if (enTauri) (await ventana()).minimize();
  }
  async function maximizar() {
    if (enTauri) (await ventana()).toggleMaximize();
  }
  function cerrar() {
    solicitarCerrarApp();
  }
</script>

<div class="barra" data-tauri-drag-region>
  <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="var(--color-accent-600)" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z"></path><path d="m3.3 7 8.7 5 8.7-5"></path><path d="M12 22V12"></path></svg>
  <span class="titulo">{sesion.negocio || 'StockAPP'} — Control de stock</span>

  <!-- INDICADOR DISCRETO DE ACTUALIZACIÓN OPCIONAL -->
  {#if estadoUpdater.disponible && !estadoUpdater.esObligatoria}
    <div class="update-banner">
      {#if estadoUpdater.descargada}
        <button class="update-pill pill-ready" onclick={reiniciarYAplicar} title="Actualización descargada. Clic para reiniciar y aplicar.">
          <span class="dot-green">●</span>
          <span>v{estadoUpdater.versionNueva} lista</span>
          <span class="action-btn-mini">Reiniciar</span>
        </button>
      {:else if estadoUpdater.descargando}
        <div class="update-pill pill-downloading">
          <span class="spinner-mini"></span>
          <span>Descargando v{estadoUpdater.versionNueva} ({estadoUpdater.progreso}%)</span>
        </div>
      {:else}
        <button class="update-pill pill-available" onclick={iniciarDescargaEInstalacion} title={estadoUpdater.notas || 'Nueva versión opcional disponible'}>
          <span class="dot-accent">●</span>
          <span>v{estadoUpdater.versionNueva} disponible</span>
          <span class="action-btn-mini">Actualizar</span>
        </button>
      {/if}
    </div>
  {/if}

  <div class="botones">
    <button class="boton" onclick={minimizar} title="Minimizar">–</button>
    <button class="boton chico" onclick={maximizar} title="Maximizar">☐</button>
    <button class="boton cerrar" onclick={cerrar} title="Cerrar">✕</button>
  </div>
</div>

<style>
  .barra {
    display: flex;
    align-items: center;
    height: 36px;
    flex: none;
    padding-left: 12px;
    background: var(--color-surface);
    border-bottom: 1px solid var(--color-divider);
    user-select: none;
  }
  .titulo {
    font-size: 12px;
    margin-left: 8px;
    pointer-events: none; /* que el texto no bloquee el arrastre de la barra */
  }

  /* PÍLDORA DE ACTUALIZACIÓN */
  .update-banner {
    margin-left: 18px;
    display: flex;
    align-items: center;
  }

  .update-pill {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 3px 10px;
    border-radius: 12px;
    font-size: 11px;
    font-weight: 600;
    border: 1px solid transparent;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .pill-available {
    background: rgba(56, 189, 248, 0.12);
    border-color: rgba(56, 189, 248, 0.3);
    color: #38bdf8;
  }

  .pill-available:hover {
    background: rgba(56, 189, 248, 0.22);
  }

  .pill-downloading {
    background: rgba(245, 158, 11, 0.12);
    border-color: rgba(245, 158, 11, 0.3);
    color: #fbbf24;
    cursor: default;
  }

  .pill-ready {
    background: rgba(34, 197, 94, 0.15);
    border-color: rgba(34, 197, 94, 0.35);
    color: #22c55e;
  }

  .pill-ready:hover {
    background: rgba(34, 197, 94, 0.25);
  }

  .dot-accent { color: #38bdf8; font-size: 8px; }
  .dot-green { color: #22c55e; font-size: 8px; }

  .action-btn-mini {
    background: rgba(255, 255, 255, 0.12);
    padding: 1px 6px;
    border-radius: 4px;
    font-size: 10px;
    text-transform: uppercase;
    font-weight: 700;
  }

  .spinner-mini {
    width: 10px;
    height: 10px;
    border: 2px solid rgba(251, 191, 36, 0.3);
    border-top-color: #fbbf24;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }

  .botones {
    margin-left: auto;
    display: flex;
    height: 100%;
  }
  .boton {
    width: 46px;
    border: 0;
    background: transparent;
    color: var(--color-text);
    font-size: 13px;
    cursor: pointer;
  }
  .boton:hover { background: color-mix(in srgb, var(--color-text) 8%, transparent); }
  .chico { font-size: 11px; }
  .cerrar:hover { background: #c0392b; color: #fff; }
</style>

