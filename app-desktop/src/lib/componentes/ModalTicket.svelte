<script>
  import Dialogo from './Dialogo.svelte';
  import { fmtPesos } from '$lib/formato.js';

  /**
   * @typedef {{
   *   id_venta: number,
   *   fecha: string,
   *   hora: string,
   *   metodo_pago: string,
   *   total_venta: number,
   *   factura?: string | null,
   *   cajero: string,
   *   negocio: { nombre: string, direccion?: string, cuit?: string, punto_venta: string, telefono?: string },
   *   items: Array<{ id_producto: string, descripcion: string, marca: string, cantidad: number, precio_unitario: number, subtotal: number }>
   * }} Ticket
   */

  let {
    abierto = false,
    ticket = /** @type {Ticket | null} */ (null),
    cargando = false,
    error = '',
    onCerrar = () => {},
    onFacturar = () => {}
  } = $props();

  function imprimir() {
    window.print();
  }
</script>

<Dialogo {abierto} ancho="480px">
  {#if cargando}
    <div class="ticket-estado">
      <span class="text-muted">Cargando ticket de venta…</span>
    </div>
  {:else if error}
    <div class="ticket-estado">
      <p class="error">{error}</p>
      <button class="btn btn-secondary" onclick={() => onCerrar()}>Cerrar</button>
    </div>
  {:else if ticket}
    <div class="contenedor-modal-ticket">
      <!-- BOTONES DE ACCIÓN SUPERIORES -->
      <div class="acciones-header sin-impresion">
        <div class="titulo-modal">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M6 2 3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6l-3-4Z"></path><path d="M3 6h18"></path><path d="M16 10a4 4 0 0 1-8 0"></path></svg>
          <span>Ticket #{ticket.id_venta}</span>
        </div>
        <button class="btn-cerrar" onclick={() => onCerrar()} title="Cerrar modal">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
        </button>
      </div>

      <!-- VISTA PREVIA DEL TICKET (ESTILO PAPEL TÉRMICO) -->
      <div class="ticket-papel" id="ticket-imprimible">
        <!-- ENCABEZADO NEGOCIO -->
        <div class="encabezado-negocio">
          <h3 class="nombre-negocio">{ticket.negocio.nombre}</h3>
          {#if ticket.negocio.direccion}
            <p class="meta-negocio">{ticket.negocio.direccion}</p>
          {/if}
          {#if ticket.negocio.cuit}
            <p class="meta-negocio">CUIT: {ticket.negocio.cuit}</p>
          {/if}
          {#if ticket.negocio.telefono}
            <p class="meta-negocio">Tel: {ticket.negocio.telefono}</p>
          {/if}
        </div>

        <div class="separador-puntos"></div>

        <!-- INFO TRANSACCIÓN -->
        <div class="datos-transaccion">
          <div class="fila-meta">
            <span>Ticket N°:</span>
            <strong>#{ticket.id_venta}</strong>
          </div>
          <div class="fila-meta">
            <span>Fecha y hora:</span>
            <span>{ticket.fecha} {ticket.hora}</span>
          </div>
          <div class="fila-meta">
            <span>Cajero:</span>
            <span>{ticket.cajero}</span>
          </div>
          <div class="fila-meta">
            <span>Pago:</span>
            <span class="metodo">{ticket.metodo_pago.toUpperCase()}</span>
          </div>
          {#if ticket.factura}
            <div class="fila-meta">
              <span>Comprobante:</span>
              <strong class="factura-badge">{ticket.factura}</strong>
            </div>
          {/if}
        </div>

        <div class="separador-puntos"></div>

        <!-- TABLA DE PRODUCTOS -->
        <div class="lista-items">
          <div class="cabecera-items">
            <span class="col-cant">Cant</span>
            <span class="col-desc">Descripción</span>
            <span class="col-sub">Total</span>
          </div>

          <div class="cuerpo-items">
            {#each ticket.items as it (it.id_producto)}
              <div class="item-renglon">
                <div class="item-linea-principal">
                  <span class="col-cant">{it.cantidad}x</span>
                  <span class="col-desc">{it.descripcion}</span>
                  <span class="col-sub">{fmtPesos(it.subtotal)}</span>
                </div>
                {#if it.cantidad > 1 || it.marca}
                  <div class="item-linea-secundaria">
                    <span class="text-muted">
                      ({fmtPesos(it.precio_unitario)} c/u) {it.marca ? `• ${it.marca}` : ''}
                    </span>
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        </div>

        <div class="separador-puntos"></div>

        <!-- TOTALES -->
        <div class="resumen-totales">
          <div class="fila-total">
            <span>TOTAL COBRADO</span>
            <strong class="monto-total">{fmtPesos(ticket.total_venta)}</strong>
          </div>
        </div>

        <div class="separador-puntos"></div>

        <div class="pie-ticket">
          <p>¡Gracias por su compra!</p>
          <span class="sistema-tag">StockAPP — Sistema de Gestión</span>
        </div>
      </div>

      <!-- BOTONES DE ACCIÓN INFERIORES -->
      <div class="acciones-footer sin-impresion">
        {#if !ticket.factura}
          <button class="btn btn-secondary" onclick={() => onFacturar()}>
            <span>Facturar (AFIP)</span>
          </button>
        {/if}
        <button class="btn btn-primary" onclick={imprimir}>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 6 2 18 2 18 9"></polyline><path d="M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2"></path><rect x="6" y="14" width="12" height="8"></rect></svg>
          <span>Imprimir Ticket</span>
        </button>
      </div>
    </div>
  {/if}
</Dialogo>

<style>
  .contenedor-modal-ticket {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .ticket-estado {
    padding: 32px;
    text-align: center;
  }

  .acciones-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-bottom: 12px;
    border-bottom: 1px solid var(--color-divider);
  }

  .titulo-modal {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 16px;
    font-weight: 700;
    color: var(--color-text);
  }

  .btn-cerrar {
    background: transparent;
    border: none;
    cursor: pointer;
    padding: 4px;
    border-radius: 50%;
    color: var(--color-text-muted, #666);
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .btn-cerrar:hover {
    background: color-mix(in srgb, var(--color-text) 10%, transparent);
  }

  /* ESTILO PAPEL TÉRMICO RECIBO */
  .ticket-papel {
    background: #ffffff;
    border: 1px solid var(--color-divider);
    border-radius: var(--radius-md);
    padding: 24px;
    font-family: var(--font-body);
    font-size: 13px;
    color: #1a1a1a;
    box-shadow: inset 0 0 8px rgba(0, 0, 0, 0.02);
  }

  .encabezado-negocio {
    text-align: center;
  }
  .nombre-negocio {
    font-size: 18px;
    font-weight: 700;
    margin: 0 0 4px 0;
    text-transform: uppercase;
    letter-spacing: 0.02em;
  }
  .meta-negocio {
    margin: 2px 0;
    font-size: 12px;
    color: #555;
  }

  .separador-puntos {
    border-bottom: 1px dashed #aaa;
    margin: 14px 0;
  }

  .datos-transaccion {
    display: flex;
    flex-direction: column;
    gap: 6px;
    font-size: 12.5px;
  }
  .fila-meta {
    display: flex;
    justify-content: space-between;
  }
  .metodo {
    font-weight: 600;
  }
  .factura-badge {
    color: var(--color-accent-700);
  }

  .lista-items {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .cabecera-items {
    display: flex;
    font-weight: 700;
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #666;
    padding-bottom: 4px;
  }
  .col-cant {
    width: 44px;
  }
  .col-desc {
    flex: 1;
  }
  .col-sub {
    width: 90px;
    text-align: right;
  }

  .cuerpo-items {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .item-renglon {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .item-linea-principal {
    display: flex;
    align-items: baseline;
    font-weight: 600;
  }
  .item-linea-secundaria {
    padding-left: 44px;
    font-size: 11px;
  }

  .resumen-totales {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .fila-total {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 15px;
    font-weight: 700;
  }
  .monto-total {
    font-size: 20px;
    color: var(--color-accent-800, #153c40);
    font-feature-settings: 'tnum';
  }

  .pie-ticket {
    text-align: center;
    font-size: 12px;
    color: #666;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .pie-ticket p {
    margin: 0;
    font-weight: 600;
  }
  .sistema-tag {
    font-size: 10px;
    color: #999;
  }

  .acciones-footer {
    display: flex;
    gap: 10px;
    justify-content: flex-end;
    padding-top: 8px;
  }

  .error {
    color: var(--color-peligro);
  }

  /* ESTILOS DE IMPRESIÓN IMPERDIBLES */
  @media print {
    :global(body *) {
      visibility: hidden !important;
    }
    #ticket-imprimible,
    #ticket-imprimible * {
      visibility: visible !important;
    }
    #ticket-imprimible {
      position: fixed !important;
      left: 0 !important;
      top: 0 !important;
      width: 80mm !important;
      padding: 8mm !important;
      margin: 0 !important;
      border: none !important;
      box-shadow: none !important;
      background: #ffffff !important;
      color: #000000 !important;
    }
    .sin-impresion {
      display: none !important;
    }
  }
</style>
