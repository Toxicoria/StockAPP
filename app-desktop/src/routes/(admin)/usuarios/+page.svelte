<script>
  import { pedirApi } from '$lib/api.js';
  import { esDueno } from '$lib/sesion.svelte.js';
  import BotonVolver from '$lib/componentes/BotonVolver.svelte';
  import Dialogo from '$lib/componentes/Dialogo.svelte';

  /** @type {any[]} */
  let usuarios = $state([]);
  let cargando = $state(true);
  let error = $state('');

  // Modal
  let modalAbierto = $state(false);
  let modoEdicion = $state(false);
  let editandoId = $state(0);
  let formNombre = $state('');
  let formEmail = $state('');
  let formPassword = $state('');
  let guardando = $state(false);
  let errorModal = $state('');

  // Confirmación de eliminar
  let confirmandoId = $state(0);
  let confirmandoNombre = $state('');
  let eliminando = $state(false);

  async function cargar() {
    cargando = true;
    error = '';
    try {
      usuarios = await pedirApi('/api/usuarios');
    } catch (e) {
      error = e instanceof Error ? e.message : 'error cargando usuarios';
    } finally {
      cargando = false;
    }
  }

  function abrirNuevo() {
    modoEdicion = false;
    editandoId = 0;
    formNombre = '';
    formEmail = '';
    formPassword = '';
    errorModal = '';
    modalAbierto = true;
  }

  /** @param {any} u */
  function abrirEditar(u) {
    modoEdicion = true;
    editandoId = u.id_usuario;
    formNombre = u.nombre;
    formEmail = u.email;
    formPassword = '';
    errorModal = '';
    modalAbierto = true;
  }

  /** @param {SubmitEvent} ev */
  async function guardar(ev) {
    ev.preventDefault();
    guardando = true;
    errorModal = '';
    try {
      if (modoEdicion) {
        await pedirApi(`/api/usuarios/${editandoId}`, {
          method: 'PUT',
          body: { nombre: formNombre, email: formEmail },
        });
      } else {
        await pedirApi('/api/usuarios', {
          method: 'POST',
          body: { nombre: formNombre, email: formEmail, password: formPassword },
        });
      }
      modalAbierto = false;
      await cargar();
    } catch (e) {
      errorModal = e instanceof Error ? e.message : 'error guardando';
    } finally {
      guardando = false;
    }
  }

  /**
   * @param {number} id
   * @param {string} nombre
   */
  function pedirConfirmacion(id, nombre) {
    confirmandoId = id;
    confirmandoNombre = nombre;
  }

  async function confirmarEliminar() {
    eliminando = true;
    try {
      await pedirApi(`/api/usuarios/${confirmandoId}`, { method: 'DELETE' });
      confirmandoId = 0;
      await cargar();
    } catch (e) {
      error = e instanceof Error ? e.message : 'error eliminando';
    } finally {
      eliminando = false;
    }
  }

  /** @param {string} iso */
  function formatearFecha(iso) {
    return new Date(iso).toLocaleDateString('es-AR', { day: '2-digit', month: '2-digit', year: 'numeric' });
  }

  $effect(() => {
    if (esDueno()) cargar();
  });
</script>

