<script>
  // @ts-nocheck
  import { onMount } from 'svelte'; import BotonVolver from '$lib/componentes/BotonVolver.svelte'; import { pedirApi } from '$lib/api.js';
  let datos=$state(null); let claveFiscal=$state(''); let guardando=$state(false); let mensaje=$state(''); let error=$state('');
  onMount(cargar); async function cargar(){try{datos=await pedirApi('/api/negocio')}catch(e){error=e.message}}
  async function guardar(){
    guardando=true; mensaje=''; error='';
    const cuerpo={nombre_negocio:datos.nombre_negocio, direccion:datos.direccion, cuit:datos.cuit, punto_venta:datos.punto_venta};
    if(claveFiscal) cuerpo.clave_fiscal=claveFiscal;
    try{await pedirApi('/api/negocio',{method:'PUT',body:cuerpo}); claveFiscal=''; mensaje='Configuración guardada.'; await cargar()}
    catch(e){error=e.message} finally{guardando=false}
  }
  async function borrarClave(){
    if(!confirm('¿Borrar la clave fiscal guardada?')) return;
    guardando=true; error='';
    try{await pedirApi('/api/negocio',{method:'PUT',body:{clave_fiscal:''}}); mensaje='Clave fiscal borrada.'; await cargar()}
    catch(e){error=e.message} finally{guardando=false}
  }
</script>
<BotonVolver />
<h2>Configuración</h2>
<p class="text-muted">Datos del negocio y credenciales para la facturación electrónica.</p>
{#if mensaje}<p class="ok">{mensaje}</p>{/if}
{#if error}<p class="error">{error}</p>{/if}
{#if datos}
<section class="card seccion">
  <h4>Datos del negocio</h4>
  <div class="campos">
    <label class="field"><span>Nombre</span><input class="input" bind:value={datos.nombre_negocio} /></label>
    <label class="field"><span>CUIT</span><input class="input" bind:value={datos.cuit} placeholder="20-12345678-9" /></label>
    <label class="field"><span>Dirección</span><input class="input" bind:value={datos.direccion} /></label>
    <label class="field"><span>Punto de venta</span><input class="input" bind:value={datos.punto_venta} placeholder="0001" maxlength="4" /></label>
  </div>
</section>
<section class="card seccion">
  <h4>ARCA — Clave fiscal</h4>
  <p class="text-muted ayuda">Se guarda cifrada. Nunca se muestra ni se envía de vuelta una vez guardada — para cambiarla, escribí una nueva.</p>
  <div class="estado">
    <span class="tag" class:tag-accent={datos.clave_fiscal_configurada} class:tag-neutral={!datos.clave_fiscal_configurada}>
      {datos.clave_fiscal_configurada ? 'Configurada' : 'Sin configurar'}
    </span>
    {#if datos.clave_fiscal_configurada}<button class="btn btn-ghost peligro" onclick={borrarClave}>Borrar</button>{/if}
  </div>
  <label class="field"><span>{datos.clave_fiscal_configurada ? 'Reemplazar clave fiscal' : 'Clave fiscal'}</span>
    <input class="input" type="password" autocomplete="off" bind:value={claveFiscal} placeholder="••••••••" /></label>
</section>
<button class="btn btn-primary" onclick={guardar} disabled={guardando}>{guardando ? 'Guardando…' : 'Guardar'}</button>
{/if}
<style>
  .seccion{margin:18px 0;max-width:520px}
  .seccion h4{margin:0 0 4px}
  .ayuda{font-size:12.5px;margin:4px 0 12px}
  .campos{display:grid;grid-template-columns:1fr 1fr;gap:12px}
  .estado{display:flex;align-items:center;gap:10px;margin-bottom:12px}
  .error,.peligro{color:var(--color-peligro)}
  .ok{color:var(--color-accent-700)}
</style>
