<script>
  import '$lib/estilos/diseno.css';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import BarraTitulo from '$lib/componentes/BarraTitulo.svelte';
  import Toast from '$lib/componentes/Toast.svelte';
  import { sesion, haySesion, refrescar } from '$lib/sesion.svelte.js';
  import { pedirApi } from '$lib/api.js';
  import { BASE_API } from '$lib/config.js';

  let { children } = $props();

  // Al arrancar, la app primero chequea si hay un negocio registrado.
  // Si no lo hay, redirige al wizard de configuración inicial.
  // Si lo hay, intenta revivir la sesión con el refresh token guardado;
  // hasta resolver eso no se muestra nada (evita el parpadeo de una
  // pantalla a la que después te echa).
  let restaurando = $state(true);

  onMount(async () => {
    try {
      // Si ya estamos en /setup, no hacer nada más.
      if (window.location.pathname.startsWith('/setup')) {
        restaurando = false;
        return;
      }

      // ¿Hay un negocio registrado?
      const resp = await fetch(`${BASE_API}/api/registro`);
      if (resp.ok) {
        const datos = await resp.json();
        if (!datos.registrado) {
          await goto('/setup');
          restaurando = false;
          return;
        }
      }

      // Hay negocio: intentar revivir la sesión.
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
  <Toast />
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
