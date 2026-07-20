<script>
  import { goto } from '$app/navigation';
  import { iniciarSesion, haySesion } from '$lib/sesion.svelte.js';

  let email = $state('');
  let password = $state('');
  let error = $state('');
  let entrando = $state(false);

  // Si ya hay sesión (p. ej. volvió con el botón del navegador), al inicio.
  $effect(() => {
    if (haySesion()) goto('/');
  });

  /** @param {SubmitEvent} ev */
  async function entrar(ev) {
    ev.preventDefault();
    error = '';
    entrando = true;
    try {
      await iniciarSesion(email, password);
      await goto('/');
    } catch (e) {
      const mensaje = e instanceof Error ? e.message : String(e);
      error = mensaje === 'Failed to fetch' ? 'no se pudo conectar con el servidor' : mensaje;
    } finally {
      entrando = false;
    }
  }
</script>

<div class="centro">
  <form class="card elev-sm tarjeta" onsubmit={entrar}>
    <div class="cabecera">
      <svg width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="var(--color-accent-600)" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z"></path><path d="m3.3 7 8.7 5 8.7-5"></path><path d="M12 22V12"></path></svg>
      <h2>Control de stock</h2>
      <p class="text-muted">Entrá con tu usuario del negocio.</p>
    </div>
    <div class="field">
      <label for="email">Email</label>
      <input id="email" class="input" type="email" bind:value={email} autocomplete="username" required />
    </div>
    <div class="field">
      <label for="password">Contraseña</label>
      <input id="password" class="input" type="password" bind:value={password} autocomplete="current-password" required />
    </div>
    {#if error}
      <p class="error">{error}</p>
    {/if}
    <button class="btn btn-primary btn-block entrar" type="submit" disabled={entrando || !email || !password}>
      {entrando ? 'Entrando…' : 'Entrar'}
    </button>
  </form>
</div>

<style>
  .centro {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100%;
  }
  .tarjeta {
    width: min(380px, 100%);
    gap: var(--space-3);
    padding: var(--space-6);
  }
  .cabecera {
    text-align: center;
    margin-bottom: var(--space-2);
  }
  .cabecera h2 { margin: 10px 0 4px; }
  .cabecera p { margin: 0; font-size: 13px; }
  .error {
    margin: 0;
    font-size: 13px;
    color: var(--color-peligro);
  }
  .entrar { min-height: 48px; font-size: 16px; }
</style>
