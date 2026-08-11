<script>
  import { onMount, tick } from 'svelte';
  import { fade, fly } from 'svelte/transition';
  import { goto } from '$app/navigation';
  import {
    iniciarSesion,
    haySesion,
    obtenerUsuariosRecientes,
    eliminarUsuarioReciente,
  } from '$lib/sesion.svelte.js';

  let email = $state('');
  let password = $state('');
  let recordarPassword = $state(true);
  let error = $state('');
  let entrando = $state(false);

  /** @type {Array<{ email: string, nombre: string, rol: string, password?: string, recordarPassword?: boolean, ultimaSesion: number }>} */
  let recientes = $state([]);
  /** @type {{ email: string, nombre: string, rol: string, password?: string, recordarPassword?: boolean } | null} */
  let usuarioSeleccionado = $state(null);
  let modoManual = $state(false);

  /** @type {HTMLInputElement | null} */
  let inputPassword = $state(null);

  const COLORES_ARCOIRIS = [
    'linear-gradient(135deg, #e53e3e 0%, #9b2c2c 100%)', // Rojo / Rubí
    'linear-gradient(135deg, #dd6b20 0%, #9c4221 100%)', // Naranja / Ámbar
    'linear-gradient(135deg, #d69e2e 0%, #975a16 100%)', // Amarillo / Dorado
    'linear-gradient(135deg, #38a169 0%, #22543d 100%)', // Verde Esmeralda
    'linear-gradient(135deg, #319795 0%, #234e52 100%)', // Turquesa / Acua
    'linear-gradient(135deg, #3182ce 0%, #2a4365 100%)', // Azul Zafiro
    'linear-gradient(135deg, #805ad5 0%, #44337a 100%)', // Violeta / Púrpura
    'linear-gradient(135deg, #d53f8c 0%, #702459 100%)', // Magenta / Rosa
  ];

  /** @param {string} texto */
  function obtenerColorArcoiris(texto) {
    if (!texto) return COLORES_ARCOIRIS[0];
    let hash = 0;
    for (let i = 0; i < texto.length; i++) {
      hash = texto.charCodeAt(i) + ((hash << 5) - hash);
    }
    const index = Math.abs(hash) % COLORES_ARCOIRIS.length;
    return COLORES_ARCOIRIS[index];
  }

  onMount(() => {
    cargarRecientes();
  });

  function cargarRecientes() {
    recientes = obtenerUsuariosRecientes();
    if (recientes.length > 0 && !usuarioSeleccionado && !modoManual) {
      seleccionarUsuario(recientes[0]);
    }
  }

  $effect(() => {
    if (haySesion()) goto('/');
  });

  /** @param {{ email: string, nombre: string, rol: string, password?: string, recordarPassword?: boolean }} u */
  async function seleccionarUsuario(u) {
    usuarioSeleccionado = u;
    email = u.email;
    error = '';
    if (u.recordarPassword && u.password) {
      password = u.password;
      recordarPassword = true;
    } else {
      password = '';
      recordarPassword = Boolean(u.recordarPassword);
    }
    await tick();
    if (inputPassword && !password) {
      inputPassword.focus();
    }
  }

  function cambiarAUsuarioManual() {
    usuarioSeleccionado = null;
    modoManual = true;
    email = '';
    password = '';
    recordarPassword = true;
    error = '';
  }

  function volverASelector() {
    modoManual = false;
    recientes = obtenerUsuariosRecientes();
    if (recientes.length > 0) {
      seleccionarUsuario(recientes[0]);
    }
  }

  /**
   * @param {MouseEvent} ev
   * @param {string} emailAEliminar
   */
  function quitarUsuario(ev, emailAEliminar) {
    ev.stopPropagation();
    eliminarUsuarioReciente(emailAEliminar);
    recientes = obtenerUsuariosRecientes();
    if (usuarioSeleccionado?.email === emailAEliminar) {
      if (recientes.length > 0) {
        seleccionarUsuario(recientes[0]);
      } else {
        usuarioSeleccionado = null;
        modoManual = true;
      }
    }
  }

  /**
   * @param {string} nombre
   * @param {string} emailStr
   */
  function obtenerIniciales(nombre, emailStr) {
    const fuente = nombre?.trim() || emailStr?.split('@')[0] || '?';
    const partes = fuente.split(/\s+/);
    if (partes.length >= 2) {
      return (partes[0][0] + partes[1][0]).toUpperCase();
    }
    return fuente.slice(0, 2).toUpperCase();
  }

  /** @param {SubmitEvent} ev */
  async function entrar(ev) {
    ev.preventDefault();
    error = '';
    entrando = true;
    try {
      await iniciarSesion(email, password, recordarPassword);
      await goto('/');
    } catch (e) {
      const mensaje = e instanceof Error ? e.message : String(e);
      const esErrorRed = mensaje === 'Failed to fetch' || mensaje === 'Load failed' || mensaje.toLowerCase().includes('fetch');
      error = esErrorRed ? 'no se pudo conectar con el servidor' : mensaje;
    } finally {
      entrando = false;
    }
  }
