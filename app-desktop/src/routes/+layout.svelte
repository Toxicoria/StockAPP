<script>
  import '$lib/estilos/diseno.css';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import BarraTitulo from '$lib/componentes/BarraTitulo.svelte';
  import { sesion, haySesion, refrescar } from '$lib/sesion.svelte.js';
  import { pedirApi } from '$lib/api.js';

  let { children } = $props();

  // Al arrancar, la app intenta revivir la sesión con el refresh token
  // guardado; hasta resolver eso no se muestra nada (evita el parpadeo
  // de una pantalla a la que después te echa).
  let restaurando = $state(true);

  onMount(async () => {
    try {
      const viva = await refrescar();
      if (viva) {
        pedirApi('/api/negocio')
          .then((n) => (sesion.negocio = n.nombre_negocio))
          .catch(() => {});
      } else if (window.location.pathname !== '/login') {
        await goto('/login');
      }
    } finally {
      restaurando = false;
    }
  });
</script>

<div class="app">
  <BarraTitulo />
  <main>
    {#if !restaurando}
      {@render children()}
    {/if}
  </main>
</div>

<style>
  .app {
    display: flex;
    flex-direction: column;
    height: 100vh;
    overflow: hidden;
  }
  main {
    flex: 1;
    min-height: 0;
    overflow: auto;
    padding: 30px 42px 34px;
  }
</style>
