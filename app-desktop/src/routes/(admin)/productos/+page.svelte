<script>
  // @ts-nocheck
  import { onMount } from 'svelte';
  import BotonVolver from '$lib/componentes/BotonVolver.svelte';
  import EditorProducto from '$lib/componentes/EditorProducto.svelte';
  import { pedirApi } from '$lib/api.js';
  import { fmtPesos } from '$lib/formato.js';

  let productos = $state([]);
  let proveedores = $state([]);
  let buscar = $state('');
  let editor = $state(null);
  let error = $state('');

  onMount(cargar);

  async function cargar() {
    try {
      [productos, proveedores] = await Promise.all([
        pedirApi('/api/stock'),
        pedirApi('/api/proveedores').catch(() => []),
      ]);
    } catch (e) {
      error = e.message;
    }
  }

  const filtrados = $derived(
    productos.filter((p) =>
      `${p.descripcion} ${p.id_producto}`.toLowerCase().includes(buscar.toLowerCase())
    )
  );

  async function guardar(datos, id, mantieneAbierto = false) {
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

  async function borrar(p) {
    if (!confirm(`¿Eliminar ${p.descripcion} del inventario?`)) return;
    try {
      await pedirApi(`/api/productos/${encodeURIComponent(p.id_producto)}`, {
        method: 'DELETE',
      });
      await cargar();
    } catch (e) {
      error = e.message;
    }
  }
</script>

<BotonVolver />
<div class="cabecera">
  <div>
    <h2>Productos</h2>
    <p class="text-muted">Catálogo e inventario del negocio.</p>
  </div>
  <button class="btn btn-primary" onclick={() => (editor = {})}>
    + Agregar producto
  </button>
</div>

<input class="input buscar" bind:value={buscar} placeholder="Buscar por nombre o código…" />

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
        <th></th>
      </tr>
    </thead>
    <tbody>
      {#each filtrados as p (p.id_producto)}
        <tr>
          <td>
            <strong>{p.descripcion}</strong><br />
            <small class="text-muted">{p.id_producto}</small>
          </td>
          <td>{p.proveedor || '—'}</td>
          <td>{fmtPesos(p.precio_venta)}</td>
          <td>
            {p.cantidad_disponible}
            {#if p.stock_minimo > 0 && p.cantidad_disponible <= p.stock_minimo}
              <span class="tag tag-accent"> Queda poco</span>
            {/if}
          </td>
          <td class="acciones">
            <button class="btn btn-ghost" onclick={() => (editor = p)}>Editar</button>
            <button class="btn btn-ghost peligro" onclick={() => borrar(p)}>Eliminar</button>
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
    alGuardar={guardar}
    alCancelar={() => (editor = null)}
  />
{/if}

<style>
  .cabecera {
    display: flex;
    justify-content: space-between;
    align-items: start;
  }
  .cabecera p {
    margin: 0;
  }
  .buscar {
    max-width: 460px;
    margin: 18px 0;
  }
  .tabla {
    padding: 4px;
    overflow: auto;
  }
  .acciones {
    display: flex;
    justify-content: flex-end;
    gap: 4px;
  }
  .error,
  .peligro {
    color: var(--color-peligro);
  }
</style>
