<script>
  import { onMount } from 'svelte';
  import BotonVolver from '$lib/componentes/BotonVolver.svelte';
  import Dialogo from '$lib/componentes/Dialogo.svelte';
  import { pedirApi } from '$lib/api.js';
  import { fmtPesos } from '$lib/formato.js';

  /** @typedef {{ id_producto: string, descripcion: string, cantidad_disponible: number, precio_venta: number }} Item */

  /** @type {Item[]} */
  let inventario = $state([]);
  let query = $state('');
  /** @type {{ item: Item, cantidad: number }[]} */
  let carrito = $state([]);
  let pago = $state('efectivo');
  let pagaCon = $state('');
  let cobrando = $state(false);
  let error = $state('');
  let ultimoTotal = $state(0);
  let ultimoPagaCon = $state(0);
  let ultimoVuelto = $state(0);
  let verConfirmacion = $state(false);
  /** @type {HTMLInputElement | undefined} */
  let inputBusqueda = $state();

  let tiempoUltimaDigitacion = 0;

  onMount(async () => {
    try {
      inventario = await pedirApi('/api/stock');
    } catch {
      error = 'no se pudo cargar el inventario';
    }
    inputBusqueda?.focus();
  });

  const resultados = $derived.by(() => {
    const q = query.trim().toLowerCase();
    if (!q) return [];
    return inventario
      .filter((p) => p.descripcion.toLowerCase().includes(q) || p.id_producto.startsWith(q))
      .slice(0, 6);
  });

  const total = $derived(carrito.reduce((t, l) => t + l.item.precio_venta * l.cantidad, 0));
  const montoPagaCon = $derived(parseFloat(String(pagaCon).replace(/[^0-9.]/g, '')) || 0);
  const vuelto = $derived(pago === 'efectivo' && montoPagaCon >= total ? montoPagaCon - total : 0);
  const falta = $derived(pago === 'efectivo' && montoPagaCon > 0 && montoPagaCon < total ? total - montoPagaCon : 0);
  const efectivoInsuficiente = $derived(pago === 'efectivo' && montoPagaCon > 0 && montoPagaCon < total);

  /** @param {Item} item */
  function agregar(item) {
    const linea = carrito.find((l) => l.item.id_producto === item.id_producto);
    if (linea) {
      linea.cantidad += 1;
      carrito = [...carrito.filter((l) => l.item.id_producto !== item.id_producto), linea];
    } else {
      carrito.push({ item, cantidad: 1 });
    }
    query = '';
    inputBusqueda?.focus();
  }

  /** @param {string} id @param {number} delta */
  function ajustar(id, delta) {
    const linea = carrito.find((l) => l.item.id_producto === id);
    if (!linea) return;
    linea.cantidad += delta;
    if (linea.cantidad <= 0) {
      carrito = carrito.filter((l) => l.item.id_producto !== id);
    }
  }

  /** @param {KeyboardEvent} ev */
  function alTeclearBuscador(ev) {
    if (ev.key !== 'Enter') return;
    ev.preventDefault();
    const codigo = query.trim();
    const exacto = inventario.find((p) => p.id_producto === codigo);
    if (exacto) agregar(exacto);
    else if (resultados.length === 1) agregar(resultados[0]);
  }

  /** @param {KeyboardEvent} ev */
  function alManejarTecladoGlobal(ev) {
    if (verConfirmacion) return;

    const target = /** @type {HTMLElement} */ (ev.target);
    const esInput = target && (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.tagName === 'SELECT');

    // 1. Tecla 'Delete' (Suprimir): Elimina el último producto y todo el bulto/cantidad completa
    if (ev.key === 'Delete') {
      ev.preventDefault();
      if (carrito.length > 0) {
        carrito.pop();
      }
      inputBusqueda?.focus();
      return;
    }

    // 2. Tecla 'Backspace' (Retroceso): Resta 1 sola unidad al último producto (elimina si llega a 0)
    if (ev.key === 'Backspace') {
      if (!esInput || (target === inputBusqueda && query === '')) {
        ev.preventDefault();
        if (carrito.length > 0) {
          const ultimo = carrito[carrito.length - 1];
          ajustar(ultimo.item.id_producto, -1);
        }
        inputBusqueda?.focus();
        return;
      }
    }

    // 3. Teclas '+' y '-': Ajustan cantidad del último producto
    if ((ev.key === '+' || ev.key === '-') && (!esInput || (target === inputBusqueda && query === ''))) {
      ev.preventDefault();
      if (carrito.length > 0) {
        const ultimo = carrito[carrito.length - 1];
        ajustar(ultimo.item.id_producto, ev.key === '+' ? 1 : -1);
      }
      return;
    }

    // 4. Teclas numéricas '0-9': Cambian la cantidad del último producto si el buscador está enfocado pero vacío
    if (/^[0-9]$/.test(ev.key) && (!esInput || (target === inputBusqueda && query === ''))) {
      if (carrito.length > 0) {
        ev.preventDefault();
        const ultimo = carrito[carrito.length - 1];
        const digito = parseInt(ev.key, 10);
        const ahora = Date.now();
        if (ahora - tiempoUltimaDigitacion < 1000 && ultimo.cantidad > 0) {
          const nuevaCant = parseInt(`${ultimo.cantidad}${digito}`, 10);
          if (nuevaCant <= 999) {
            ultimo.cantidad = nuevaCant;
          }
        } else {
          if (digito > 0) {
            ultimo.cantidad = digito;
          }
        }
        tiempoUltimaDigitacion = ahora;
        return;
      }
    }

    // 5. Captura de tipeo global: Dirigir escritura al buscador si se presiona una tecla común fuera de inputs
    if (!esInput && ev.key.length === 1 && !ev.ctrlKey && !ev.metaKey && !ev.altKey) {
      inputBusqueda?.focus();
    }
  }

  async function cobrar() {
    if (carrito.length === 0 || cobrando || efectivoInsuficiente) return;
    cobrando = true;
    error = '';
    try {
      const venta = await pedirApi('/api/ventas', {
        method: 'POST',
        body: {
          metodo_pago: pago,
          items: carrito.map((l) => ({ id_producto: l.item.id_producto, cantidad: l.cantidad })),
        },
      });
      ultimoTotal = venta.total_venta;
      ultimoPagaCon = montoPagaCon;
      ultimoVuelto = vuelto;
      carrito = [];
      pagaCon = '';
      verConfirmacion = true;
      pedirApi('/api/stock').then((datos) => (inventario = datos)).catch(() => {});
    } catch (e) {
      error = e instanceof Error ? e.message : 'no se pudo cobrar';
    } finally {
      cobrando = false;
    }
  }

  function vaciarCarrito() {
    carrito = [];
    pagaCon = '';
    inputBusqueda?.focus();
  }

  function cerrarConfirmacion() {
    verConfirmacion = false;
    inputBusqueda?.focus();
  }
