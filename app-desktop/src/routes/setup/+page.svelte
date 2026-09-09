<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { pedirApi } from '$lib/api.js';
  import { sesion, haySesion, cerrarSesion } from '$lib/sesion.svelte.js';
  import { agregarToast } from '$lib/toast.svelte.js';

  let cargandoInicial = $state(true);
  let guardando = $state(false);
  let error = $state('');

  // Campos del negocio
  let nombreNegocio = $state('');
  let nombreDueno = $state('');
  let direccion = $state('');
  let telefono = $state('');
  let emailNegocio = $state('');

  const formularioValido = $derived(
    nombreNegocio.trim() !== '' &&
    nombreDueno.trim() !== '' &&
    direccion.trim() !== '' &&
    telefono.trim() !== ''
  );

  onMount(async () => {
    if (!haySesion()) {
      await goto('/login');
      return;
    }

    try {
      const datos = await pedirApi('/api/negocio');
      if (datos) {
        nombreNegocio = datos.nombre_negocio || '';
        nombreDueno = datos.nombre_dueno || sesion.nombre || '';
        direccion = datos.direccion || '';
        telefono = datos.telefono || '';
        emailNegocio = datos.email_negocio || '';

        // Si el nombre de negocio era el default generado "Negocio de ...", lo limpiamos para que el usuario ingrese el nombre de fantasía real
        if (nombreNegocio.startsWith('Negocio de ')) {
          nombreNegocio = '';
        }

        // Si ya tenía todo completo, no hace falta que esté acá
        if (nombreNegocio.trim() && nombreDueno.trim() && direccion.trim() && telefono.trim()) {
          sesion.perfilCompleto = true;
          await goto('/');
          return;
        }
      }
    } catch (e) {
      console.warn('Error precargando datos del negocio:', e);
    } finally {
      cargandoInicial = false;
    }
  });

  /** @param {SubmitEvent} ev */
  async function guardarConfiguracion(ev) {
    ev.preventDefault();
    if (!formularioValido) {
      error = 'Por favor completá los campos obligatorios marcados con asterisco (*).';
      return;
    }

    error = '';
    guardando = true;

    try {
      await pedirApi('/api/negocio', {
        method: 'PUT',
        body: {
          nombre_negocio: nombreNegocio.trim(),
          nombre_dueno: nombreDueno.trim(),
          direccion: direccion.trim(),
          telefono: telefono.trim(),
          email_negocio: emailNegocio.trim(),
        },
      });

      sesion.negocio = nombreNegocio.trim();
      sesion.perfilCompleto = true;
      agregarToast('¡Negocio configurado con éxito!', 'exito');
      await goto('/');
    } catch (e) {
      const mensaje = e instanceof Error ? e.message : String(e);
      error = mensaje || 'No se pudieron guardar los datos del negocio.';
    } finally {
      guardando = false;
    }
  }
</script>

