<script>
  import { sesion } from '$lib/sesion.svelte.js';

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
  async function cerrar() {
    if (enTauri) (await ventana()).close();
  }
</script>

<div class="barra" data-tauri-drag-region>
  <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="var(--color-accent-600)" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z"></path><path d="m3.3 7 8.7 5 8.7-5"></path><path d="M12 22V12"></path></svg>
  <span class="titulo">{sesion.negocio || 'StockAPP'} — Control de stock</span>
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
