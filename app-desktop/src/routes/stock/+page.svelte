<script>
  // @ts-nocheck
  import { onMount } from 'svelte';
  import BotonVolver from '$lib/componentes/BotonVolver.svelte';
  import EditorProducto from '$lib/componentes/EditorProducto.svelte';
  import { pedirApi } from '$lib/api.js';
  import { fmtPesos } from '$lib/formato.js';
  import { esDueno } from '$lib/sesion.svelte.js';

  let items = $state([]);
  let proveedores = $state([]);
  let buscar = $state('');
  let editor = $state(null);
  let error = $state('');

  onMount(cargar);

  async function cargar() {
    try {
      const [stk, prov] = await Promise.all([
        pedirApi('/api/stock'),
        pedirApi('/api/proveedores').catch(() => []),
      ]);
      items = stk ?? [];
      proveedores = prov ?? [];
    } catch (e) {
      error = e.message;
    }
  }

  async function ajustar(item, delta) {
    const anterior = item.cantidad_disponible;
    item.cantidad_disponible = Math.max(0, anterior + delta);
    try {
      const r = await pedirApi(`/api/stock/${encodeURIComponent(item.id_producto)}`, {
        method: 'PUT',
        body: { delta },
      });
      item.cantidad_disponible = r.cantidad_disponible;
    } catch (e) {
      item.cantidad_disponible = anterior;
      error = e.message;
    }
  }

  async function guardarProducto(datos, id, mantieneAbierto = false) {
    if (id) {
      delete datos.codigo_barras;
      await pedirApi(`/api/productos/${encodeURIComponent(id)}`, {
        method: 'PUT',
        body: datos,
      });
    } else {
      await pedirApi('/api/productos', {
        method: 'POST',
        body: datos,
      });
    }
    if (!mantieneAbierto) {
      editor = null;
    }
    await cargar();
  }

  const filtrados = $derived(
    items.filter((p) =>
      `${p.descripcion} ${p.id_producto}`.toLowerCase().includes(buscar.toLowerCase())
    )
  );

  const alertas = $derived(
    items.filter((p) => p.stock_minimo > 0 && p.cantidad_disponible <= p.stock_minimo)
  );
</script>

<BotonVolver />
<div class="cabecera">
  <div>
    <h2>Stock</h2>
    <p class="text-muted">Ajustá cantidades desde el mostrador. Los cambios se guardan al instante.</p>
  </div>
  <div class="acciones-cab">
    <span class="tag tag-neutral">{items.length} productos</span>
    {#if esDueno()}
      <button class="btn btn-primary" onclick={() => (editor = {})}>
        + Agregar producto
      </button>
    {/if}
  </div>
</div>

<div class="herramientas">
  <input class="input buscar" bind:value={buscar} placeholder="Buscar por nombre o código…" />
</div>

{#if alertas.length}
  <section class="card alertas">
    <strong>Para reponer ({alertas.length}):</strong>
    <div>
      {#each alertas as p}
        <span class="tag tag-accent">{p.descripcion}: {p.cantidad_disponible}</span>
      {/each}
    </div>
  </section>
{/if}

{#if error}
  <p class="error">{error}</p>
{/if}

<div class="card tabla">
  <table class="table">
    <thead>
      <tr>
        <th>Producto</th>
        <th>Proveedor</th>
        <th>Precio</th>
        <th>Stock</th>
        <th>Ajuste rápido</th>
      </tr>
    </thead>
    <tbody>
      {#each filtrados as p (p.id_producto)}
        <tr>
          <td>
            <strong>{p.descripcion}</strong>
            {#if p.marca || p.cantidad_presentacion}
              <span class="det-presentacion">
                · {p.marca ? p.marca + ' ' : ''}{p.cantidad_presentacion ?? ''} {p.unidad_medida ?? ''}
              </span>
            {/if}
            <br />
            <small class="text-muted">{p.id_producto}</small>
          </td>
          <td>{p.proveedor || '—'}</td>
          <td>{fmtPesos(p.precio_venta)}</td>
          <td>
            <span
              class:tag={p.stock_minimo > 0 && p.cantidad_disponible <= p.stock_minimo}
              class:tag-accent={p.stock_minimo > 0 && p.cantidad_disponible <= p.stock_minimo}
            >
              {p.cantidad_disponible}
            </span>
          </td>
          <td>
            <button class="btn btn-secondary mini" onclick={() => ajustar(p, -1)}>−</button>
            <button class="btn btn-secondary mini" onclick={() => ajustar(p, 1)}>+</button>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>

{#if editor}
  <EditorProducto
    producto={editor && Object.keys(editor).length ? editor : null}
    {proveedores}
    alGuardar={guardarProducto}
    alCancelar={() => (editor = null)}
  />
{/if}

<style>
  .cabecera {
    display: flex;
    justify-content: space-between;
    align-items: start;
    margin-bottom: 18px;
  }
  .cabecera p {
    margin: 0;
  }
  .acciones-cab {
    display: flex;
    align-items: center;
    gap: 12px;
  }
  .herramientas {
    margin-bottom: 16px;
  }
  .buscar {
    max-width: 460px;
  }
  .alertas {
    margin-bottom: 16px;
  }
  .alertas div {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }
  .tabla {
    padding: 4px;
    overflow: auto;
  }
  .det-presentacion {
    font-size: 12.5px;
    color: color-mix(in srgb, var(--color-text) 65%, transparent);
    font-weight: normal;
  }
  .mini {
    min-height: 30px;
    padding: 3px 10px;
    margin-left: 4px;
  }
  .error {
    color: var(--color-peligro);
  }
</style>