</script>

<svelte:window onkeydown={alManejarTecladoGlobal} />

<BotonVolver />

<div class="pantalla">
  <div class="seccion-principal">
    <!-- FILA SUPERIOR: BARRA DE BÚSQUEDA Y BOTÓN DE ATAJOS -->
    <div class="fila-superior">
      <div class="buscador">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" class="lupa"><circle cx="11" cy="11" r="8"></circle><path d="m21 21-4.3-4.3"></path></svg>
        <input
          class="input entrada"
          bind:this={inputBusqueda}
          bind:value={query}
          onkeydown={alTeclearBuscador}
          placeholder="Escaneá el código o escribí el nombre del producto…"
        />
        {#if resultados.length > 0}
          <div class="desplegable">
            {#each resultados as p (p.id_producto)}
              <button class="resultado" onclick={() => agregar(p)}>
                <span class="res-nombre">{p.descripcion}</span>
                <span class="text-muted res-stock">{p.cantidad_disponible} en stock</span>
                <span class="res-precio">{fmtPesos(p.precio_venta)}</span>
              </button>
            {/each}
          </div>
        {/if}
      </div>

      <!-- BOTÓN Y MENÚ FLOTANTE DE ATAJOS -->
      <div class="contenedor-atajos">
        <button type="button" class="btn-atajos">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="20" height="16" x="2" y="4" rx="2"></rect><path d="M6 8h.01"></path><path d="M10 8h.01"></path><path d="M14 8h.01"></path><path d="M18 8h.01"></path><path d="M8 12h.01"></path><path d="M12 12h.01"></path><path d="M16 12h.01"></path><path d="M7 16h10"></path></svg>
          <span>Atajos del teclado</span>
        </button>
        <div class="menu-flotante-atajos">
          <span class="leyenda-titulo">Atajos del teclado</span>
          <ul class="atajos-lista">
            <li><kbd>0-9</kbd> <span>Cambiar cantidad</span></li>
            <li><kbd>+</kbd> <kbd>-</kbd> <span>Aumentar / reducir</span></li>
            <li><kbd>Backspace</kbd> <span>Restar 1 unidad</span></li>
            <li><kbd>Supr</kbd> <span>Eliminar producto</span></li>
          </ul>
        </div>
      </div>
    </div>

    <!-- ESTRUCTURA PRINCIPAL EN 2 COLUMNAS -->
    <div class="grilla-vender">
      <!-- COLUMNA IZQUIERDA: PRODUCTOS EN EL TICKET -->
      <div class="card elev-sm ticket-productos">
        <div class="ticket-cabecera">
          <div class="cabecera-izq">
            <h4>Ticket</h4>
            <span class="badge-items">{carrito.length} {carrito.length === 1 ? 'producto' : 'productos'}</span>
          </div>
          <button class="btn btn-ghost chico" onclick={vaciarCarrito} disabled={carrito.length === 0}>Vaciar</button>
        </div>
        <hr class="hr sin-margen" />

        {#if carrito.length === 0}
          <p class="text-muted vacio">
            Todavía no hay nada en el ticket.<br />Escaneá o buscá un producto para empezar.
          </p>
        {:else}
          <div class="lineas">
            {#each carrito as l, i (l.item.id_producto)}
              <div class="linea {i === carrito.length - 1 ? 'activa' : ''}">
                <div class="contador">
                  <button class="paso" onclick={() => ajustar(l.item.id_producto, -1)}>−</button>
                  <input
                    type="number"
                    min="1"
                    max="999"
                    class="cant-input"
                    bind:value={l.cantidad}
                    onchange={(e) => {
                      const val = parseInt(e.currentTarget.value, 10);
                      if (isNaN(val) || val <= 0) l.cantidad = 1;
                    }}
                  />
                  <button class="paso" onclick={() => ajustar(l.item.id_producto, 1)}>+</button>
                </div>
                <div class="linea-info">
                  <span class="linea-nombre">{l.item.descripcion}</span>
                  <span class="linea-unitario text-muted">{fmtPesos(l.item.precio_venta)} c/u</span>
                </div>
                <span class="linea-sub">{fmtPesos(l.item.precio_venta * l.cantidad)}</span>
                <button class="btn-eliminar-linea" title="Quitar producto" onclick={() => (carrito = carrito.filter((c) => c.item.id_producto !== l.item.id_producto))}>
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                </button>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <!-- COLUMNA DERECHA: TOTAL / MÉTODOS DE PAGO / COBRAR -->
      <div class="card elev-sm ticket-resumen">
        <div class="resumen-cabecera">
          <span class="text-muted total-label">Total a cobrar</span>
          <span class="total-monto">{fmtPesos(total)}</span>
        </div>

        <div class="campo-pago">
          <span class="pago-label">Método de pago</span>
          <div class="seg pago">
            <label class="seg-opt"><input type="radio" name="pago" value="efectivo" bind:group={pago} />Efectivo</label>
            <label class="seg-opt"><input type="radio" name="pago" value="transferencia" bind:group={pago} />Transferencia</label>
            <label class="seg-opt"><input type="radio" name="pago" value="tarjeta" bind:group={pago} />Tarjeta</label>
          </div>
        </div>

        <!-- CÁLCULO DE VUELTO EN EFECTIVO -->
        {#if pago === 'efectivo' && carrito.length > 0}
          <div class="caja-efectivo">
            <div class="field">
              <label for="pagaCon">Paga con (Efectivo)</label>
              <div class="campo-monto">
                <span class="simbolo">$</span>
                <input
                  id="pagaCon"
                  type="number"
                  min="0"
                  step="10"
                  class="input entrada-monto"
                  placeholder="Ej: {total}"
                  bind:value={pagaCon}
                />
              </div>
            </div>

            <!-- Botones de montos rápidos -->
            <div class="montos-rapidos">
              <button type="button" class="btn btn-secondary chico" onclick={() => (pagaCon = String(total))}>
                Exacto
              </button>
              {#each [1000, 2000, 5000, 10000, 20000, 50000] as m}
                {#if m >= total}
                  <button type="button" class="btn btn-secondary chico" onclick={() => (pagaCon = String(m))}>
                    {fmtPesos(m)}
                  </button>
                {/if}
              {/each}
            </div>

            <!-- Banner de Vuelto / Faltante -->
            {#if montoPagaCon > 0}
              {#if montoPagaCon >= total}
                <div class="banner-vuelto exito">
                  <span class="vuelto-label">Vuelto:</span>
                  <span class="vuelto-monto">{fmtPesos(vuelto)}</span>
                </div>
              {:else}
                <div class="banner-vuelto faltante">
                  <span class="vuelto-label">Faltan:</span>
                  <span class="vuelto-monto">{fmtPesos(falta)}</span>
                </div>
              {/if}
            {/if}
          </div>
        {/if}

        {#if error}
          <p class="error">{error}</p>
        {/if}

        <button
          class="btn btn-primary btn-block cobrar"
          onclick={cobrar}
          disabled={carrito.length === 0 || cobrando || efectivoInsuficiente}
        >
          {cobrando ? 'Cobrando…' : 'Cobrar'}
        </button>
      </div>
    </div>
  </div>
</div>

<Dialogo abierto={verConfirmacion}>
  <div class="confirmacion">
    <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="var(--color-accent)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><path d="m9 12 2 2 4-4"></path></svg>
    <div class="dialog-title">Venta cobrada</div>
    <div class="conf-total">{fmtPesos(ultimoTotal)}</div>

    {#if ultimoPagaCon >= ultimoTotal && ultimoPagaCon > 0}
      <div class="conf-vuelto-card">
        <span class="vuelto-titulo-conf">Vuelto a entregar</span>
        <span class="vuelto-monto-conf">{fmtPesos(ultimoVuelto)}</span>
      </div>
    {/if}

    <p class="dialog-body">Se descontó el stock y quedó anotada en Ventas.</p>
    <div class="dialog-actions conf-acciones">
      <button class="btn btn-primary conf-boton" onclick={cerrarConfirmacion}>Vender otra</button>
    </div>
  </div>
</Dialogo>

<style>
  .pantalla {
    max-width: 1180px;
    margin: 0 auto;
  }
  .seccion-principal {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }
  .fila-superior {
    display: flex;
    align-items: center;
    gap: 16px;
  }
  .buscador {
    flex: 1;
    position: relative;
  }

  .contenedor-atajos {
    position: relative;
    display: inline-block;
  }
  .btn-atajos {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    height: 52px;
    padding: 0 18px;
    background: #ffffff;
    border: 1px solid var(--color-divider);
    border-radius: var(--radius-lg, 12px);
    font-family: var(--font-body);
    font-size: 13.5px;
    font-weight: 600;
    color: color-mix(in srgb, var(--color-text) 75%, transparent);
    cursor: pointer;
    box-shadow: var(--shadow-sm);
    transition: all 0.15s ease;
  }
  .contenedor-atajos:hover .btn-atajos {
    border-color: var(--color-accent-400, #6cabb1);
    color: var(--color-accent-700, #1f5459);
    background: var(--color-accent-50, #f0f7f7);
  }
  .menu-flotante-atajos {
    position: absolute;
    top: calc(100% + 6px);
    right: 0;
    z-index: 40;
    width: 250px;
    background: #ffffff;
    border: 1px solid var(--color-divider);
    border-radius: var(--radius-lg, 12px);
    padding: 14px 16px;
    box-shadow: var(--shadow-md);
    opacity: 0;
    visibility: hidden;
    transform: translateY(-4px);
    transition: opacity 0.15s ease, transform 0.15s ease, visibility 0.15s;
    pointer-events: none;
  }
  .contenedor-atajos:hover .menu-flotante-atajos {
    opacity: 1;
    visibility: visible;
    transform: translateY(0);
  }
  .leyenda-titulo {
    display: block;
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.07em;
    text-transform: uppercase;
    color: color-mix(in srgb, var(--color-text) 50%, transparent);
    margin-bottom: 10px;
  }
  .atajos-lista {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .atajos-lista li {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12.5px;
    color: color-mix(in srgb, var(--color-text) 75%, transparent);
  }
  .atajos-lista li span {
    flex: 1;
    white-space: nowrap;
  }
  kbd {
    display: inline-block;
    padding: 2px 6px;
    font-family: var(--font-body);
    font-size: 11px;
    font-weight: 600;
    line-height: 1;
    color: var(--color-text);
    background: color-mix(in srgb, var(--color-text) 8%, transparent);
    border: 1px solid var(--color-divider);
    border-radius: 4px;
    box-shadow: 0 1px 1px rgba(0, 0, 0, 0.05);
  }

  .lupa {
    position: absolute;
    left: 16px;
    top: 50%;
    transform: translateY(-50%);
    opacity: 0.45;
  }
  .entrada {
    font-size: 16px;
    min-height: 52px;
    padding-left: 48px;
    background: var(--color-bg-card, #ffffff);
    border-radius: var(--radius-lg, 12px);
    box-shadow: var(--shadow-sm);
  }
  .desplegable {
    position: absolute;
    top: calc(100% + 6px);
    left: 0;
    right: 0;
    z-index: 30;
    background: #ffffff;
    border: 1px solid var(--color-divider);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-md);
    max-height: 280px;
    overflow-y: auto;
  }
  .resultado {
    display: flex;
    align-items: center;
    gap: 12px;
    width: 100%;
    padding: 12px 16px;
    border: 0;
    border-bottom: 1px solid var(--color-divider);
    background: transparent;
    font-family: var(--font-body);
    font-size: 15px;
    cursor: pointer;
    text-align: left;
    color: var(--color-text);
  }
  .resultado:hover { background: color-mix(in srgb, var(--color-accent) 10%, transparent); }
  .res-nombre { flex: 1; font-weight: 500; }
  .res-stock { font-size: 12px; }
  .res-precio { font-feature-settings: 'tnum'; min-width: 80px; text-align: right; font-weight: 600; }

  /* ESTRUCTURA PRINCIPAL EN 2 COLUMNAS */
  .grilla-vender {
    display: grid;
    grid-template-columns: 1fr 380px;
    gap: 24px;
    align-items: start;
  }

  /* COLUMNA IZQUIERDA: PRODUCTOS EN EL TICKET */
  .ticket-productos {
    padding: var(--space-5);
    gap: var(--space-3);
    background: #ffffff;
    min-height: 420px;
  }
  .ticket-cabecera {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .cabecera-izq {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .ticket-cabecera h4 { margin: 0; font-size: 20px; }
  .badge-items {
    font-size: 12px;
    background: var(--color-accent-50, #f0f7f7);
    color: var(--color-accent-700, #1f5459);
    padding: 2px 8px;
    border-radius: 12px;
    font-weight: 600;
  }
  .chico { font-size: 12px; min-height: 30px; }
  .sin-margen { margin: 0; }
  .vacio {
    font-size: 14px;
    margin: 0;
    text-align: center;
    padding: 48px 0;
    line-height: 1.5;
  }

  .lineas {
    display: flex;
    flex-direction: column;
    gap: 8px;
    max-height: 480px;
    overflow-y: auto;
  }
  .linea {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 12px;
    border: 1px solid var(--color-divider);
    border-radius: var(--radius-md);
    background: var(--color-bg, #f2f4f4);
    transition: background 0.15s, border-color 0.15s;
  }
  .linea.activa {
    border-color: var(--color-accent-400, #6cabb1);
    background: var(--color-accent-50, #f0f7f7);
  }
  .contador {
    display: flex;
    align-items: center;
    border: 1px solid var(--color-divider);
    border-radius: var(--radius-md);
    background: #ffffff;
    overflow: hidden;
    flex: none;
  }
  .paso {
    width: 32px;
    height: 32px;
    border: 0;
    background: transparent;
    cursor: pointer;
    font-size: 16px;
    font-weight: 600;
    color: var(--color-text);
  }
  .paso:hover { background: color-mix(in srgb, var(--color-accent) 12%, transparent); }
  .cant-input {
    width: 42px;
    height: 32px;
    border: 0;
    text-align: center;
    font-size: 14px;
    font-weight: 600;
    font-feature-settings: 'tnum';
    color: var(--color-text);
    -moz-appearance: textfield;
    appearance: textfield;
  }
  .cant-input::-webkit-outer-spin-button,
  .cant-input::-webkit-inner-spin-button {
    -webkit-appearance: none;
    margin: 0;
  }
  .linea-info {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 0;
  }
  .linea-nombre {
    font-size: 14px;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .linea-unitario {
    font-size: 12px;
  }
  .linea-sub {
    font-size: 15px;
    font-weight: 700;
    font-feature-settings: 'tnum';
    color: var(--color-accent-800, #153c40);
  }
  .btn-eliminar-linea {
    border: none;
    background: transparent;
    color: var(--color-text-muted, #718096);
    padding: 6px;
    border-radius: 50%;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: color 0.15s, background 0.15s;
  }
  .btn-eliminar-linea:hover {
    background: rgba(229, 62, 62, 0.15);
    color: var(--color-peligro, #e53e3e);
  }

  /* COLUMNA DERECHA: TOTAL / MÉTODOS DE PAGO / COBRAR */
  .ticket-resumen {
    padding: var(--space-5);
    gap: var(--space-4);
    background: #ffffff;
    position: sticky;
    top: 20px;
  }
  .resumen-cabecera {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding-bottom: var(--space-3);
    border-bottom: 1px solid var(--color-divider);
  }
  .total-label {
    font-size: 13px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    font-weight: 600;
  }
  .total-monto {
    font-family: var(--font-heading);
    font-size: 44px;
    font-weight: 700;
    line-height: 1;
    font-feature-settings: 'tnum';
    color: var(--color-accent-700, #1f5459);
  }
  .campo-pago {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .pago-label {
    font-size: 12.5px;
    font-weight: 600;
    color: color-mix(in srgb, var(--color-text) 68%, transparent);
  }
  .pago { width: 100%; }
  .pago .seg-opt { flex: 1; }

  /* CAJA Y CÁLCULO DE VUELTO */
  .caja-efectivo {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    padding: var(--space-3);
    background: var(--color-bg, #f2f4f4);
    border-radius: var(--radius-md);
    border: 1px solid var(--color-divider);
  }
  .campo-monto {
    position: relative;
    display: flex;
    align-items: center;
  }
  .simbolo {
    position: absolute;
    left: 12px;
    font-weight: 700;
    font-size: 16px;
    color: var(--color-accent-700);
  }
  .entrada-monto {
    padding-left: 28px;
    font-size: 16px;
    font-weight: 600;
  }
  .montos-rapidos {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
  }
  .banner-vuelto {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 14px;
    border-radius: var(--radius-md);
    font-weight: 700;
  }
  .banner-vuelto.exito {
    background: #e6f4ea;
    color: #137333;
    border: 1px solid #ceebd6;
  }
  .banner-vuelto.faltante {
    background: #fce8e6;
    color: #c5221f;
    border: 1px solid #f7c6c5;
  }
  .vuelto-label { font-size: 14px; }
  .vuelto-monto { font-size: 18px; font-feature-settings: 'tnum'; }

  .error { margin: 0; font-size: 13px; color: var(--color-peligro); }
  .cobrar { min-height: 56px; font-size: 19px; font-weight: 700; letter-spacing: 0.03em; }

  /* MODAL DE CONFIRMACIÓN */
  .confirmacion {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-3);
    text-align: center;
  }
  .conf-total {
    font-family: var(--font-heading);
    font-size: 40px;
    line-height: 1;
    font-feature-settings: 'tnum';
  }
  .conf-vuelto-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    padding: 12px 24px;
    background: #e6f4ea;
    border: 1px solid #ceebd6;
    border-radius: var(--radius-md);
    width: 100%;
  }
  .vuelto-titulo-conf {
    font-size: 13px;
    font-weight: 600;
    color: #137333;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }
  .vuelto-monto-conf {
    font-family: var(--font-heading);
    font-size: 32px;
    font-weight: 700;
    color: #137333;
    font-feature-settings: 'tnum';
  }
  .conf-acciones { justify-content: center; width: 100%; }
  .conf-boton { min-height: 48px; min-width: 180px; font-size: 16px; }
</style>
