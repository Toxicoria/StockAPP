<script>
  // @ts-nocheck
  import { onMount, tick } from 'svelte';
  import { pedirApi } from '$lib/api.js';
  import { agregarToast } from '$lib/toast.svelte.js';

  let { producto = null, proveedores = [], alGuardar, alCancelar } = $props();

  let datos = $state(
    producto
      ? {
          ...producto,
          codigo_barras: producto.id_producto,
          cantidad_presentacion: producto.cantidad_presentacion ?? '',
          unidad_medida: producto.unidad_medida ?? '',
          marca: producto.marca ?? '',
        }
      : {
          descripcion: '',
          codigo_barras: '',
          marca: '',
          cantidad_presentacion: '',
          unidad_medida: 'unidad',
          precio_venta: 0,
          cantidad_disponible: 1,
          stock_minimo: 5,
          id_proveedor: '',
        }
  );

  // Modo escáner continuo (carga masiva inicial)
  let modoMasivo = $state(false);

  let error = $state('');
  let guardando = $state(false);
  let intentoGuardar = $state(false);

  // Autocompletado del catálogo maestro
  let sugerencias = $state([]);
  let buscandoMaestro = $state(false);
  let mostrarSugerencias = $state(false);
  let esMaestro = $state(false);
  let timerBusqueda = null;

  // Referencias a inputs para control de foco
  let inputCodigoRef = $state(null);
  let inputNombreRef = $state(null);
  let inputPrecioRef = $state(null);
  let inputStockRef = $state(null);

  // Validaciones
  const descInvalida = $derived(intentoGuardar && !datos.descripcion.trim());
  const precioInvalido = $derived(intentoGuardar && (Number(datos.precio_venta) <= 0 || isNaN(Number(datos.precio_venta))));
  const stockInvalido = $derived(intentoGuardar && (Number(datos.cantidad_disponible) < 0 || isNaN(Number(datos.cantidad_disponible))));

  onMount(() => {
    if (!producto) {
      setTimeout(() => inputCodigoRef?.focus(), 80);
    } else {
      setTimeout(() => inputPrecioRef?.focus(), 80);
    }
  });

  // Búsqueda al escribir en el nombre
  function alCambiarNombre(valor) {
    datos.descripcion = valor;
    if (producto) return;

    esMaestro = false;
    if (timerBusqueda) clearTimeout(timerBusqueda);

    const texto = valor.trim();
    if (texto.length < 2) {
      sugerencias = [];
      mostrarSugerencias = false;
      return;
    }

    timerBusqueda = setTimeout(async () => {
      buscandoMaestro = true;
      try {
        const res = await pedirApi(`/api/productos?buscar=${encodeURIComponent(texto)}&limite=8`);
        sugerencias = res ?? [];
        mostrarSugerencias = sugerencias.length > 0;
      } catch (e) {
        sugerencias = [];
      } finally {
        buscandoMaestro = false;
      }
    }, 200);
  }

  // Búsqueda directa por escáner / código de barras
  async function buscarPorCodigo(codigo) {
    if (!codigo || producto) return;
    const ean = codigo.trim();
    if (ean.length < 3) return;

    buscandoMaestro = true;
    error = '';
    try {
      const res = await pedirApi(`/api/productos?buscar=${encodeURIComponent(ean)}&limite=5`);
      if (res && res.length > 0) {
        const coincidencia = res.find((p) => p.id_producto === ean) || res[0];
        seleccionarSugerencia(coincidencia);
      } else {
        esMaestro = false;
        mostrarSugerencias = false;
        await tick();
        inputNombreRef?.focus();
      }
    } catch (e) {
      // Ignorar error de red puntual en escaneo
    } finally {
      buscandoMaestro = false;
    }
  }

  function alKeyDownCodigo(ev) {
    if (ev.key === 'Enter') {
      ev.preventDefault();
      buscarPorCodigo(datos.codigo_barras);
    }
  }

  function seleccionarSugerencia(item) {
    datos.descripcion = item.descripcion;
    datos.codigo_barras = item.id_producto;
    datos.marca = item.marca ?? '';
    datos.cantidad_presentacion = item.cantidad_presentacion ?? '';
    datos.unidad_medida = item.unidad_medida ?? 'unidad';

    esMaestro = true;
    mostrarSugerencias = false;
    sugerencias = [];

    tick().then(() => {
      if (inputPrecioRef) {
        inputPrecioRef.focus();
        inputPrecioRef.select();
      }
    });
  }

  function cerrarSugerencias() {
    setTimeout(() => {
      mostrarSugerencias = false;
    }, 200);
  }

  async function guardar() {
    intentoGuardar = true;
    error = '';

    if (!datos.descripcion.trim() || Number(datos.precio_venta) <= 0) {
      error = 'Revisá los campos marcados en rojo antes de guardar.';
      agregarToast(error, 'error');
      return;
    }

    guardando = true;
    try {
      const nombreGuardado = datos.descripcion;
      const cuerpo = {
        ...datos,
        descripcion: datos.descripcion.trim(),
        codigo_barras: datos.codigo_barras.trim() || undefined,
        marca: datos.marca ? datos.marca.trim() : undefined,
        cantidad_presentacion: datos.cantidad_presentacion ? String(datos.cantidad_presentacion).trim() : undefined,
        unidad_medida: datos.unidad_medida ? String(datos.unidad_medida).trim() : undefined,
        precio_venta: Number(datos.precio_venta) || 0,
        cantidad_disponible: Number(datos.cantidad_disponible) || 0,
        stock_minimo: Number(datos.stock_minimo) || 0,
        id_proveedor: datos.id_proveedor ? Number(datos.id_proveedor) : undefined,
      };

      await alGuardar(cuerpo, producto?.id_producto, modoMasivo);

      agregarToast(
        producto
          ? `Producto "${nombreGuardado}" actualizado correctamente`
          : `Producto "${nombreGuardado}" agregado correctamente`,
        'exito'
      );

      if (modoMasivo && !producto) {
        datos.descripcion = '';
        datos.codigo_barras = '';
        datos.marca = '';
        datos.cantidad_presentacion = '';
        datos.unidad_medida = 'unidad';
        datos.precio_venta = 0;
        datos.cantidad_disponible = 1;
        esMaestro = false;
        intentoGuardar = false;

        await tick();
        inputCodigoRef?.focus();
      }
    } catch (e) {
      error = e.message ?? 'Error al guardar el producto';
      agregarToast(error, 'error');
    } finally {
      guardando = false;
    }
  }
