<script>
  import { goto } from '$app/navigation';
  import { sesion, esDueno, haySesion, tienePermiso, cerrarSesion } from '$lib/sesion.svelte.js';
  import { pedirApi } from '$lib/api.js';
  import ModalConfirmacion from '$lib/componentes/ModalConfirmacion.svelte';
  import {
    estadoConfirmacion,
    solicitarCerrarSesion,
    cancelarCerrarSesion,
  } from '$lib/confirmacion.svelte.js';

  let alertas = $state(0);

  $effect(() => {
    if (!haySesion()) return;
    if (tienePermiso('stock')) {
      pedirApi('/api/stock')
        .then((/** @type {any[]} */ items) => {
          alertas = items.filter((i) => i.stock_minimo > 0 && i.cantidad_disponible <= i.stock_minimo).length;
        })
        .catch(() => {});
    }
  });

  async function ejecutarCerrarSesion() {
    cancelarCerrarSesion();
    await cerrarSesion();
    goto('/login');
  }

  const rolEtiqueta = $derived(esDueno() ? 'Dueño/a · Ve todo' : 'Empleado · Acceso personalizado');
</script>

<div class="hub">
  <div class="portada">
    <h1>{sesion.negocio || 'Mi negocio'}</h1>
    <p class="text-muted">Hola, {sesion.nombre} · ¿Qué querés hacer?</p>
  </div>

  <div class="grilla">
    {#if tienePermiso('vender')}
      <button class="accion destacada" onclick={() => goto('/vender')}>
        <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="var(--color-accent-600)" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><circle cx="8" cy="21" r="1"></circle><circle cx="19" cy="21" r="1"></circle><path d="M2.05 2.05h2l2.66 12.42a2 2 0 0 0 2 1.58h9.78a2 2 0 0 0 1.95-1.57l1.65-7.43H5.12"></path></svg>
        <span class="nombre">Vender</span>
        <span class="text-muted detalle">Cobrar en el mostrador. Descuenta el stock solo.</span>
      </button>
    {/if}

    {#if tienePermiso('stock')}
      <button class="accion" onclick={() => goto('/stock')}>
        <span class="fila">
          <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="var(--color-accent-600)" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><rect width="20" height="5" x="2" y="3" rx="1"></rect><path d="M4 8v11a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8"></path><path d="M10 12h4"></path></svg>
          {#if alertas > 0}<span class="tag tag-accent">{alertas} para reponer</span>{/if}
        </span>
        <span class="nombre">Stock</span>
        <span class="text-muted detalle">Ver qué hay y cargar mercadería que llega.</span>
      </button>
    {/if}

    {#if tienePermiso('ventas')}
      <button class="accion" onclick={() => goto('/ventas')}>
        <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="var(--color-accent-600)" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M4 2v20l2-1 2 1 2-1 2 1 2-1 2 1 2-1 2 1V2l-2 1-2-1-2 1-2-1-2 1-2-1-2 1Z"></path><path d="M14 8H8"></path><path d="M16 12H8"></path><path d="M13 16H8"></path></svg>
        <span class="nombre">Ventas</span>
        <span class="text-muted detalle">La caja del día, venta por venta.</span>
      </button>
    {/if}

    {#if tienePermiso('resumen')}
      <button class="accion" onclick={() => goto('/resumen')}>
        <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="var(--color-accent-600)" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><rect width="7" height="9" x="3" y="3" rx="1"></rect><rect width="7" height="5" x="14" y="3" rx="1"></rect><rect width="7" height="9" x="14" y="12" rx="1"></rect><rect width="7" height="5" x="3" y="16" rx="1"></rect></svg>
        <span class="nombre">Resumen</span>
        <span class="text-muted detalle">Cómo viene el día y la semana.</span>
      </button>
    {/if}

    {#if tienePermiso('productos')}
      <button class="accion" onclick={() => goto('/productos')}>
        <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="var(--color-accent-600)" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z"></path><path d="m3.3 7 8.7 5 8.7-5"></path><path d="M12 22V12"></path></svg>
        <span class="nombre">Productos</span>
        <span class="text-muted detalle">Agregar productos y cambiar precios.</span>
      </button>
    {/if}

    {#if tienePermiso('precios')}
      <button class="accion" onclick={() => goto('/precios')}>
        <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="var(--color-accent-600)" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><line x1="19" x2="5" y1="5" y2="19"></line><circle cx="6.5" cy="6.5" r="2.5"></circle><circle cx="17.5" cy="17.5" r="2.5"></circle></svg>
        <span class="nombre">Precios</span>
        <span class="text-muted detalle">Aplicar el aumento mensual de cada proveedor.</span>
      </button>
    {/if}

    {#if tienePermiso('facturas')}
      <button class="accion" onclick={() => goto('/facturas')}>
        <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="var(--color-accent-600)" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z"></path><path d="M14 2v4a2 2 0 0 0 2 2h4"></path><path d="M10 9H8"></path><path d="M16 13H8"></path><path d="M16 17H8"></path></svg>
        <span class="nombre">Facturación</span>
        <span class="text-muted detalle">Las facturas emitidas, todas juntas.</span>
      </button>
    {/if}

    {#if tienePermiso('usuarios')}
      <button class="accion" onclick={() => goto('/usuarios')}>
        <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="var(--color-accent-600)" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M22 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path></svg>
        <span class="nombre">Usuarios</span>
        <span class="text-muted detalle">Crear y gestionar empleados del negocio.</span>
      </button>
    {/if}

    {#if tienePermiso('configuracion')}
      <button class="accion" onclick={() => goto('/configuracion')}>
        <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="var(--color-accent-600)" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"></path><circle cx="12" cy="12" r="3"></circle></svg>
        <span class="nombre">Configuración</span>
        <span class="text-muted detalle">Datos del negocio y clave fiscal de ARCA.</span>
      </button>
    {/if}
  </div>

  <div class="pie">
    <span class="text-muted">{sesion.nombre} · {rolEtiqueta}</span>
    <button class="btn btn-ghost" onclick={solicitarCerrarSesion}>Cerrar sesión</button>
  </div>
</div>

<ModalConfirmacion
  abierto={estadoConfirmacion.cerrarSesion}
  titulo="¿Cerrar sesión?"
  mensaje="¿Estás seguro de que querés salir de tu cuenta?"
  textoConfirmar="Sí, cerrar sesión"
  textoCancelar="Cancelar"
  variante="peligro"
  onconfirmar={ejecutarCerrarSesion}
  oncancelar={cancelarCerrarSesion}
/>

<style>
  .hub { max-width: 980px; margin: 0 auto; }
  .portada {
    text-align: center;
    padding: 26px 0 30px;
  }
  .portada h1 { margin-bottom: 6px; }
  .portada p { font-size: 13.5px; margin: 0; }
  .grilla {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(230px, 1fr));
    gap: 16px;
  }
  .accion {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
    min-height: 132px;
    padding: 20px;
    border: 1px solid var(--color-divider);
    border-radius: var(--radius-lg);
    background: transparent;
    font-family: var(--font-body);
    cursor: pointer;
    text-align: left;
    color: var(--color-text);
  }
  .accion:hover {
    background: color-mix(in srgb, var(--color-accent) 8%, transparent);
    border-color: var(--color-accent);
  }
  .destacada { border-color: var(--color-accent-400); }
  .fila { display: flex; align-items: center; gap: 10px; }
  .nombre { font-size: 17px; font-weight: 600; }
  .detalle { font-size: 12.5px; line-height: 1.4; }
  .pie {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 18px;
    margin-top: 34px;
    flex-wrap: wrap;
    font-size: 12px;
  }
</style>
