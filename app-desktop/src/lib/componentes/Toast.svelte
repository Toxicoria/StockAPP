<script>
  import { toasts, removerToast } from '$lib/toast.svelte.js';
</script>

{#if toasts.length > 0}
  <div class="toast-container">
    {#each toasts as t (t.id)}
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <div
        class="toast toast-{t.tipo}"
        onclick={() => removerToast(t.id)}
        role="alert"
      >
        <span class="toast-icono">
          {#if t.tipo === 'exito'}
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#137333" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>
          {:else if t.tipo === 'error'}
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="var(--color-peligro)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="15" y1="9" x2="9" y2="15"></line><line x1="9" y1="9" x2="15" y2="15"></line></svg>
          {:else}
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="var(--color-accent)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>
          {/if}
        </span>
        <span class="toast-mensaje">{t.mensaje}</span>
        <button class="toast-cerrar" type="button">×</button>
      </div>
    {/each}
  </div>
{/if}

<style>
  .toast-container {
    position: fixed;
    bottom: 24px;
    right: 24px;
    z-index: 10000;
    display: flex;
    flex-direction: column;
    gap: 10px;
    max-width: 400px;
    pointer-events: auto;
  }
  .toast {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
    border-radius: var(--radius-md);
    background: #ffffff;
    color: var(--color-text);
    font-size: 13.5px;
    font-weight: 500;
    box-shadow: 0 8px 24px color-mix(in srgb, #000 16%, transparent);
    border: 1px solid var(--color-divider);
    cursor: pointer;
    animation: toastIn 0.25s cubic-bezier(0.16, 1, 0.3, 1);
    transition: opacity 0.2s ease, transform 0.2s ease;
  }
  @keyframes toastIn {
    from {
      opacity: 0;
      transform: translateY(14px) scale(0.95);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
  .toast-exito {
    border-left: 4px solid #137333;
    background: #f6fbf7;
  }
  .toast-error {
    border-left: 4px solid var(--color-peligro);
    background: #fdf5f5;
  }
  .toast-info {
    border-left: 4px solid var(--color-accent);
    background: var(--color-accent-100);
  }
  .toast-icono {
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .toast-mensaje {
    flex: 1;
    line-height: 1.4;
  }
  .toast-cerrar {
    background: none;
    border: none;
    font-size: 18px;
    color: color-mix(in srgb, var(--color-text) 45%, transparent);
    cursor: pointer;
    padding: 0 4px;
  }
</style>
