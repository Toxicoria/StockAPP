<script>
  // @ts-nocheck
  import { onMount } from 'svelte'; import BotonVolver from '$lib/componentes/BotonVolver.svelte'; import { pedirApi } from '$lib/api.js'; import { fmtPesos } from '$lib/formato.js';
  let datos = $state({ total_dia: 0, cantidad: 0, promedio: 0, ventas: [] }); let error = $state('');
  onMount(cargar); async function cargar(){try{datos=await pedirApi('/api/ventas')}catch(e){error=e.message}}
  async function facturar(v){try{const r=await pedirApi(`/api/ventas/${v.id_venta}/factura`,{method:'POST'});v.factura=r.factura}catch(e){error=e.message}}
</script>
<BotonVolver /><div class="cabecera"><div><h2>Ventas</h2><p class="text-muted">La caja de hoy, venta por venta.</p></div></div>
<div class="stats"><div class="card"><span class="card-kicker">Total del día</span><strong>{fmtPesos(datos.total_dia)}</strong></div><div class="card"><span class="card-kicker">Ventas</span><strong>{datos.cantidad}</strong></div><div class="card"><span class="card-kicker">Ticket promedio</span><strong>{fmtPesos(datos.promedio)}</strong></div></div>
{#if error}<p class="error">{error}</p>{/if}<div class="card tabla"><table class="table"><thead><tr><th>Hora</th><th>Detalle</th><th>Pago</th><th>Total</th><th>Factura</th></tr></thead><tbody>{#each datos.ventas as v (v.id_venta)}<tr><td>{v.hora}</td><td>{v.detalle}</td><td><span class="tag tag-neutral">{v.metodo_pago}</span></td><td>{fmtPesos(v.total_venta)}</td><td>{#if v.factura}<span class="tag tag-accent">{v.factura}</span>{:else}<button class="btn btn-secondary" onclick={() => facturar(v)}>Facturar</button>{/if}</td></tr>{/each}</tbody></table></div>
<style>.cabecera{margin-bottom:18px}.cabecera p{margin:0}.stats{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:14px;margin-bottom:18px}.stats strong{font-size:22px}.tabla{padding:4px;overflow:auto}.error{color:var(--color-peligro)}</style>
