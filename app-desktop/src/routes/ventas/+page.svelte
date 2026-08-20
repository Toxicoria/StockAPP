<script>
  // @ts-nocheck
  import { onMount } from 'svelte';
  import BotonVolver from '$lib/componentes/BotonVolver.svelte';
  import ModalTicket from '$lib/componentes/ModalTicket.svelte';
  import { pedirApi } from '$lib/api.js';
  import { fmtPesos } from '$lib/formato.js';

  let datos = $state({ total_dia: 0, cantidad: 0, promedio: 0, ventas: [] });
  let error = $state('');

  // Estado del modal de ticket
  let ticketSeleccionado = $state(null);
  let cargandoTicket = $state(false);
  let errorTicket = $state('');
  let modalAbierto = $state(false);

  onMount(cargar);

  async function cargar() {
    try {
      datos = await pedirApi('/api/ventas');
    } catch (e) {
      error = e.message;
    }
  }

  async function abrirTicket(v) {
    modalAbierto = true;
    cargandoTicket = true;
    errorTicket = '';
    ticketSeleccionado = null;
    try {
      const detalle = await pedirApi(`/api/ventas/${v.id_venta}`);
      ticketSeleccionado = detalle;
    } catch (e) {
      errorTicket = e.message || 'No se pudo cargar el ticket';
    } finally {
      cargandoTicket = false;
    }
  }

  async function facturar(v, ev) {
    if (ev) ev.stopPropagation();
    try {
      const r = await pedirApi(`/api/ventas/${v.id_venta}/factura`, { method: 'POST' });
      v.factura = r.factura;
      if (ticketSeleccionado && ticketSeleccionado.id_venta === v.id_venta) {
        ticketSeleccionado.factura = r.factura;
      }
    } catch (e) {
      error = e.message;
    }
  }
</script>

<BotonVolver />

<div class="cabecera">
  <div>
    <h2>Ventas</h2>
    <p class="text-muted">La caja de hoy, venta por venta. Clic en cualquier fila para ver el ticket e imprimir.</p>
  </div>
</div>

<div class="stats">
  <div class="card">
    <span class="card-kicker">Total del día</span>
    <strong>{fmtPesos(datos.total_dia)}</strong>
  </div>
  <div class="card">
    <span class="card-kicker">Ventas</span>
    <strong>{datos.cantidad}</strong>
  </div>
  <div class="card">
    <span class="card-kicker">Ticket promedio</span>
    <strong>{fmtPesos(datos.promedio)}</strong>
  </div>
</div>

{#if error}
  <p class="error">{error}</p>
{/if}

<div class="card tabla">
  <table class="table">
    <thead>
      <tr>
        <th>Hora</th>
        <th>Detalle (Resumen)</th>
        <th>Pago</th>
        <th>Total</th>
        <th>Factura</th>
        <th class="text-derecha">Acción</th>
      </tr>
    </thead>
    <tbody>
      {#if datos.ventas.length === 0}
        <tr>
          <td colspan="6" class="text-center text-muted vacio">No hay ventas registradas en el día de hoy.</td>
        </tr>
      {/if}
      {#each datos.ventas as v (v.id_venta)}
        <tr class="fila-venta" onclick={() => abrirTicket(v)}>
          <td class="col-hora"><strong>{v.hora}</strong></td>
          <td class="col-detalle">{v.detalle}</td>
          <td><span class="tag tag-neutral">{v.metodo_pago}</span></td>
          <td class="col-total">{fmtPesos(v.total_venta)}</td>
          <td>
            {#if v.factura}
              <span class="tag tag-accent">{v.factura}</span>
            {:else}
              <button class="btn btn-secondary chico" onclick={(ev) => facturar(v, ev)}>
                Facturar
              </button>
            {/if}
          </td>
          <td class="text-derecha">
            <button class="btn btn-ghost chico btn-ver-ticket" onclick={(ev) => { ev.stopPropagation(); abrirTicket(v); }}>
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line><polyline points="10 9 9 9 8 9"></polyline></svg>
              <span>Ver Ticket</span>
            </button>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>

<ModalTicket
  abierto={modalAbierto}
  ticket={ticketSeleccionado}
  cargando={cargandoTicket}
  error={errorTicket}
  onCerrar={() => (modalAbierto = false)}
  onFacturar={() => {
    if (ticketSeleccionado) {
      const v = datos.ventas.find((item) => item.id_venta === ticketSeleccionado.id_venta);
      if (v) facturar(v);
    }
  }}
/>

<style>
  .cabecera {
    margin-bottom: 18px;
  }
  .cabecera p {
    margin: 0;
  }
  .stats {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 14px;
    margin-bottom: 18px;
  }
  .stats strong {
    font-size: 22px;
  }
  .tabla {
    padding: 4px;
    overflow: auto;
  }
  .table {
    width: 100%;
    border-collapse: collapse;
  }
  .fila-venta {
    cursor: pointer;
    transition: background 0.15s ease;
  }
  .fila-venta:hover {
    background: color-mix(in srgb, var(--color-accent) 8%, transparent);
  }
  .col-hora {
    font-feature-settings: 'tnum';
    white-space: nowrap;
  }
  .col-detalle {
    color: var(--color-text);
  }
  .col-total {
    font-weight: 700;
    font-feature-settings: 'tnum';
  }
  .text-derecha {
    text-align: right;
  }
  .text-center {
    text-align: center;
  }
  .vacio {
    padding: 24px 0;
  }
  .chico {
    font-size: 12px;
    padding: 4px 10px;
    min-height: 28px;
  }
  .btn-ver-ticket {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .error {
    color: var(--color-peligro);
  }
</style>