</script>

<div class="centro">
  <div class="card elev-sm tarjeta">
    <div class="cabecera">
      <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="var(--color-accent-600)" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z"></path><path d="m3.3 7 8.7 5 8.7-5"></path><path d="M12 22V12"></path></svg>
      <h2>Control de stock</h2>
      <p class="text-muted">
        {#if !modoManual && recientes.length > 0}
          Seleccioná tu perfil para ingresar
        {:else}
          Entrá con tu usuario del negocio
        {/if}
      </p>
    </div>

    <div class="cuerpo-tarjeta">
      {#if !modoManual && recientes.length > 0}
        <!-- MODO SELECTOR DE USUARIOS RECIENTES -->
        <div class="vista-login" in:fly={{ y: 8, duration: 220, delay: 90 }} out:fade={{ duration: 120 }}>
          <div class="selector-cuentas">
            <div class="grilla-usuarios">
              {#each recientes as u (u.email)}
                <div
                  class="tarjeta-usuario {usuarioSeleccionado?.email === u.email ? 'activa' : ''}"
                  role="button"
                  tabindex="0"
                  onclick={() => seleccionarUsuario(u)}
                  onkeydown={(e) => e.key === 'Enter' && seleccionarUsuario(u)}
                >
                  <button
                    type="button"
                    class="btn-quitar"
                    title="Olvidar usuario en este equipo"
                    onclick={(e) => quitarUsuario(e, u.email)}
                  >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                  </button>

                  <div class="avatar" style="background: {obtenerColorArcoiris(u.email || u.nombre)}">
                    {obtenerIniciales(u.nombre, u.email)}
                  </div>

                  <div class="info-usuario">
                    <span class="nombre-usuario" title={u.nombre}>{u.nombre}</span>
                  </div>
                </div>
              {/each}
            </div>
          </div>

          <form onsubmit={entrar} class="form-login">
            <div class="field">
              <label for="password">Contraseña para <strong>{usuarioSeleccionado?.nombre}</strong></label>
              <input
                id="password"
                bind:this={inputPassword}
                class="input"
                type="password"
                bind:value={password}
                autocomplete="current-password"
                placeholder="Ingresá tu contraseña"
                required
              />
            </div>

            <label class="checkbox-label">
              <input type="checkbox" bind:checked={recordarPassword} />
              <span>Recordar contraseña en este equipo</span>
            </label>

            {#if error}
              <p class="error">{error}</p>
            {/if}

            <button class="btn btn-primary btn-block entrar" type="submit" disabled={entrando || !password}>
              {entrando ? 'Entrando…' : 'Entrar'}
            </button>
          </form>

          <button type="button" class="btn-link" onclick={cambiarAUsuarioManual}>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
            Ingresar con otra cuenta
          </button>
        </div>
      {:else}
        <!-- MODO MANUAL (FORMULARIO EMAIL + PASSWORD) -->
        <div class="vista-login" in:fly={{ y: 8, duration: 220, delay: 90 }} out:fade={{ duration: 120 }}>
          <form onsubmit={entrar} class="form-login">
            <div class="field">
              <label for="email">Email</label>
              <input id="email" class="input" type="email" bind:value={email} autocomplete="username" placeholder="ejemplo@negocio.com" required />
            </div>
            <div class="field">
              <label for="password">Contraseña</label>
              <input id="password" class="input" type="password" bind:value={password} autocomplete="current-password" placeholder="••••••••" required />
            </div>

            <label class="checkbox-label">
              <input type="checkbox" bind:checked={recordarPassword} />
              <span>Recordar contraseña en este equipo</span>
            </label>

            {#if error}
              <p class="error">{error}</p>
            {/if}

            <button class="btn btn-primary btn-block entrar" type="submit" disabled={entrando || !email || !password}>
              {entrando ? 'Entrando…' : 'Entrar'}
            </button>
          </form>

          {#if recientes.length > 0}
            <button type="button" class="btn-link" onclick={volverASelector}>
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 18 9 12 15 6"></polyline></svg>
              Volver al selector de usuarios
            </button>
          {/if}
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .centro {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100%;
    padding: var(--space-4);
  }
  .tarjeta {
    width: min(440px, 100%);
    min-height: 480px;
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    padding: var(--space-6);
    transition: min-height 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }
  .cabecera {
    text-align: center;
    margin-bottom: var(--space-1);
  }
  .cabecera h2 { margin: 10px 0 4px; font-size: 22px; }
  .cabecera p { margin: 0; font-size: 13px; }

  .cuerpo-tarjeta {
    display: grid;
    grid-template-columns: 1fr;
    align-items: start;
    flex: 1;
  }

  .vista-login {
    grid-area: 1 / 1;
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    width: 100%;
    height: 100%;
    justify-content: space-between;
  }

  .selector-cuentas {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }
  .grilla-usuarios {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(110px, 1fr));
    gap: var(--space-3);
    max-height: 240px;
    overflow-y: auto;
    padding: 2px;
  }
  .tarjeta-usuario {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: var(--space-4) var(--space-2);
    border-radius: var(--radius-md, 8px);
    border: 2px solid var(--color-border, #e2e8f0);
    background: var(--color-bg-card, #ffffff);
    cursor: pointer;
    transition: all 0.15s ease-in-out;
  }
  .tarjeta-usuario:hover {
    border-color: var(--color-accent-400, #429398);
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  }
  .tarjeta-usuario.activa {
    border-color: var(--color-accent-600, #2e6e73);
    background: var(--color-accent-50, #f0f7f7);
    box-shadow: 0 0 0 1px var(--color-accent-600, #2e6e73);
  }

  .btn-quitar {
    position: absolute;
    top: 4px;
    right: 4px;
    width: 22px;
    height: 22px;
    border-radius: 50%;
    border: none;
    background: transparent;
    color: var(--color-text-muted, #718096);
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    opacity: 0.6;
    transition: opacity 0.15s, background 0.15s;
  }
  .tarjeta-usuario:hover .btn-quitar {
    opacity: 1;
  }
  .btn-quitar:hover {
    background: rgba(229, 62, 62, 0.15);
    color: var(--color-peligro, #e53e3e);
  }

  .avatar {
    width: 52px;
    height: 52px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 700;
    font-size: 17px;
    color: #ffffff;
    margin-bottom: var(--space-2);
    box-shadow: 0 3px 8px rgba(0, 0, 0, 0.15);
  }

  .info-usuario {
    display: flex;
    flex-direction: column;
    align-items: center;
    width: 100%;
    text-align: center;
  }
  .nombre-usuario {
    font-size: 14px;
    font-weight: 600;
    color: var(--color-text-heading, #2d3748);
    max-width: 100%;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .form-login {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .checkbox-label {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-size: 13px;
    color: var(--color-text-muted, #4a5568);
    cursor: pointer;
    user-select: none;
  }
  .checkbox-label input[type="checkbox"] {
    accent-color: var(--color-accent-600, #2e6e73);
    width: 16px;
    height: 16px;
    cursor: pointer;
  }

  .error {
    margin: 0;
    font-size: 13px;
    color: var(--color-peligro, #e53e3e);
  }

  .entrar {
    min-height: 44px;
    font-size: 15px;
    font-weight: 600;
  }

  .btn-link {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    background: none;
    border: none;
    color: var(--color-accent-600, #2e6e73);
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    padding: var(--space-2);
    border-radius: var(--radius-sm, 4px);
    transition: background 0.15s;
    margin-top: var(--space-1);
  }
  .btn-link:hover {
    background: var(--color-accent-50, #f0f7f7);
    text-decoration: underline;
  }
</style>