</script>

<div class="dialog-backdrop">
  <section class="dialog modal-producto">
    <div class="cabecera-dialogo">
      <div>
        <h3 class="dialog-title">{producto ? 'Editar producto' : 'Cargar producto en inventario'}</h3>
        <p class="subtitulo-dialogo">
          {producto
            ? 'Modificá precios, stock o presentación de este producto.'
            : 'Escaneá el código de barras o buscá por nombre en el catálogo.'}
        </p>
      </div>
      {#if !producto}
        <span class="tag" class:tag-accent={esMaestro} class:tag-neutral={!esMaestro}>
          {esMaestro ? 'Catálogo maestro' : 'Producto personalizado'}
        </span>
      {/if}
    </div>

    <!-- Toggle Carga Masiva (Modo Escáner Continuo) -->
    {#if !producto}
      <div class="barra-modo-masivo">
        <label class="checkbox-label">
          <input type="checkbox" bind:checked={modoMasivo} />
          <span>
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="var(--color-accent-600)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="icono-zap"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon></svg>
            <strong>Modo Carga Masiva</strong> (mantener ventana abierta para escanear varios productos seguidos)
          </span>
        </label>
      </div>
    {/if}

    <div class="campos">
      <!-- 1. Código de Barras (Escáner) -->
      <div class="field campo-full pos-rel">
        <label for="p_codigo">
          Código de Barras (EAN / Escáner)
          <span class="text-muted font-normal">(Presioná Enter o escaneá)</span>
        </label>
        <div class="input-contenedor">
          <input
            id="p_codigo"
            bind:this={inputCodigoRef}
            class="input input-codigo"
            type="text"
            bind:value={datos.codigo_barras}
            onkeydown={alKeyDownCodigo}
            onblur={() => { if (datos.codigo_barras && !esMaestro) buscarPorCodigo(datos.codigo_barras); }}
            disabled={!!producto}
            placeholder="Escanear producto aquí..."
          />
          <button
            class="btn btn-secondary btn-buscar-ean"
            type="button"
            onclick={() => buscarPorCodigo(datos.codigo_barras)}
            disabled={!!producto || !datos.codigo_barras.trim()}
          >
            Buscar EAN
          </button>
        </div>
      </div>

      <!-- 2. Nombre con Autocompletado del catálogo maestro -->
      <div class="field campo-full pos-rel">
        <label for="p_nombre">
          Nombre / Descripción del producto *
        </label>
        <div class="input-contenedor">
          <input
            id="p_nombre"
            bind:this={inputNombreRef}
            class="input"
            class:input-error={descInvalida}
            type="text"
            value={datos.descripcion}
            oninput={(e) => alCambiarNombre(e.target.value)}
            onfocus={() => { if (sugerencias.length > 0 && !producto) mostrarSugerencias = true; }}
            onblur={cerrarSugerencias}
            placeholder="Ej: Coca Cola Original, Pan Lactal, Galletitas Chocolinas..."
            autocomplete="off"
            required
          />
          {#if buscandoMaestro}
            <span class="spinner-cargando"></span>
          {/if}
        </div>
        {#if descInvalida}
          <span class="error-campo">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="var(--color-peligro)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>
            Ingresá el nombre o elegí una opción del catálogo.
          </span>
        {/if}

        {#if mostrarSugerencias && !producto}
          <ul class="dropdown-sugerencias">
            <li class="dropdown-header">Sugerencias encontradas en catálogo:</li>
            {#each sugerencias as sug}
              <!-- svelte-ignore a11y_click_events_have_key_events -->
              <li
                class="dropdown-item"
                onmousedown={() => seleccionarSugerencia(sug)}
              >
                <div class="sug-info">
                  <span class="sug-nombre">{sug.descripcion}</span>
                  {#if sug.marca || sug.cantidad_presentacion}
                    <span class="sug-det">
                      {sug.marca ?? ''} {sug.cantidad_presentacion ?? ''} {sug.unidad_medida ?? ''}
                    </span>
                  {/if}
                </div>
                <span class="sug-ean">EAN: {sug.id_producto}</span>
              </li>
            {/each}
          </ul>
        {/if}
      </div>

      <!-- 3. Marca y Presentación (Cantidad + Unidad de Medida) -->
      <div class="field">
        <label for="p_marca">Marca</label>
        <input id="p_marca" class="input" type="text" bind:value={datos.marca} placeholder="Ej: Coca Cola, Bagel, La Serenísima..." />
      </div>

      <div class="field">
        <label for="p_pres_cant">Presentación (Cant. + Unidad)</label>
        <div class="input-grupo-doble">
          <input
            id="p_pres_cant"
            class="input"
            type="text"
            bind:value={datos.cantidad_presentacion}
            placeholder="Ej: 1.5, 500, 1"
          />
          <select id="p_pres_uni" class="input input-uni" bind:value={datos.unidad_medida}>
            <option value="unidad">unidades</option>
            <option value="L">Litros (L)</option>
            <option value="ml">Mililitros (ml)</option>
            <option value="cm3">cm³</option>
            <option value="g">Gramos (g)</option>
            <option value="kg">Kilos (kg)</option>
            <option value="pack">Pack</option>
            <option value="cc">cc</option>
          </select>
        </div>
      </div>

      <!-- 4. Precio de Venta -->
      <div class="field">
        <label for="p_precio">
          Precio de Venta ($) *
        </label>
        <input
          id="p_precio"
          bind:this={inputPrecioRef}
          class="input"
          class:input-error={precioInvalido}
          type="number"
          step="0.01"
          min="0"
          bind:value={datos.precio_venta}
          placeholder="0.00"
          required
        />
        {#if precioInvalido}
          <span class="error-campo">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="var(--color-peligro)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>
            Ingresá el precio de venta mayor a $0.
          </span>
        {/if}
      </div>

      <!-- 5. Stock Inicial -->
      <div class="field">
        <label for="p_stock">
          Stock disponible *
        </label>
        <input
          id="p_stock"
          bind:this={inputStockRef}
          class="input"
          class:input-error={stockInvalido}
          type="number"
          step="1"
          min="0"
          bind:value={datos.cantidad_disponible}
          required
        />
        {#if stockInvalido}
          <span class="error-campo">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="var(--color-peligro)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>
            La cantidad no puede ser negativa.
          </span>
        {/if}
      </div>

      <!-- 6. Proveedor -->
      <div class="field">
        <label for="p_prov">Proveedor (Opcional)</label>
        <select id="p_prov" class="input" bind:value={datos.id_proveedor}>
          <option value="">Sin proveedor asignado</option>
          {#each proveedores as p}
            <option value={p.id_proveedor}>{p.nombre}</option>
          {/each}
        </select>
      </div>

      <!-- 7. Stock Mínimo -->
      <div class="field">
        <label for="p_minimo">Stock Mínimo (Alerta de reposición)</label>
        <input id="p_minimo" class="input" type="number" step="1" min="0" bind:value={datos.stock_minimo} />
      </div>
    </div>

    {#if error}
      <p class="error-general">{error}</p>
    {/if}

    <div class="dialog-actions">
      <button class="btn btn-secondary" onclick={alCancelar} type="button">
        {modoMasivo && !producto ? 'Cerrar ventana' : 'Cancelar'}
      </button>
      <button class="btn btn-primary btn-guardar-zap" onclick={guardar} disabled={guardando} type="button">
        {#if guardando}
          Guardando…
        {:else if modoMasivo}
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon></svg>
          Guardar y continuar escaneando
        {:else}
          Guardar en inventario
        {/if}
      </button>
    </div>
  </section>
</div>

<style>
  .modal-producto {
    width: min(620px, 100%);
    gap: 16px;
  }
  .cabecera-dialogo {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
  }
  .subtitulo-dialogo {
    margin: 4px 0 0;
    font-size: 12.5px;
    color: color-mix(in srgb, var(--color-text) 60%, transparent);
  }
  .barra-modo-masivo {
    padding: 10px 14px;
    background: var(--color-accent-100);
    border: 1px solid var(--color-accent-300);
    border-radius: var(--radius-md);
  }
  .checkbox-label {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    color: var(--color-accent-800);
    cursor: pointer;
  }
  .campos {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 14px;
  }
  .campo-full {
    grid-column: 1 / -1;
  }
  .pos-rel {
    position: relative;
  }
  .input-contenedor {
    position: relative;
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .input-codigo {
    font-family: monospace;
    font-weight: 600;
    letter-spacing: 0.05em;
  }
  .input-grupo-doble {
    display: flex;
    gap: 8px;
  }
  .input-uni {
    width: 120px;
  }
  .btn-buscar-ean {
    white-space: nowrap;
    min-height: 42px;
  }
  .font-normal {
    font-weight: normal;
  }
  .spinner-cargando {
    position: absolute;
    right: 12px;
    width: 16px;
    height: 16px;
    border: 2px solid var(--color-divider);
    border-top-color: var(--color-accent);
    border-radius: 50%;
    animation: spin 0.6s linear infinite;
  }
  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  /* Resaltado de errores en rojo */
  .input-error {
    border-color: var(--color-peligro) !important;
    background-color: color-mix(in srgb, var(--color-peligro) 4%, #fff) !important;
    box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-peligro) 30%, transparent) !important;
  }
  .error-campo {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    font-size: 11.5px;
    color: var(--color-peligro);
    margin-top: 4px;
    font-weight: 600;
  }
  .btn-guardar-zap {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .error-general {
    color: var(--color-peligro);
    font-size: 13px;
    margin: 0;
    font-weight: 600;
  }

  /* Desplegable Autocompletado */
  .dropdown-sugerencias {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    z-index: 100;
    margin: 4px 0 0;
    padding: 6px 0;
    list-style: none;
    background: #fff;
    border: 1px solid var(--color-accent-300);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-lg);
    max-height: 220px;
    overflow-y: auto;
  }
  .dropdown-header {
    padding: 4px 12px;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: color-mix(in srgb, var(--color-text) 50%, transparent);
    border-bottom: 1px solid var(--color-divider);
    margin-bottom: 4px;
  }
  .dropdown-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    cursor: pointer;
    transition: background 0.1s ease;
  }
  .dropdown-item:hover {
    background: color-mix(in srgb, var(--color-accent) 10%, transparent);
  }
  .sug-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .sug-nombre {
    font-size: 13.5px;
    font-weight: 600;
  }
  .sug-det {
    font-size: 11.5px;
    color: color-mix(in srgb, var(--color-text) 60%, transparent);
  }
  .sug-ean {
    font-size: 11.5px;
    font-family: monospace;
    background: var(--color-surface);
    padding: 2px 6px;
    border-radius: var(--radius-sm);
    color: var(--color-accent-700);
  }
</style>
