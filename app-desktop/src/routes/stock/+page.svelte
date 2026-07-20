<script>
  // @ts-nocheck
  import { onMount } from 'svelte';
  import BotonVolver from '$lib/componentes/BotonVolver.svelte';
  import { pedirApi } from '$lib/api.js';
  import { fmtPesos } from '$lib/formato.js';

  let items = $state([]);
  let error = $state('');
  onMount(cargar);
  async function cargar() { try { items = await pedirApi('/api/stock'); } catch (e) { error = e.message; } }
  async function ajustar(item, delta) {
    const anterior = item.cantidad_disponible;
    item.cantidad_disponible = Math.max(0, anterior + delta);
    try { const r = await pedirApi(`/api/stock/${encodeURIComponent(item.id_producto)}`, { method: 'PUT', body: { delta } }); item.cantidad_disponible = r.cantidad_disponible; }
    catch (e) { item.cantidad_disponible = anterior; error = e.message; }
  }
  const alertas = $derived(items.filter((p) => p.stock_minimo > 0 && p.cantidad_disponible <= p.stock_minimo));
</script>

<BotonVolver />
<div class="cabecera"><div><h2>Stock</h2><p class="text-muted">Ajustá cantidades desde el mostrador. Los cambios se guardan al instante.</p></div><span class="tag tag-neutral">{items.length} productos</span></div>
{#if alertas.length}<section class="card alertas"><strong>Para reponer</strong><div>{#each alertas as p}<span class="tag tag-accent">{p.descripcion}: {p.cantidad_disponible}</span>{/each}</div></section>{/if}
{#if error}<p class="error">{error}</p>{/if}
<div class="card tabla"><table class="table"><thead><tr><th>Producto</th><th>Proveedor</th><th>Precio</th><th>Stock</th><th></th></tr></thead><tbody>{#each items as p (p.id_producto)}<tr><td><strong>{p.descripcion}</strong><br><small class="text-muted">{p.id_producto}</small></td><td>{p.proveedor || '—'}</td><td>{fmtPesos(p.precio_venta)}</td><td><span class:tag={p.stock_minimo > 0 && p.cantidad_disponible <= p.stock_minimo} class:tag-accent={p.stock_minimo > 0 && p.cantidad_disponible <= p.stock_minimo}>{p.cantidad_disponible}</span></td><td><button class="btn btn-secondary mini" onclick={() => ajustar(p, -1)}>−</button><button class="btn btn-secondary mini" onclick={() => ajustar(p, 1)}>+</button></td></tr>{/each}</tbody></table></div>
<style>.cabecera{display:flex;justify-content:space-between;align-items:start;margin-bottom:18px}.cabecera p{margin:0}.alertas{margin-bottom:16px}.alertas div{display:flex;gap:8px;flex-wrap:wrap}.tabla{padding:4px;overflow:auto}.mini{min-height:30px;padding:3px 10px;margin-left:4px}.error{color:var(--color-peligro)}</style>