<div class="centro">
  <div class="card elev-sm tarjeta-setup">
    {#if cargandoInicial}
      <div class="cargando-box">
        <svg class="anim-girar" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="var(--color-accent-600)" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 12a9 9 0 1 1-6.219-8.56"></path>
        </svg>
        <span class="text-muted">Cargando configuración...</span>
      </div>
    {:else}
      <div class="cabecera">
        <div class="icono-box">
          <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="var(--color-accent-600)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="m2 7 4.41-4.41A2 2 0 0 1 7.83 2h8.34a2 2 0 0 1 1.42.59L22 7"></path>
            <path d="M4 12v8a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-8"></path>
            <path d="M15 22v-4a2 2 0 0 0-2-2h-2a2 2 0 0 0-2 2v4"></path>
            <path d="M2 7h20"></path>
          </svg>
        </div>
        <div class="titulos">
          <h2>Configuración Inicial de tu Negocio</h2>
          <p class="text-muted">Completá los datos comerciales y de contacto para comenzar a operar con StockAPP.</p>
        </div>
      </div>

      {#if error}
        <div class="alerta-error">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"></circle>
            <line x1="12" y1="8" x2="12" y2="12"></line>
            <line x1="12" y1="16" x2="12.01" y2="16"></line>
          </svg>
          <span>{error}</span>
        </div>
      {/if}

      <form onsubmit={guardarConfiguracion} class="formulario">
        <div class="field">
          <label for="nombre_negocio">Nombre de fantasía / del negocio *</label>
          <input
            id="nombre_negocio"
            class="input"
            type="text"
            bind:value={nombreNegocio}
            placeholder="Ej: Kiosco Don Pedro"
            required
          />
          <small class="hint">Es el nombre visible en la barra de título, tickets y reportes.</small>
        </div>

        <div class="field">
          <label for="nombre_dueno">Nombre del dueño/a o responsable *</label>
          <input
            id="nombre_dueno"
            class="input"
            type="text"
            bind:value={nombreDueno}
            placeholder="Ej: Pedro García"
            required
          />
        </div>

        <div class="fila-doble">
          <div class="field">
            <label for="direccion">Dirección comercial *</label>
            <input
              id="direccion"
              class="input"
              type="text"
              bind:value={direccion}
              placeholder="Ej: Av. San Martín 1234"
              required
            />
          </div>

          <div class="field">
            <label for="telefono">Teléfono de contacto *</label>
            <input
              id="telefono"
              class="input"
              type="text"
              bind:value={telefono}
              placeholder="Ej: 2966-456789"
              required
            />
          </div>
        </div>

        <div class="field">
          <label for="email_negocio">Correo electrónico del negocio <span class="opcional">(opcional)</span></label>
          <input
            id="email_negocio"
            class="input"
            type="email"
            bind:value={emailNegocio}
            placeholder="Ej: contacto@kioscodonpedro.com"
          />
        </div>

        <div class="info-nota">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"></circle>
            <line x1="12" y1="16" x2="12" y2="12"></line>
            <line x1="12" y1="8" x2="12.01" y2="8"></line>
          </svg>
          <span>Los datos de facturación electrónica y AFIP se podrán configurar más adelante desde el menú de Configuración.</span>
        </div>

        <div class="acciones">
          <button
            type="button"
            class="btn btn-secondary"
            onclick={async () => {
              await cerrarSesion();
              await goto('/login');
            }}
          >
            Cerrar sesión
          </button>
          <button
            type="submit"
            class="btn btn-primary btn-guardar"
            disabled={guardando || !formularioValido}
          >
            {#if guardando}
              <svg class="anim-girar" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 12a9 9 0 1 1-6.219-8.56"></path>
              </svg>
              <span>Guardando configuración...</span>
            {:else}
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="20 6 9 17 4 12"></polyline>
              </svg>
              <span>Guardar y Comenzar</span>
            {/if}
          </button>
        </div>
      </form>
    {/if}
  </div>
</div>

<style>
  .centro {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: calc(100vh - 40px);
    padding: var(--space-4);
    background: var(--color-bg);
  }

  .tarjeta-setup {
    width: 100%;
    max-width: 540px;
    padding: var(--space-5);
    background: #fbfcfc;
    border: 1px solid var(--color-divider);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-md);
  }

  .cargando-box {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-3);
    padding: var(--space-6) 0;
  }

  .cabecera {
    display: flex;
    align-items: flex-start;
    gap: var(--space-3);
    margin-bottom: var(--space-4);
    padding-bottom: var(--space-3);
    border-bottom: 1px solid var(--color-divider);
  }

  .icono-box {
    width: 44px;
    height: 44px;
    min-width: 44px;
    border-radius: 50%;
    background: var(--color-accent-100);
    border: 1px solid var(--color-accent-300);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .titulos h2 {
    font-size: 19px;
    color: var(--color-text);
    margin: 0 0 4px 0;
  }

  .titulos p {
    font-size: 13px;
    margin: 0;
    line-height: 1.4;
  }

  .formulario {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .field label {
    font-size: 13px;
    font-weight: 600;
    color: var(--color-text);
  }

  .opcional {
    font-weight: 400;
    color: color-mix(in srgb, var(--color-text) 60%, transparent);
    font-size: 12px;
  }

  .hint {
    font-size: 11.5px;
    color: color-mix(in srgb, var(--color-text) 60%, transparent);
  }

  .fila-doble {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-3);
  }

  @media (max-width: 520px) {
    .fila-doble {
      grid-template-columns: 1fr;
    }
  }

  .input {
    width: 100%;
    height: 38px;
    padding: 0 12px;
    border: 1px solid var(--color-divider);
    border-radius: var(--radius-sm);
    font-family: inherit;
    font-size: 13.5px;
    background: #ffffff;
    color: var(--color-text);
    transition: border-color 0.15s ease;
  }

  .input:focus {
    outline: none;
    border-color: var(--color-accent-600);
  }

  .info-nota {
    display: flex;
    align-items: flex-start;
    gap: var(--space-2);
    padding: 10px 14px;
    background: var(--color-surface);
    border: 1px solid var(--color-divider);
    border-radius: var(--radius-sm);
    font-size: 12.5px;
    color: color-mix(in srgb, var(--color-text) 80%, transparent);
    line-height: 1.4;
    margin-top: var(--space-1);
  }

  .info-nota svg {
    min-width: 18px;
    color: var(--color-accent-600);
    margin-top: 1px;
  }

  .alerta-error {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 14px;
    background: color-mix(in srgb, var(--color-peligro) 10%, transparent);
    border: 1px solid color-mix(in srgb, var(--color-peligro) 30%, transparent);
    color: var(--color-peligro);
    border-radius: var(--radius-sm);
    font-size: 13px;
    margin-bottom: var(--space-3);
  }

  .acciones {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: var(--space-3);
    margin-top: var(--space-2);
    padding-top: var(--space-3);
    border-top: 1px solid var(--color-divider);
  }

  .btn-guardar {
    height: 42px;
    padding: 0 22px;
    font-size: 14px;
    font-weight: 600;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .anim-girar {
    animation: girar 1s linear infinite;
  }

  @keyframes girar {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }
</style>
