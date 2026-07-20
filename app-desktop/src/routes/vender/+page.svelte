<script>
  import { onMount } from 'svelte';
  import BotonVolver from '$lib/componentes/BotonVolver.svelte';
  import Dialogo from '$lib/componentes/Dialogo.svelte';
  import { pedirApi } from '$lib/api.js';
  import { fmtPesos } from '$lib/formato.js';

  /** @typedef {{ id_producto: string, descripcion: string, cantidad_disponible: number, precio_venta: number }} Item */

  /** @type {Item[]} */
  let inventario = $state([]);
  /** @type {Item[]} */
  let frecuentes = $state([]);
  let query = $state('');
  /** @type {{ item: Item, cantidad: number }[]} */
  let carrito = $state([]);
  let pago = $state('efectivo');
  let cobrando = $state(false);
  let error = $state('');
  let ultimoTotal = $state(0);
  let verConfirmacion = $state(false);
  /** @type {HTMLInputElement | undefined} */
  let inputBusqueda = $state();

  onMount(async () => {
    try {
      [inventario, frecuentes] = await Promise.all([
        pedirApi('/api/stock'),
        pedirApi('/api/stock/frecuentes?limite=8'),
      ]);
    } catch {
      error = 'no se pudo cargar el inventario';
    }
    inputBusqueda?.focus();
  });

  // El catálogo del negocio es chico: se filtra en memoria mientras se tipea.
  const resultados = $derived.by(() => {
    const q = query.trim().toLowerCase();
    if (!q) return [];
    return inventario
      .filter((p) => p.descripcion.toLowerCase().includes(q) || p.id_producto.startsWith(q))
      .slice(0, 6);
  });

  /** @param {Item} item */
  function agregar(item) {
    const linea = carrito.find((l) => l.item.id_producto === item.id_producto);
    if (linea) linea.cantidad += 1;
    else carrito.push({ item, cantidad: 1 });
    query = '';
    inputBusqueda?.focus();
  }

  /** @param {string} id @param {number} delta */
  function ajustar(id, delta) {
    const linea = carrito.find((l) => l.item.id_producto === id);
    if (!linea) return;
    linea.cantidad += delta;
    if (linea.cantidad <= 0) carrito = carrito.filter((l) => l.item.id_producto !== id);
  }

  /** @param {KeyboardEvent} ev */
  function alTeclear(ev) {
    if (ev.key !== 'Enter') return;
    ev.preventDefault();
    // El escáner "tipea" el código y manda Enter: match exacto agrega directo.
    const codigo = query.trim();
    const exacto = inventario.find((p) => p.id_producto === codigo);
    if (exacto) agregar(exacto);
    else if (resultados.length === 1) agregar(resultados[0]);
  }

  const total = $derived(carrito.reduce((t, l) => t + l.item.precio_venta * l.cantidad, 0));

  async function cobrar() {
    if (carrito.length === 0 || cobrando) return;
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
      carrito = [];
      verConfirmacion = true;
      // El stock cambió: refrescar en segundo plano para el buscador.
      pedirApi('/api/stock').then((datos) => (inventario = datos)).catch(() => {});
    } catch (e) {
      error = e instanceof Error ? e.message : 'no se pudo cobrar';
    } finally {
      cobrando = false;
    }
  }

  function cerrarConfirmacion() {
    verConfirmacion = false;
    inputBusqueda?.focus();
  }
</script>

<BotonVolver />

