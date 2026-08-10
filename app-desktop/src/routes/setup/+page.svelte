<script>
  import { goto } from '$app/navigation';
  import { BASE_API } from '$lib/config.js';

  let paso = $state(1);
  let enviando = $state(false);
  let error = $state('');

  // Paso 1 — Datos del negocio
  let nombreNegocio = $state('');
  let direccion = $state('');
  let cuit = $state('');
  let nombreDueno = $state('');
  let telefono = $state('');
  let emailNegocio = $state('');

  // Paso 2 — Cuenta del dueño
  let emailLogin = $state('');
  let password = $state('');
  let passwordConfirm = $state('');

  const paso1Valido = $derived(nombreNegocio.trim() !== '' && nombreDueno.trim() !== '');
  const paso2Valido = $derived(
    emailLogin.trim() !== '' &&
    password.length >= 8 &&
    password === passwordConfirm
  );

  function irAPaso2() {
    if (!paso1Valido) return;
    error = '';
    paso = 2;
  }

  /** @param {SubmitEvent} ev */
  async function registrar(ev) {
    ev.preventDefault();
    if (!paso2Valido) return;

    error = '';
    enviando = true;
    try {
      const resp = await fetch(`${BASE_API}/api/registro`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          nombre_negocio: nombreNegocio.trim(),
          direccion: direccion.trim(),
          cuit: cuit.trim(),
          nombre_dueno: nombreDueno.trim(),
          telefono: telefono.trim(),
          email_negocio: emailNegocio.trim(),
          nombre_admin: nombreDueno.trim(),
          email_admin: emailLogin.trim(),
          password: password,
        }),
      });
      const datos = await resp.json();
      if (!resp.ok) throw new Error(datos.error ?? 'no se pudo registrar');
      await goto('/login');
    } catch (e) {
      const mensaje = e instanceof Error ? e.message : String(e);
      error = mensaje === 'Failed to fetch' ? 'no se pudo conectar con el servidor' : mensaje;
    } finally {
      enviando = false;
    }
  }
</script>

<div class="centro">
  <div class="card elev-sm tarjeta">
    <!-- Indicador de pasos -->
    <div class="pasos">
      <div class="paso-indicador" class:activo={paso >= 1}>
        <span class="paso-num">1</span>
        <span class="paso-label">Negocio</span>
      </div>
      <div class="paso-linea" class:activo={paso >= 2}></div>
      <div class="paso-indicador" class:activo={paso >= 2}>
        <span class="paso-num">2</span>
        <span class="paso-label">Cuenta</span>
      </div>
    </div>

    {#if paso === 1}
      <div class="cabecera">
        <h2>Tu negocio</h2>
        <p class="text-muted">Contanos sobre tu negocio para configurar el sistema.</p>
      </div>

      <div class="field">
        <label for="nombre_negocio">Nombre del negocio *</label>
        <input id="nombre_negocio" class="input" type="text" bind:value={nombreNegocio} placeholder="Ej: Almacén Don Pedro" required />
      </div>
      <div class="field">
        <label for="nombre_dueno">Nombre del dueño/a *</label>
        <input id="nombre_dueno" class="input" type="text" bind:value={nombreDueno} placeholder="Ej: Pedro García" required />
      </div>
      <div class="field">
        <label for="direccion">Dirección</label>
        <input id="direccion" class="input" type="text" bind:value={direccion} placeholder="Ej: Av. San Martín 1234" />
      </div>
      <div class="fila-doble">
        <div class="field">
          <label for="cuit">CUIT</label>
          <input id="cuit" class="input" type="text" bind:value={cuit} placeholder="20-12345678-9" />
        </div>
        <div class="field">
          <label for="telefono">Teléfono</label>
          <input id="telefono" class="input" type="tel" bind:value={telefono} placeholder="Ej: 2944-123456" />
        </div>
      </div>
      <div class="field">
        <label for="email_negocio">Email del negocio</label>
        <input id="email_negocio" class="input" type="email" bind:value={emailNegocio} placeholder="contacto@minegocio.com" />
      </div>

      <button class="btn btn-primary btn-block boton-grande" disabled={!paso1Valido} onclick={irAPaso2}>
        Siguiente →
      </button>
    {:else}
      <div class="cabecera">
        <h2>Tu cuenta</h2>
        <p class="text-muted">Creá tu usuario para acceder al sistema.</p>
      </div>

      <form onsubmit={registrar}>
        <div class="campos-paso2">
          <div class="field">
            <label for="email_login">Email para iniciar sesión *</label>
            <input id="email_login" class="input" type="email" bind:value={emailLogin} autocomplete="email" required />
          </div>
          <div class="field">
            <label for="pass">Contraseña * <span class="text-muted">(mínimo 8 caracteres)</span></label>
            <input id="pass" class="input" type="password" bind:value={password} autocomplete="new-password" required />
          </div>
          <div class="field">
            <label for="pass_confirm">Confirmar contraseña *</label>
            <input id="pass_confirm" class="input" type="password" bind:value={passwordConfirm} autocomplete="new-password" required />
            {#if passwordConfirm && password !== passwordConfirm}
              <span class="error-inline">Las contraseñas no coinciden</span>
            {/if}
          </div>
        </div>

        {#if error}
          <p class="error">{error}</p>
        {/if}

        <div class="acciones">
          <button type="button" class="btn btn-secondary" onclick={() => { paso = 1; error = ''; }}>
            ← Volver
          </button>
          <button type="submit" class="btn btn-primary boton-registrar" disabled={enviando || !paso2Valido}>
            {enviando ? 'Creando…' : 'Crear negocio'}
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
    min-height: 100%;
    padding: var(--space-5);
  }
  .tarjeta {
    width: min(480px, 100%);
    gap: var(--space-3);
    padding: var(--space-6);
  }
  .cabecera {
    text-align: center;
    margin-bottom: var(--space-2);
  }
  .cabecera h2 { margin: 0 0 6px; }
  .cabecera p { margin: 0; font-size: 13px; }

  /* Indicador de pasos */
  .pasos {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0;
    margin-bottom: var(--space-3);
  }
  .paso-indicador {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    opacity: 0.4;
    transition: opacity 0.2s;
  }
  .paso-indicador.activo { opacity: 1; }
  .paso-num {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    border-radius: 50%;
    background: var(--color-surface);
    font-size: 12px;
    font-weight: 600;
    color: var(--color-text);
  }
  .paso-indicador.activo .paso-num {
    background: var(--color-accent);
    color: #fff;
  }
  .paso-label {
    font-size: 11px;
    color: color-mix(in srgb, var(--color-text) 60%, transparent);
  }
  .paso-linea {
    width: 60px;
    height: 2px;
    background: var(--color-divider);
    margin: 0 var(--space-3);
    margin-bottom: 18px;
    transition: background 0.2s;
  }
  .paso-linea.activo { background: var(--color-accent); }

  .fila-doble {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-3);
  }

  .campos-paso2 {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    margin-bottom: var(--space-4);
  }

  .boton-grande { min-height: 46px; font-size: 15px; margin-top: var(--space-2); }

  .acciones {
    display: flex;
    gap: var(--space-3);
    margin-top: var(--space-2);
  }
  .boton-registrar { flex: 1; min-height: 46px; font-size: 15px; }

  .error {
    margin: 0;
    font-size: 13px;
    color: var(--color-peligro);
  }
  .error-inline {
    font-size: 12px;
    color: var(--color-peligro);
  }
</style>
