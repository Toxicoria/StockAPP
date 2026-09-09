<script>
  import { onMount } from 'svelte';
  import { pedirApi } from '$lib/api.js';
  import { agregarToast } from '$lib/toast.svelte.js';

  let { abierto = $bindable(false) } = $props();

  let cargando = $state(false);
  let error = $state('');
  let qrDataUrl = $state('');
  let mobileUrl = $state('');
  let activos = $state(1);
  let maximo = $state(4);
  let copiado = $state(false);

  $effect(() => {
    if (abierto) {
      cargarQR();
    }
  });

  async function cargarQR() {
    cargando = true;
    error = '';
    copiado = false;
    try {
      const data = await pedirApi('/api/movil/generar-qr', { method: 'POST' });
      qrDataUrl = data.qr_data_url;
      mobileUrl = data.mobile_url;
      activos = data.dispositivos_activos;
      maximo = data.max_dispositivos;
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      error = msg || 'No se pudo generar el código QR de vinculación';
    } finally {
      cargando = false;
    }
  }

  function copiarEnlace() {
    if (!mobileUrl) return;
    navigator.clipboard.writeText(mobileUrl);
    copiado = true;
    agregarToast('Enlace copiado al portapapeles', 'exito');
    setTimeout(() => {
      copiado = false;
    }, 2500);
  }

  function cerrar() {
    abierto = false;
  }
</script>

{#if abierto}
  <div class="modal-backdrop" onclick={cerrar} role="presentation">
    <div class="modal-card" onclick={(e) => e.stopPropagation()} role="dialog" aria-modal="true">
      <div class="modal-header">
        <div class="header-tit">
          <div class="icono-movil">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="var(--color-accent-600)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="5" y="2" width="14" height="20" rx="2" ry="2"></rect>
              <line x1="12" y1="18" x2="12.01" y2="18"></line>
            </svg>
          </div>
          <div>
            <h3>Supervisión Móvil en Vivo</h3>
            <p class="subtit">Monitoreo de caja, ventas y stock para el dueño</p>
          </div>
        </div>
        <button class="btn-cerrar" onclick={cerrar} title="Cerrar">✕</button>
      </div>

      <div class="modal-body">
        {#if cargando}
          <div class="cargando-box">
            <div class="spinner"></div>
            <p>Generando código QR seguro...</p>
          </div>
        {:else if error}
          <div class="error-box">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
            <p>{error}</p>
            <button class="btn btn-secondary" onclick={cargarQR} style="margin-top: 8px;">Reintentar</button>
          </div>
        {:else}
          <p class="instruccion">
            Escaneá este código QR con la <strong>cámara de tu celular</strong> para abrir el panel de control en tiempo real:
          </p>

          <div class="qr-marco">
            {#if qrDataUrl}
              <img src={qrDataUrl} alt="Código QR para acceso móvil" class="qr-img" />
            {/if}
          </div>

          <div class="info-cupos">
            <div class="cupo-badge">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect><line x1="8" y1="21" x2="16" y2="21"></line><line x1="12" y1="17" x2="12" y2="21"></line></svg>
              <span>{activos} de {maximo} equipos conectados simultáneamente</span>
            </div>
            <p class="aclaracion-seguridad">
              🔒 <strong>Solo Lectura:</strong> Desde el celular no se pueden realizar cobros ni registrar ventas. Solo es para observar lo que ocurre en el negocio.
            </p>
          </div>

          <div class="acciones-qr">
            <button class="btn btn-secondary btn-copiar" onclick={copiarEnlace}>
              {#if copiado}
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"></polyline></svg>
                <span>¡Enlace Copiado!</span>
              {:else}
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
                <span>Copiar Enlace Directo</span>
              {/if}
            </button>
            <button class="btn btn-secondary" onclick={cargarQR} title="Regenerar QR">
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 2v6h-6"></path><path d="M3 12a9 9 0 0 1 15-6.7L21 8"></path><path d="M3 22v-6h6"></path><path d="M21 12a9 9 0 0 1-15 6.7L3 16"></path></svg>
            </button>
          </div>
        {/if}
      </div>

      <div class="modal-footer">
        <button class="btn btn-primary" onclick={cerrar}>Cerrar</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background: rgba(15, 23, 42, 0.55);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 999;
    padding: var(--space-3);
  }

  .modal-card {
    background: #ffffff;
    border: 1px solid var(--color-divider);
    border-radius: var(--radius-lg);
    width: 100%;
    max-width: 440px;
    box-shadow: var(--shadow-lg);
    overflow: hidden;
    animation: aparecer 0.2s ease-out;
  }

  @keyframes aparecer {
    from { opacity: 0; transform: scale(0.96) translateY(6px); }
    to { opacity: 1; transform: scale(1) translateY(0); }
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    border-bottom: 1px solid var(--color-divider);
    background: var(--color-surface);
  }

  .header-tit {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .icono-movil {
    width: 38px;
    height: 38px;
    border-radius: var(--radius-sm);
    background: var(--color-accent-100);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .header-tit h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 700;
    color: var(--color-text);
  }

  .subtit {
    margin: 0;
    font-size: 12px;
    color: var(--color-text-muted);
  }

  .btn-cerrar {
    background: none;
    border: none;
    font-size: 18px;
    color: var(--color-text-muted);
    cursor: pointer;
    padding: 4px;
    border-radius: 4px;
  }
  .btn-cerrar:hover { color: var(--color-text); }

  .modal-body {
    padding: 20px;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
  }

  .instruccion {
    font-size: 13.5px;
    color: var(--color-text);
    margin-bottom: 16px;
    line-height: 1.4;
  }

  .qr-marco {
    background: #ffffff;
    border: 2px solid var(--color-accent-300);
    border-radius: var(--radius-md);
    padding: 10px;
    box-shadow: var(--shadow-sm);
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 14px;
  }

  .qr-img {
    width: 210px;
    height: 210px;
    display: block;
    border-radius: var(--radius-sm);
  }

  .info-cupos {
    width: 100%;
    margin-bottom: 14px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .cupo-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    background: var(--color-surface);
    border: 1px solid var(--color-divider);
    padding: 6px 12px;
    border-radius: 20px;
    font-size: 12px;
    font-weight: 600;
    color: var(--color-text);
  }

  .aclaracion-seguridad {
    font-size: 12px;
    color: var(--color-text-muted);
    line-height: 1.4;
  }

  .acciones-qr {
    display: flex;
    gap: 8px;
    width: 100%;
    justify-content: center;
  }

  .btn-copiar {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    font-size: 13px;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    padding: 12px 20px;
    background: var(--color-surface);
    border-top: 1px solid var(--color-divider);
  }

  .cargando-box {
    padding: 40px 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    color: var(--color-text-muted);
    font-size: 13.5px;
  }

  .spinner {
    width: 32px;
    height: 32px;
    border: 3px solid var(--color-divider);
    border-top-color: var(--color-accent-600);
    border-radius: 50%;
    animation: giro 0.8s linear infinite;
  }

  @keyframes giro {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }

  .error-box {
    padding: 24px 0;
    color: var(--color-peligro);
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    font-size: 13px;
  }
</style>
