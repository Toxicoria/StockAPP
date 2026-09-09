<script>
  import { estadoUpdater, reiniciarYAplicar, iniciarDescargaEInstalacion } from '$lib/updater.svelte.js';
  import { solicitarCerrarApp } from '$lib/confirmacion.svelte.js';
</script>

{#if estadoUpdater.disponible && estadoUpdater.esObligatoria}
  <div class="bloqueo-fondo">
    <div class="tarjeta-modal" role="alertdialog" aria-modal="true">
      <div class="cabecera">
        <div class="icono-alerta">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
            <line x1="12" y1="9" x2="12" y2="13"></line>
            <line x1="12" y1="17" x2="12.01" y2="17"></line>
          </svg>
        </div>
        <div class="titulos">
          <div class="fila-titulo">
            <h2>Actualización Requerida</h2>
            <span class="badge-version">v{estadoUpdater.versionNueva}</span>
          </div>
          <span class="subtitulo text-muted">Es necesario actualizar para continuar utilizando StockAPP.</span>
        </div>
      </div>

      {#if estadoUpdater.motivo}
        <div class="motivo-caja">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"></circle>
            <line x1="12" y1="16" x2="12" y2="12"></line>
            <line x1="12" y1="8" x2="12.01" y2="8"></line>
          </svg>
          <span><strong>Motivo:</strong> {estadoUpdater.motivo}</span>
        </div>
      {/if}

      {#if estadoUpdater.notas}
        <div class="notas-caja">
          <span class="notas-titulo">Novedades de esta versión:</span>
          <p class="notas-texto">{estadoUpdater.notas}</p>
        </div>
      {/if}

      <!-- PROGRESO (SOLO VISIBLE CUANDO COMIENZA LA DESCARGA O FINALIZA) -->
      {#if estadoUpdater.descargando || estadoUpdater.descargada}
        <div class="progreso-contenedor">
          <div class="progreso-etiquetas">
            <span class="estado-texto">
              {#if estadoUpdater.descargada}
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="var(--color-accent-600)" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                <span>Descarga completa. Lista para instalar.</span>
              {:else}
                <svg class="anim-girar" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M21 12a9 9 0 1 1-6.219-8.56"></path>
                </svg>
                <span>Descargando actualización...</span>
              {/if}
            </span>
            <span class="porcentaje">{estadoUpdater.progreso}%</span>
          </div>

          <div class="barra-track">
            <div class="barra-fill" style="width: {estadoUpdater.progreso}%;"></div>
          </div>
        </div>
      {/if}

      {#if estadoUpdater.error}
        <div class="alerta-error">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"></circle>
            <line x1="12" y1="8" x2="12" y2="12"></line>
            <line x1="12" y1="16" x2="12.01" y2="16"></line>
          </svg>
          <span class="error-texto">{estadoUpdater.error}</span>
          <button type="button" class="btn-reintentar" onclick={iniciarDescargaEInstalacion}>Reintentar</button>
        </div>
      {/if}

      <!-- ACCIONES -->
      <div class="acciones-pie">
        <button type="button" class="btn btn-secondary" onclick={solicitarCerrarApp}>
          Salir del sistema
        </button>

        {#if !estadoUpdater.descargando && !estadoUpdater.descargada}
          <button
            type="button"
            class="btn btn-primary btn-accion"
            onclick={iniciarDescargaEInstalacion}
          >
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
              <polyline points="7 10 12 15 17 10"></polyline>
              <line x1="12" y1="15" x2="12" y2="3"></line>
            </svg>
            <span>Actualizar</span>
          </button>
        {:else if estadoUpdater.descargando}
          <button type="button" class="btn btn-primary btn-accion" disabled>
            <svg class="anim-girar" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 12a9 9 0 1 1-6.219-8.56"></path>
            </svg>
            <span>Descargando ({estadoUpdater.progreso}%)...</span>
          </button>
        {:else}
          <button
            type="button"
            class="btn btn-primary btn-accion"
            onclick={reiniciarYAplicar}
          >
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="23 4 23 10 17 10"></polyline>
              <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"></path>
            </svg>
            <span>Reiniciar e Instalar</span>
          </button>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .bloqueo-fondo {
    position: fixed;
    inset: 0;
    background: rgba(15, 41, 44, 0.72);
    backdrop-filter: blur(8px);
    z-index: 9999;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--space-4);
    user-select: none;
  }

  .tarjeta-modal {
    background: #fbfcfc;
    border: 1px solid var(--color-divider);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-lg);
    max-width: 500px;
    width: 100%;
    padding: var(--space-5);
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    animation: aparecer 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes aparecer {
    from { opacity: 0; transform: scale(0.97); }
    to { opacity: 1; transform: scale(1); }
  }

  .cabecera {
    display: flex;
    align-items: flex-start;
    gap: var(--space-3);
  }

  .icono-alerta {
    width: 42px;
    height: 42px;
    min-width: 42px;
    border-radius: 50%;
    background: color-mix(in srgb, var(--color-accent) 12%, transparent);
    color: var(--color-accent-600);
    border: 1px solid var(--color-accent-200);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .titulos {
    display: flex;
    flex-direction: column;
    gap: 3px;
  }

  .fila-titulo {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }

  h2 {
    font-family: var(--font-heading);
    font-size: 18px;
    font-weight: 700;
    color: var(--color-text);
    margin: 0;
  }

  .badge-version {
    background: var(--color-accent-100);
    color: var(--color-accent-700);
    border: 1px solid var(--color-accent-300);
    font-family: monospace;
    font-weight: 700;
    font-size: 12px;
    padding: 2px 8px;
    border-radius: var(--radius-sm);
  }

  .subtitulo {
    font-size: 13px;
    line-height: 1.4;
  }

  .motivo-caja {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: 8px 12px;
    background: color-mix(in srgb, var(--color-accent) 8%, transparent);
    border: 1px solid var(--color-accent-200);
    border-radius: var(--radius-sm);
    font-size: 13px;
    color: var(--color-accent-800);
  }

  .notas-caja {
    width: 100%;
    background: var(--color-surface);
    border: 1px solid var(--color-divider);
    border-radius: var(--radius-md);
    padding: var(--space-3) var(--space-4);
    display: flex;
    flex-direction: column;
    gap: 4px;
    max-height: 130px;
    overflow-y: auto;
  }

  .notas-titulo {
    font-size: 12px;
    font-weight: 600;
    color: var(--color-text);
  }

  .notas-texto {
    font-size: 13px;
    color: color-mix(in srgb, var(--color-text) 80%, transparent);
    line-height: 1.45;
    margin: 0;
  }

  /* BARRA DE PROGRESO */
  .progreso-contenedor {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin-top: 2px;
  }

  .progreso-etiquetas {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 12px;
    color: var(--color-text);
  }

  .estado-texto {
    display: flex;
    align-items: center;
    gap: 6px;
    color: color-mix(in srgb, var(--color-text) 80%, transparent);
  }

  .porcentaje {
    font-weight: 700;
    color: var(--color-accent-600);
    font-family: monospace;
  }

  .barra-track {
    width: 100%;
    height: 8px;
    background: var(--color-surface);
    border: 1px solid var(--color-divider);
    border-radius: 99px;
    overflow: hidden;
  }

  .barra-fill {
    height: 100%;
    background: var(--color-accent-600);
    border-radius: 99px;
    transition: width 0.3s ease;
  }

  .alerta-error {
    background: color-mix(in srgb, var(--color-peligro) 10%, transparent);
    border: 1px solid color-mix(in srgb, var(--color-peligro) 30%, transparent);
    color: var(--color-peligro);
    padding: 8px 12px;
    border-radius: var(--radius-sm);
    font-size: 12px;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .error-texto {
    flex: 1;
  }

  .btn-reintentar {
    background: var(--color-peligro);
    border: none;
    color: #ffffff;
    padding: 4px 10px;
    border-radius: var(--radius-sm);
    cursor: pointer;
    font-size: 12px;
    font-weight: 600;
  }

  /* ACCIONES */
  .acciones-pie {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-2);
    margin-top: var(--space-2);
    padding-top: var(--space-3);
    border-top: 1px solid var(--color-divider);
  }

  .btn-accion {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .anim-girar {
    animation: girar 1s linear infinite;
  }

  @keyframes girar {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }
</style>
