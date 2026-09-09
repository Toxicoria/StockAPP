<script>
  import { estadoUpdater, reiniciarYAplicar, iniciarDescargaEInstalacion } from '$lib/updater.svelte.js';
  import { solicitarCerrarApp } from '$lib/confirmacion.svelte.js';
</script>

{#if estadoUpdater.disponible && estadoUpdater.esObligatoria}
  <div class="bloqueo-fondo">
    <div class="tarjeta-modal">
      <div class="icono-alerta">
        <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"></path>
          <line x1="12" y1="9" x2="12" y2="13"></line>
          <line x1="12" y1="17" x2="12.01" y2="17"></line>
        </svg>
      </div>

      <h2>Actualización Obligatoria Requerida</h2>
      <div class="badge-version">v{estadoUpdater.versionNueva}</div>

      <p class="mensaje-explicativo">
        Esta versión contiene cambios estructurales indispensables para operar con el servidor central de StockAPP.
        {#if estadoUpdater.motivo}
          <span class="motivo-caja">
            <strong>Motivo:</strong> {estadoUpdater.motivo}
          </span>
        {/if}
      </p>

      <!-- BARRA DE PROGRESO DE DESCARGA -->
      <div class="progreso-contenedor">
        <div class="progreso-etiquetas">
          <span>
            {#if estadoUpdater.descargada}
              ✅ Descarga completa. Lista para instalar.
            {:else if estadoUpdater.descargando}
              Descargando paquete de actualización...
            {:else}
              Preparando descarga...
            {/if}
          </span>
          <span class="porcentaje">{estadoUpdater.progreso}%</span>
        </div>

        <div class="barra-track">
          <div class="barra-fill" style="width: {estadoUpdater.progreso}%;"></div>
        </div>
      </div>

      {#if estadoUpdater.notas}
        <div class="notas-caja">
          <span class="notas-titulo">Novedades de esta versión:</span>
          <p>{estadoUpdater.notas}</p>
        </div>
      {/if}

      {#if estadoUpdater.error}
        <div class="alerta-error">
          <span>{estadoUpdater.error}</span>
          <button class="btn-reintentar" onclick={iniciarDescargaEInstalacion}>Reintentar</button>
        </div>
      {/if}

      <!-- ACCIONES -->
      <div class="acciones-pie">
        <button class="btn-salir" onclick={solicitarCerrarApp}>
          Salir del sistema
        </button>

        <button
          class="btn-instalar"
          disabled={!estadoUpdater.descargada}
          onclick={reiniciarYAplicar}
        >
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="23 4 23 10 17 10"></polyline>
            <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"></path>
          </svg>
          <span>{estadoUpdater.descargada ? 'Reiniciar e Instalar Ahora' : 'Descargando...'}</span>
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .bloqueo-fondo {
    position: fixed;
    inset: 0;
    background: rgba(11, 19, 23, 0.88);
    backdrop-filter: blur(8px);
    z-index: 9999;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
    user-select: none;
  }

  .tarjeta-modal {
    background: #121d23;
    border: 1px solid rgba(245, 158, 11, 0.4);
    border-radius: 16px;
    box-shadow: 0 24px 48px rgba(0, 0, 0, 0.6), 0 0 32px rgba(245, 158, 11, 0.1);
    max-width: 520px;
    width: 100%;
    padding: 32px 28px 26px;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 14px;
    animation: aparecer 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes aparecer {
    from { opacity: 0; transform: scale(0.96); }
    to { opacity: 1; transform: scale(1); }
  }

  .icono-alerta {
    width: 60px;
    height: 60px;
    border-radius: 50%;
    background: rgba(245, 158, 11, 0.15);
    color: #f59e0b;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  h2 {
    font-size: 19px;
    font-weight: 700;
    color: #f1f5f9;
    margin: 0;
  }

  .badge-version {
    background: rgba(245, 158, 11, 0.2);
    color: #fbbf24;
    border: 1px solid rgba(245, 158, 11, 0.35);
    font-family: monospace;
    font-weight: 700;
    font-size: 13px;
    padding: 3px 10px;
    border-radius: 6px;
  }

  .mensaje-explicativo {
    font-size: 13.5px;
    color: #94a3b8;
    line-height: 1.5;
    margin: 0;
  }

  .motivo-caja {
    display: block;
    margin-top: 8px;
    padding: 8px 12px;
    background: rgba(255, 255, 255, 0.04);
    border-radius: 8px;
    font-size: 12.5px;
    color: #cbd5e1;
  }

  /* BARRA DE PROGRESO */
  .progreso-contenedor {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin-top: 8px;
  }

  .progreso-etiquetas {
    display: flex;
    justify-content: space-between;
    font-size: 12px;
    color: #94a3b8;
  }

  .porcentaje {
    font-weight: 700;
    color: #fbbf24;
    font-family: monospace;
  }

  .barra-track {
    width: 100%;
    height: 8px;
    background: #1e293b;
    border-radius: 8px;
    overflow: hidden;
  }

  .barra-fill {
    height: 100%;
    background: linear-gradient(90deg, #f59e0b, #eab308);
    transition: width 0.3s ease;
  }

  .notas-caja {
    width: 100%;
    background: rgba(0, 0, 0, 0.25);
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 8px;
    padding: 10px 14px;
    text-align: left;
    font-size: 12px;
    color: #94a3b8;
    max-height: 80px;
    overflow-y: auto;
  }

  .notas-titulo {
    display: block;
    font-weight: 600;
    color: #cbd5e1;
    margin-bottom: 3px;
  }

  .alerta-error {
    background: rgba(239, 68, 68, 0.15);
    border: 1px solid rgba(239, 68, 68, 0.3);
    color: #fca5a5;
    padding: 8px 12px;
    border-radius: 8px;
    font-size: 12px;
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .btn-reintentar {
    background: rgba(239, 68, 68, 0.3);
    border: none;
    color: #ffffff;
    padding: 3px 8px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 11px;
  }

  /* BOTONES */
  .acciones-pie {
    width: 100%;
    display: flex;
    gap: 12px;
    margin-top: 10px;
  }

  .btn-salir {
    flex: 1;
    height: 42px;
    background: transparent;
    border: 1px solid rgba(255, 255, 255, 0.15);
    color: #94a3b8;
    border-radius: 10px;
    font-weight: 600;
    font-size: 13px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .btn-salir:hover {
    background: rgba(255, 255, 255, 0.06);
    color: #f1f5f9;
  }

  .btn-instalar {
    flex: 1.8;
    height: 42px;
    background: #f59e0b;
    border: none;
    color: #0b1317;
    border-radius: 10px;
    font-weight: 700;
    font-size: 13.5px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    box-shadow: 0 4px 14px rgba(245, 158, 11, 0.35);
    transition: all 0.15s ease;
  }

  .btn-instalar:hover:not(:disabled) {
    background: #d97706;
    transform: translateY(-1px);
  }

  .btn-instalar:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    box-shadow: none;
  }
</style>