<div class="pagina">
  <div class="cabecera">
    <div>
      <BotonVolver />
      <h2>Usuarios</h2>
      <p class="text-muted">Gestioná los empleados que acceden al sistema.</p>
    </div>
    <button class="btn btn-primary" onclick={abrirNuevo}>
      + Nuevo cajero
    </button>
  </div>

  {#if cargando}
    <p class="text-muted">Cargando…</p>
  {:else if error}
    <p class="error">{error}</p>
  {:else if usuarios.length === 0}
    <div class="vacio">
      <p class="text-muted">No hay empleados registrados todavía.</p>
      <button class="btn btn-primary" onclick={abrirNuevo}>Crear el primero</button>
    </div>
  {:else}
    <table class="table">
      <thead>
        <tr>
          <th>Nombre</th>
          <th>Email</th>
          <th>Rol</th>
          <th>Alta</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#each usuarios as u (u.id_usuario)}
          <tr>
            <td>{u.nombre}</td>
            <td>{u.email}</td>
            <td>
              <span class="tag" class:tag-accent={u.rol === 'dueño'} class:tag-neutral={u.rol === 'cajero'}>
                {u.rol}
              </span>
            </td>
            <td class="text-muted">{formatearFecha(u.fecha_alta)}</td>
            <td class="acciones-fila">
              {#if u.rol !== 'dueño'}
                <button class="btn btn-ghost btn-sm" onclick={() => abrirEditar(u)}>Editar</button>
                <button class="btn btn-ghost btn-sm btn-peligro" onclick={() => pedirConfirmacion(u.id_usuario, u.nombre)}>Eliminar</button>
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>

<!-- Modal crear/editar -->
<Dialogo abierto={modalAbierto}>
  <span class="dialog-title">{modoEdicion ? 'Editar usuario' : 'Nuevo cajero'}</span>
  <form onsubmit={guardar}>
    <div class="campos-modal">
      <div class="field">
        <label for="m_nombre">Nombre</label>
        <input id="m_nombre" class="input" type="text" bind:value={formNombre} required />
      </div>
      <div class="field">
        <label for="m_email">Email</label>
        <input id="m_email" class="input" type="email" bind:value={formEmail} required />
      </div>
      {#if !modoEdicion}
        <div class="field">
          <label for="m_pass">Contraseña <span class="text-muted">(mínimo 8 caracteres)</span></label>
          <input id="m_pass" class="input" type="password" bind:value={formPassword} minlength="8" required />
        </div>
      {/if}
    </div>
    {#if errorModal}
      <p class="error">{errorModal}</p>
    {/if}
    <div class="dialog-actions">
      <button type="button" class="btn btn-secondary" onclick={() => (modalAbierto = false)}>Cancelar</button>
      <button type="submit" class="btn btn-primary" disabled={guardando}>
        {guardando ? 'Guardando…' : modoEdicion ? 'Guardar' : 'Crear'}
      </button>
    </div>
  </form>
</Dialogo>

<!-- Confirmación de eliminar -->
<Dialogo abierto={confirmandoId > 0} ancho="360px">
  <span class="dialog-title">Eliminar usuario</span>
  <p class="dialog-body">
    ¿Seguro que querés eliminar a <strong>{confirmandoNombre}</strong>? Esta acción no se puede deshacer.
  </p>
  <div class="dialog-actions">
    <button class="btn btn-secondary" onclick={() => (confirmandoId = 0)}>Cancelar</button>
    <button class="btn btn-peligro-fill" disabled={eliminando} onclick={confirmarEliminar}>
      {eliminando ? 'Eliminando…' : 'Eliminar'}
    </button>
  </div>
</Dialogo>

<style>
  .pagina { max-width: 800px; }
  .cabecera {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    margin-bottom: var(--space-5);
  }
  .cabecera h2 { margin: var(--space-2) 0 4px; }
  .cabecera p { margin: 0; font-size: 13px; }
  .error { color: var(--color-peligro); font-size: 13px; }
  .vacio {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-6) 0;
  }
  .acciones-fila {
    display: flex;
    gap: var(--space-1);
    justify-content: flex-end;
  }
  .btn-sm { font-size: 12px; min-height: 30px; padding: 4px 10px; }
  .btn-peligro { color: var(--color-peligro); }
  .btn-peligro:hover { background: color-mix(in srgb, var(--color-peligro) 10%, transparent); }
  .btn-peligro-fill {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    min-height: 38px;
    padding: 8px 16px;
    border: 1px solid var(--color-peligro);
    border-radius: var(--radius-md);
    background: var(--color-peligro);
    font-family: var(--font-body);
    font-size: 14px;
    color: #fff;
    cursor: pointer;
    transition: background 0.12s ease;
  }
  .btn-peligro-fill:hover { background: color-mix(in srgb, var(--color-peligro) 85%, #000); }
  .btn-peligro-fill:disabled { opacity: 0.45; pointer-events: none; }
  .campos-modal {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    margin-bottom: var(--space-4);
  }
</style>