<div class="pantalla">
  <div>
    <h2>Vender</h2>
    <p class="text-muted ayuda">
      Escaneá el código de barras o escribí el nombre. Cada venta descuenta el stock sola — no hay
      que anotar nada.
    </p>

    <div class="buscador">
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" class="lupa"><circle cx="11" cy="11" r="8"></circle><path d="m21 21-4.3-4.3"></path></svg>
      <input
        class="input entrada"
        bind:this={inputBusqueda}
        bind:value={query}
        onkeydown={alTeclear}
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

    <h6>Los de siempre</h6>
    <div class="frecuentes">
      {#each frecuentes as p (p.id_producto)}
        <button class="frecuente" onclick={() => agregar(p)}>
          <span class="fre-nombre">{p.descripcion}</span>
          <span class="fre-precio">{fmtPesos(p.precio_venta)}</span>
        </button>
      {/each}
    </div>
  </div>

  <div class="card elev-sm ticket">
    <div class="ticket-cabecera">
      <h4>Ticket</h4>
      <button class="btn btn-ghost chico" onclick={() => (carrito = [])}>Vaciar</button>
    </div>
    <hr class="hr sin-margen" />

    {#if carrito.length === 0}
      <p class="text-muted vacio">Todavía no hay nada.<br />Tocá un producto para empezar.</p>
    {/if}

    <div class="lineas">
      {#each carrito as l (l.item.id_producto)}
        <div class="linea">
          <div class="contador">
            <button class="paso" onclick={() => ajustar(l.item.id_producto, -1)}>−</button>
            <span class="cant">{l.cantidad}</span>
            <button class="paso" onclick={() => ajustar(l.item.id_producto, 1)}>+</button>
          </div>
          <span class="linea-nombre">{l.item.descripcion}</span>
          <span class="linea-sub">{fmtPesos(l.item.precio_venta * l.cantidad)}</span>
        </div>
      {/each}
    </div>

    <div class="total-fila">
      <span class="text-muted total-label">Total</span>
      <span class="total-monto">{fmtPesos(total)}</span>
    </div>

    <div class="seg pago">
      <label class="seg-opt"><input type="radio" name="pago" value="efectivo" bind:group={pago} />Efectivo</label>
      <label class="seg-opt"><input type="radio" name="pago" value="transferencia" bind:group={pago} />Transferencia</label>
      <label class="seg-opt"><input type="radio" name="pago" value="tarjeta" bind:group={pago} />Tarjeta</label>
    </div>

    {#if error}
      <p class="error">{error}</p>
    {/if}

    <button class="btn btn-primary btn-block cobrar" onclick={cobrar} disabled={carrito.length === 0 || cobrando}>
      {cobrando ? 'Cobrando…' : 'Cobrar'}
    </button>
  </div>
</div>

<Dialogo abierto={verConfirmacion}>
  <div class="confirmacion">
    <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="var(--color-accent)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><path d="m9 12 2 2 4-4"></path></svg>
    <div class="dialog-title">Venta cobrada</div>
    <div class="conf-total">{fmtPesos(ultimoTotal)}</div>
    <p class="dialog-body">Se descontó el stock y quedó anotada en Ventas.</p>
    <div class="dialog-actions conf-acciones">
      <button class="btn btn-primary conf-boton" onclick={cerrarConfirmacion}>Vender otra</button>
    </div>
  </div>
</Dialogo>

<style>
  .pantalla {
    display: grid;
    grid-template-columns: 1fr 390px;
    gap: 32px;
    max-width: 1240px;
  }
  h2 { margin-bottom: 6px; }
  .ayuda { font-size: 13.5px; max-width: 52ch; margin-top: 0; }
  h6 { color: var(--color-accent); margin-bottom: 12px; }

  .buscador { position: relative; margin: 14px 0 26px; }
  .lupa {
    position: absolute;
    left: 14px;
    top: 50%;
    transform: translateY(-50%);
    opacity: 0.45;
  }
  .entrada {
    font-size: 16px;
    min-height: 50px;
    padding-left: 44px;
    background: var(--color-bg);
  }
  .desplegable {
    position: absolute;
    top: calc(100% + 4px);
    left: 0;
    right: 0;
    z-index: 20;
    background: var(--color-bg);
    border: 1px solid var(--color-divider);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-md);
    overflow: hidden;
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
  .res-nombre { flex: 1; }
  .res-stock { font-size: 12px; }
  .res-precio { font-feature-settings: 'tnum'; min-width: 80px; text-align: right; }

  .frecuentes {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 12px;
  }
  .frecuente {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 6px;
    min-height: 76px;
    padding: 13px 15px;
    border: 1px solid var(--color-divider);
    border-radius: var(--radius-md);
    background: transparent;
    font-family: var(--font-body);
    cursor: pointer;
    text-align: left;
    color: var(--color-text);
  }
  .frecuente:hover {
    border-color: var(--color-accent);
    background: color-mix(in srgb, var(--color-accent) 7%, transparent);
  }
  .frecuente:active { background: color-mix(in srgb, var(--color-accent) 16%, transparent); }
  .fre-nombre { font-size: 14.5px; line-height: 1.3; }
  .fre-precio {
    font-family: var(--font-heading);
    font-weight: 600;
    font-size: 17px;
    font-feature-settings: 'tnum';
    color: var(--color-accent-700);
  }

  .ticket {
    align-self: start;
    position: sticky;
    top: 0;
    padding: var(--space-4);
    gap: var(--space-3);
  }
  .ticket-cabecera {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
  }
  .ticket-cabecera h4 { margin: 0; }
  .chico { font-size: 12px; min-height: 30px; }
  .sin-margen { margin: 0; }
  .vacio {
    font-size: 13.5px;
    margin: 6px 0;
    text-align: center;
    padding: 22px 0;
  }
  .lineas { display: flex; flex-direction: column; }
  .linea {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 9px 0;
    border-bottom: 1px solid var(--color-divider);
  }
  .contador {
    display: flex;
    align-items: center;
    border: 1px solid var(--color-divider);
    border-radius: var(--radius-md);
    overflow: hidden;
    flex: none;
  }
  .paso {
    width: 30px;
    height: 30px;
    border: 0;
    background: transparent;
    cursor: pointer;
    font-size: 16px;
    color: var(--color-text);
  }
  .paso:hover { background: color-mix(in srgb, var(--color-accent) 12%, transparent); }
  .cant { width: 26px; text-align: center; font-size: 14px; font-feature-settings: 'tnum'; }
  .linea-nombre { flex: 1; font-size: 14px; line-height: 1.3; }
  .linea-sub { font-size: 14px; font-feature-settings: 'tnum'; }

  .total-fila {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    margin-top: 4px;
  }
  .total-label {
    font-size: 13px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }
  .total-monto {
    font-family: var(--font-heading);
    font-size: 44px;
    line-height: 1;
    font-feature-settings: 'tnum';
  }
  .pago { width: 100%; }
  .pago .seg-opt { flex: 1; }
  .error { margin: 0; font-size: 13px; color: var(--color-peligro); }
  .cobrar { min-height: 56px; font-size: 19px; letter-spacing: 0.03em; }

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
  .conf-acciones { justify-content: center; width: 100%; }
  .conf-boton { min-height: 48px; min-width: 180px; font-size: 16px; }
</style>
