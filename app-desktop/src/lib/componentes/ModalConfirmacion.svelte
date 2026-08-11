<script>
  import { fade, fly } from 'svelte/transition';

  /**
   * @type {{
   *   abierto?: boolean,
   *   titulo?: string,
   *   mensaje?: string,
   *   textoConfirmar?: string,
   *   textoCancelar?: string,
   *   variante?: 'peligro' | 'primario',
   *   onconfirmar?: () => void,
   *   oncancelar?: () => void
   * }}
   */
  let {
    abierto = false,
    titulo = '¿Confirmar acción?',
    mensaje = '',
    textoConfirmar = 'Confirmar',
    textoCancelar = 'Cancelar',
    variante = 'peligro', // 'peligro' | 'primario'
    onconfirmar = () => {},
    oncancelar = () => {},
  } = $props();

  /** @param {MouseEvent} ev */
  function alClickFondo(ev) {
    if (ev.target === ev.currentTarget) {
      oncancelar();
    }
  }

  /** @param {KeyboardEvent} ev */
  function alTeclaAbajo(ev) {
    if (ev.key === 'Escape') {
      oncancelar();
    }
  }
</script>

{#if abierto}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="dialog-backdrop"
    transition:fade={{ duration: 150 }}
    onclick={alClickFondo}
    onkeydown={alTeclaAbajo}
  >
    <div
      class="dialog"
      transition:fly={{ y: 12, duration: 200 }}
      role="dialog"
      aria-modal="true"
    >
      <div class="dialog-cabecera">
        <div class="icono-modal {variante}">
          {#if variante === 'peligro'}
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>
          {:else}
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"></path><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>
          {/if}
        </div>
        <h3 class="dialog-title">{titulo}</h3>
      </div>

      {#if mensaje}
        <p class="dialog-body text-muted">{mensaje}</p>
      {/if}

      <div class="dialog-actions">
        <button type="button" class="btn btn-secondary" onclick={oncancelar}>
          {textoCancelar}
        </button>
        <button
          type="button"
          class="btn {variante === 'peligro' ? 'btn-peligro' : 'btn-primary'}"
          onclick={onconfirmar}
        >
          {textoConfirmar}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .dialog-cabecera {
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }
  .icono-modal {
    width: 38px;
    height: 38px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }
  .icono-modal.peligro {
    background: rgba(168, 68, 60, 0.12);
    color: var(--color-peligro, #a8443c);
  }
  .icono-modal.primario {
    background: rgba(46, 110, 115, 0.12);
    color: var(--color-accent-600, #2e6e73);
  }
  .btn-peligro {
    background: var(--color-peligro, #a8443c);
    border-color: var(--color-peligro, #a8443c);
    color: #fff;
  }
  .btn-peligro:hover {
    background: #8e3831;
    border-color: #8e3831;
  }
</style>
