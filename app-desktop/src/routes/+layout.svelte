<script>
  import '$lib/estilos/diseno.css';
  import { onMount, onDestroy } from 'svelte';
  import { goto } from '$app/navigation';
  import BarraTitulo from '$lib/componentes/BarraTitulo.svelte';
  import Toast from '$lib/componentes/Toast.svelte';
  import ModalConfirmacion from '$lib/componentes/ModalConfirmacion.svelte';
  import ModalActualizacionForzada from '$lib/componentes/ModalActualizacionForzada.svelte';
  import {
    estadoConfirmacion,
    permitiendoCierre,
    solicitarCerrarApp,
    cancelarCerrarApp,
    confirmarCerrarApp,
  } from '$lib/confirmacion.svelte.js';
  import { BASE_API } from '$lib/config.js';
  import {
    verificarActualizaciones,
    iniciarEscuchaEventos,
    detenerEscuchaEventos,
  } from '$lib/updater.svelte.js';

  let { children } = $props();

  let restaurando = $state(true);
  /** @type {(() => void) | null} */
  let desescucharCierre = null;

  onMount(async () => {
    // Interceptar solicitudes de cierre nativas en Tauri (Alt+F4, botón de cerrar de ventana del SO, etc.)
    if (typeof window !== 'undefined' && '__TAURI_INTERNALS__' in window) {
      try {
        const { getCurrentWindow } = await import('@tauri-apps/api/window');
        const appWindow = getCurrentWindow();
        desescucharCierre = await appWindow.onCloseRequested((event) => {
          if (permitiendoCierre.activo) {
            // El usuario ya aceptó cerrar la app: permitir cierre nativo sin volver a interceptar
            return;
          }
          event.preventDefault();
          solicitarCerrarApp();
        });
      } catch (e) {
        console.error('No se pudo registrar handler de cierre en Tauri:', e);
      }
    }

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

      // Hay negocio: al arrancar la app siempre mostramos la pantalla de inicio de sesión (/login)
      if (window.location.pathname !== '/login') {
        await goto('/login');
      }
    } catch (e) {
      console.warn('Error conectando a la API al iniciar:', e);
      if (window.location.pathname !== '/login') {
        await goto('/login');
      }
    } finally {
      restaurando = false;
      // Verificar actualizaciones del sistema en segundo plano al arrancar
      verificarActualizaciones(true);
      // Iniciar canal push pasivo en tiempo real (SSE)
      iniciarEscuchaEventos();
    }
  });

  onDestroy(() => {
    detenerEscuchaEventos();
    if (desescucharCierre) {
      desescucharCierre();
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

  <ModalActualizacionForzada />

  <ModalConfirmacion
    abierto={estadoConfirmacion.cerrarApp}
    titulo="¿Salir de StockAPP?"
    mensaje="¿Estás seguro de que querés cerrar la aplicación?"
    textoConfirmar="Sí, salir"
    textoCancelar="Cancelar"
    variante="peligro"
    onconfirmar={confirmarCerrarApp}
    oncancelar={cancelarCerrarApp}
  />
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
