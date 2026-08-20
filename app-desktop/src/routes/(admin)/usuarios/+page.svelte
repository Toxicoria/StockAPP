<script>
  import { pedirApi } from '$lib/api.js';
  import { esDueno, sesion } from '$lib/sesion.svelte.js';
  import { agregarToast } from '$lib/toast.svelte.js';
  import BotonVolver from '$lib/componentes/BotonVolver.svelte';
  import Dialogo from '$lib/componentes/Dialogo.svelte';

  /** @type {any[]} */
  let usuarios = $state([]);
  let cargando = $state(true);
  let error = $state('');

  // Wizard de creación / edición
  let modalAbierto = $state(false);
  let modoEdicion = $state(false);
  let editandoId = $state(0);
  let pasoWizard = $state(1);

  let formNombre = $state('');
  let formUsuario = $state('');
  let formEmail = $state('');
  let formPassword = $state('');
  let formConfirmPassword = $state('');
  /** @type {string[]} */
  let formPermisos = $state(['vender', 'stock']);

  // Errores por campo en Paso 1
  let errorNombre = $state('');
  let errorUsuario = $state('');
  let errorEmail = $state('');

  let guardando = $state(false);
  let errorModal = $state('');

  // Modal de Detalle de Usuario & Cambiar Contraseña
  let modalDetalleAbierto = $state(false);
  /** @type {any} */
  let usuarioSeleccionado = $state(null);
  let passNueva = $state('');
  let passConfirmar = $state('');
  let cambiandoPass = $state(false);
  let errorPass = $state('');

  // Confirmación de eliminar
  let confirmandoId = $state(0);
  let confirmandoNombre = $state('');
  let eliminando = $state(false);

  const REGEX_USUARIO = /^[a-zA-Z0-9._-]+$/;
  const REGEX_EMAIL = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;

  const MODULOS_DISPONIBLES = [
    { id: 'vender', nombre: 'Vender', desc: 'Pantalla de cobro en mostrador' },
    { id: 'stock', nombre: 'Stock', desc: 'Consulta e ingreso de mercadería' },
    { id: 'ventas', nombre: 'Historial de Ventas', desc: 'Caja del día y registro de operaciones' },
    { id: 'resumen', nombre: 'Resumen', desc: 'Métricas de ventas y rendimiento' },
    { id: 'productos', nombre: 'Productos', desc: 'Catálogo de productos y precios' },
    { id: 'precios', nombre: 'Aumento de Precios', desc: 'Actualización por proveedores' },
    { id: 'facturas', nombre: 'Facturación', desc: 'Comprobantes emitidos' },
    { id: 'usuarios', nombre: 'Usuarios', desc: 'Gestión de usuarios y accesos' },
    { id: 'configuracion', nombre: 'Configuración', desc: 'Datos del comercio' },
  ];

  const COLORES_ARCOIRIS = [
    '#e53935', // Rojo
    '#fb8c00', // Naranja
    '#fdd835', // Amarillo
    '#43a047', // Verde
    '#00acc1', // Cian
    '#1e88e5', // Azul
    '#8e24aa', // Violeta
    '#d81b60', // Rosa
  ];

  /** @param {string} texto */
  function obtenerColorArcoiris(texto) {
    if (!texto) return COLORES_ARCOIRIS[0];
    let hash = 0;
    for (let i = 0; i < texto.length; i++) {
      hash = texto.charCodeAt(i) + ((hash << 5) - hash);
    }
    const index = Math.abs(hash) % COLORES_ARCOIRIS.length;
    return COLORES_ARCOIRIS[index];
  }

  /** @param {string} nombre */
  function obtenerIniciales(nombre) {
    if (!nombre) return '?';
    const partes = nombre.trim().split(/\s+/);
    if (partes.length >= 2) {
      return (partes[0][0] + partes[1][0]).toUpperCase();
    }
    return nombre.substring(0, 2).toUpperCase();
  }

  async function cargar() {
    cargando = true;
    error = '';
    try {
      usuarios = await pedirApi('/api/usuarios');
    } catch (e) {
      error = e instanceof Error ? e.message : 'error cargando usuarios';
    } finally {
      cargando = false;
    }
  }

  function abrirNuevo() {
    modoEdicion = false;
    editandoId = 0;
    pasoWizard = 1;
    formNombre = '';
    formUsuario = '';
    formEmail = '';
    formPassword = '';
    formConfirmPassword = '';
    formPermisos = ['vender', 'stock'];
    errorNombre = '';
    errorUsuario = '';
    errorEmail = '';
    errorModal = '';
    modalAbierto = true;
  }

  /** @param {any} u */
  function abrirDetalle(u) {
    usuarioSeleccionado = u;
    passNueva = '';
    passConfirmar = '';
    errorPass = '';
    modalDetalleAbierto = true;
  }

  /** @param {any} u */
  function abrirEditarDesdeDetalle(u) {
    modalDetalleAbierto = false;
    abrirEditar(u);
  }

  /** @param {any} u */
  function abrirEditar(u) {
    modoEdicion = true;
    editandoId = u.id_usuario;
    pasoWizard = 1;
    formNombre = u.nombre;
    formUsuario = u.usuario || '';
    formEmail = u.email || '';
    formPassword = '';
    formConfirmPassword = '';
    formPermisos = Array.isArray(u.permisos) ? u.permisos : ['vender', 'stock'];
    errorNombre = '';
    errorUsuario = '';
    errorEmail = '';
    errorModal = '';
    modalAbierto = true;
  }

  function siguientePaso() {
    errorNombre = '';
    errorUsuario = '';
    errorEmail = '';
    errorModal = '';

    if (pasoWizard === 1) {
      let tieneError = false;
      const nombreClean = formNombre.trim();
      const usuarioClean = formUsuario.trim();
      const emailClean = formEmail.trim();

      if (!nombreClean) {
        errorNombre = 'El nombre del empleado es obligatorio';
        tieneError = true;
      }

      if (!usuarioClean) {
        errorUsuario = 'El nombre de usuario es obligatorio';
        tieneError = true;
      } else if (!REGEX_USUARIO.test(usuarioClean)) {
        errorUsuario = 'Sin espacios ni caracteres especiales (ej: acoria)';
        tieneError = true;
      }

      if (emailClean && !REGEX_EMAIL.test(emailClean)) {
        errorEmail = 'Formato de correo inválido (ej: usuario@mail.com)';
        tieneError = true;
      }

      if (tieneError) return;

      if (modoEdicion) {
        pasoWizard = 3;
      } else {
        pasoWizard = 2;
      }
    } else if (pasoWizard === 2) {
      if (!formPassword) {
        errorModal = 'Ingresá la contraseña para el nuevo usuario';
        return;
      }
      if (formPassword.length < 8) {
        errorModal = 'La contraseña debe tener al menos 8 caracteres';
        return;
      }
      if (formPassword !== formConfirmPassword) {
        errorModal = 'Las contraseñas no coinciden';
        return;
      }
      pasoWizard = 3;
    } else if (pasoWizard === 3) {
      if (formPermisos.length === 0) {
        errorModal = 'Debes seleccionar al menos un permiso de vista';
        return;
      }
      pasoWizard = 4;
    }
  }

  function anteriorPaso() {
    errorModal = '';
    if (pasoWizard === 3 && modoEdicion) {
      pasoWizard = 1;
    } else if (pasoWizard > 1) {
      pasoWizard -= 1;
    }
  }

  /** @param {string} modId */
  function togglePermiso(modId) {
    if (formPermisos.includes(modId)) {
      formPermisos = formPermisos.filter((p) => p !== modId);
    } else {
      formPermisos = [...formPermisos, modId];
    }
  }

  function seleccionarTodosPermisos() {
    formPermisos = MODULOS_DISPONIBLES.map((m) => m.id);
  }

  function seleccionarBasicosPermisos() {
    formPermisos = ['vender', 'stock'];
  }

  async function guardar() {
    guardando = true;
    errorModal = '';
    try {
      if (modoEdicion) {
        await pedirApi(`/api/usuarios/${editandoId}`, {
          method: 'PUT',
          body: {
            nombre: formNombre.trim(),
            usuario: formUsuario.trim() || null,
            email: formEmail.trim() || null,
            permisos: formPermisos,
          },
        });
        agregarToast('Usuario actualizado correctamente', 'exito');
      } else {
        await pedirApi('/api/usuarios', {
          method: 'POST',
          body: {
            nombre: formNombre.trim(),
            usuario: formUsuario.trim() || null,
            email: formEmail.trim() || null,
            password: formPassword,
            permisos: formPermisos,
          },
        });
        agregarToast('Nuevo cajero creado con éxito', 'exito');
      }
      modalAbierto = false;
      await cargar();
    } catch (e) {
      errorModal = e instanceof Error ? e.message : 'error guardando';
    } finally {
      guardando = false;
    }
  }

  async function actualizarContrasena() {
    errorPass = '';
    if (!passNueva) {
      errorPass = 'Ingresá la nueva contraseña';
      return;
    }
    if (passNueva.length < 8) {
      errorPass = 'La contraseña debe tener al menos 8 caracteres';
      return;
    }
    if (passNueva !== passConfirmar) {
      errorPass = 'Las contraseñas no coinciden';
      return;
    }

    cambiandoPass = true;
    try {
      await pedirApi(`/api/usuarios/${usuarioSeleccionado.id_usuario}/password`, {
        method: 'PUT',
        body: { password: passNueva },
      });
      agregarToast(`Contraseña de ${usuarioSeleccionado.nombre} actualizada correctamente`, 'exito');
      passNueva = '';
      passConfirmar = '';
    } catch (e) {
      errorPass = e instanceof Error ? e.message : 'Error actualizando la contraseña';
    } finally {
      cambiandoPass = false;
    }
  }

  /**
   * @param {number} id
   * @param {string} nombre
   */
  function pedirConfirmacion(id, nombre) {
    confirmandoId = id;
    confirmandoNombre = nombre;
  }

  async function confirmarEliminar() {
    eliminando = true;
    try {
      await pedirApi(`/api/usuarios/${confirmandoId}`, { method: 'DELETE' });
      confirmandoId = 0;
      modalDetalleAbierto = false;
      agregarToast('Usuario eliminado', 'info');
      await cargar();
    } catch (e) {
      error = e instanceof Error ? e.message : 'error eliminando';
    } finally {
      eliminando = false;
    }
  }

  /** @param {string} iso */
  function formatearFecha(iso) {
    if (!iso) return '-';
    return new Date(iso).toLocaleDateString('es-AR', { day: '2-digit', month: '2-digit', year: 'numeric' });
  }

  $effect(() => {
    if (esDueno()) cargar();
  });
</script>

<div class="pagina">
  <div class="cabecera">
    <div>
      <BotonVolver />
      <h2>Usuarios</h2>
      <p class="text-muted">Gestioná los empleados y cajeros que acceden al sistema. Hacé clic en un usuario para ver su detalle.</p>
    </div>
    <button class="btn btn-primary" onclick={abrirNuevo}>
      + Nuevo cajero
    </button>
  </div>

  {#if cargando}
    <p class="text-muted">Cargando…</p>
  {:else if error}
    <p class="error">{error}</p>
  {:else if usuarios.length === 0}
    <div class="vacio">
      <p class="text-muted">No hay empleados registrados todavía.</p>
      <button class="btn btn-primary" onclick={abrirNuevo}>Crear el primero</button>
    </div>
  {:else}
    <table class="table">
      <thead>
        <tr>
          <th>Empleado</th>
          <th>Usuario</th>
          <th>Email</th>
          <th>Rol</th>
          <th>Vistas Permitidas</th>
          <th>Alta</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#each usuarios as u (u.id_usuario)}
          <tr class="fila-usuario" onclick={() => abrirDetalle(u)}>
            <td>
              <div class="celda-usuario">
                <div class="avatar-mini" style="background: {obtenerColorArcoiris(u.nombre)}">
                  {obtenerIniciales(u.nombre)}
                </div>
                <strong>{u.nombre}</strong>
              </div>
            </td>
            <td>
              <span class="badge-username">@{u.usuario || u.nombre.toLowerCase().replace(/\s+/g, '')}</span>
            </td>
            <td class="text-muted">{u.email || '—'}</td>
            <td>
              <span class="tag" class:tag-accent={u.rol === 'dueño'} class:tag-neutral={u.rol === 'cajero'}>
                {u.rol}
              </span>
            </td>
            <td>
              {#if u.rol === 'dueño'}
                <span class="badge-permiso todo">Todas las vistas</span>
              {:else if Array.isArray(u.permisos) && u.permisos.length > 0}
                <div class="lista-badges-permisos">
                  {#each u.permisos as p}
                    <span class="badge-permiso">{p}</span>
                  {/each}
                </div>
              {:else}
                <span class="text-muted">Vender, Stock</span>
              {/if}
            </td>
            <td class="text-muted">{formatearFecha(u.fecha_alta)}</td>
            <td class="acciones-fila">
              {#if u.rol !== 'dueño'}
                <button
                  class="btn btn-ghost btn-sm"
                  onclick={(e) => {
                    e.stopPropagation();
                    abrirEditar(u);
                  }}
                >
                  Editar
                </button>
                <button
                  class="btn btn-ghost btn-sm btn-peligro"
                  onclick={(e) => {
                    e.stopPropagation();
                    pedirConfirmacion(u.id_usuario, u.nombre);
                  }}
                >
                  Eliminar
                </button>
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>

<!-- MODAL WIZARD DE CREACIÓN / EDICIÓN CON DIMENSIONES FIJAS -->
<Dialogo abierto={modalAbierto} ancho="560px">
  <div class="wizard-contenedor-fijo">
    <!-- CONTENIDO DEL PASO ACTIVO (ALTURA FIJA) -->
    <div class="paso-cuerpo-fijo">
      {#if pasoWizard === 1}
        <div class="paso-contenido">
          <h4>Identificación del empleado</h4>
          <p class="text-muted desc-paso">Ingresá el nombre completo del empleado y su nombre de usuario sin espacios.</p>

          <div class="campos-modal">
            <div class="field">
              <label for="w_nombre">Nombre del Empleado *</label>
              <input
                id="w_nombre"
                class="input {errorNombre ? 'input-error' : ''}"
                type="text"
                placeholder="Ej: Augusto Coria"
                bind:value={formNombre}
                oninput={() => (errorNombre = '')}
                required
              />
              <div class="slot-leyenda">
                {#if errorNombre}
                  <span class="leyenda-error">{errorNombre}</span>
                {/if}
              </div>
            </div>

            <div class="field">
              <div class="label-con-badge">
                <label for="w_usuario">Nombre de Usuario (Login) *</label>
                <span class="badge-opcional">Sin espacios ni símbolos</span>
              </div>
              <input
                id="w_usuario"
                class="input {errorUsuario ? 'input-error' : ''}"
                type="text"
                placeholder="Ej: acoria"
                bind:value={formUsuario}
                oninput={() => (errorUsuario = '')}
                required
              />
              <div class="slot-leyenda">
                {#if errorUsuario}
                  <span class="leyenda-error">{errorUsuario}</span>
                {:else}
                  <span class="leyenda-ayuda">Solo letras, números, puntos y guiones</span>
                {/if}
              </div>
            </div>

            <div class="field">
              <div class="label-con-badge">
                <label for="w_email">Email</label>
                <span class="badge-opcional">Opcional — No necesario para cajeros</span>
              </div>
              <input
                id="w_email"
                class="input {errorEmail ? 'input-error' : ''}"
                type="email"
                placeholder="ejemplo@mail.com"
                bind:value={formEmail}
                oninput={() => (errorEmail = '')}
              />
              <div class="slot-leyenda">
                {#if errorEmail}
                  <span class="leyenda-error">{errorEmail}</span>
                {/if}
              </div>
            </div>
          </div>
        </div>
      {/if}

      {#if pasoWizard === 2 && !modoEdicion}
        <div class="paso-contenido">
          <h4>Contraseña de acceso</h4>
          <p class="text-muted desc-paso">Definí la clave secreta para este cajero (mínimo 8 caracteres).</p>

          <div class="campos-modal">
            <div class="field">
              <label for="w_pass">Contraseña *</label>
              <input
                id="w_pass"
                class="input"
                type="password"
                placeholder="••••••••"
                bind:value={formPassword}
                minlength="8"
                required
              />
              <div class="slot-leyenda"></div>
            </div>

            <div class="field">
              <label for="w_pass_conf">Confirmar Contraseña *</label>
              <input
                id="w_pass_conf"
                class="input"
                type="password"
                placeholder="••••••••"
                bind:value={formConfirmPassword}
                minlength="8"
                required
              />
              <div class="slot-leyenda"></div>
            </div>
          </div>
        </div>
      {/if}

      {#if pasoWizard === 3}
        <div class="paso-contenido">
          <div class="cabecera-permisos">
            <div>
              <h4>Vistas permitidas</h4>
              <p class="text-muted desc-paso">Marcá las pantallas a las que tendrá acceso este usuario.</p>
            </div>
            <div class="acciones-rapidas-permisos">
              <button type="button" class="btn-link-sm" onclick={seleccionarBasicosPermisos}>Solo básicos</button>
              <button type="button" class="btn-link-sm" onclick={seleccionarTodosPermisos}>Todas</button>
            </div>
          </div>

          <div class="grilla-permisos">
            {#each MODULOS_DISPONIBLES as mod (mod.id)}
              <label class="tarjeta-permiso {formPermisos.includes(mod.id) ? 'marcada' : ''}">
                <input
                  type="checkbox"
                  checked={formPermisos.includes(mod.id)}
                  onchange={() => togglePermiso(mod.id)}
                />
                <div class="permiso-info">
                  <span class="permiso-nombre">{mod.nombre}</span>
                  <span class="permiso-desc">{mod.desc}</span>
                </div>
              </label>
            {/each}
          </div>
        </div>
      {/if}

      {#if pasoWizard === 4}
        <div class="paso-contenido">
          <h4>Confirmación de la ficha</h4>
          <p class="text-muted desc-paso">Revisá los datos antes de registrar al usuario en tu comercio.</p>

          <div class="ficha-resumen">
            <div class="ficha-fila">
              <span class="ficha-etiqueta">Negocio vinculado:</span>
              <span class="ficha-valor destacado">{sesion.negocio || 'Tu Comercio'}</span>
            </div>
            <div class="ficha-fila">
              <span class="ficha-etiqueta">Empleado:</span>
              <span class="ficha-valor">{formNombre}</span>
            </div>
            <div class="ficha-fila">
              <span class="ficha-etiqueta">Nombre de Usuario:</span>
              <span class="ficha-valor badge-username">@{formUsuario}</span>
            </div>
            <div class="ficha-fila">
              <span class="ficha-etiqueta">Email:</span>
              <span class="ficha-valor">{formEmail || 'Sin email'}</span>
            </div>
            <div class="ficha-fila">
              <span class="ficha-etiqueta">Rol:</span>
              <span class="tag tag-neutral">Cajero / Empleado</span>
            </div>
            <div class="ficha-bloque">
              <span class="ficha-etiqueta">Vistas autorizadas ({formPermisos.length}):</span>
              <div class="lista-badges-permisos grande">
                {#each formPermisos as p}
                  <span class="badge-permiso">{p}</span>
                {/each}
              </div>
            </div>
          </div>
        </div>
      {/if}
    </div>

    <!-- PIE DEL WIZARD: BARRA DE PUNTOS DE PROGRESO DEBAJO Y BOTONES -->
    <div class="wizard-pie">
      <div class="barra-pasos-puntos">
        <div class="paso-punto {pasoWizard === 1 ? 'actual' : pasoWizard > 1 ? 'completado' : ''}">
          <span class="num-punto">1</span>
          <span class="lbl-punto">Datos</span>
        </div>
        <div class="conector-puntos {pasoWizard > 1 ? 'activo' : ''}">
          <span class="dot">•</span><span class="dot">•</span><span class="dot">•</span>
        </div>

        {#if !modoEdicion}
          <div class="paso-punto {pasoWizard === 2 ? 'actual' : pasoWizard > 2 ? 'completado' : ''}">
            <span class="num-punto">2</span>
            <span class="lbl-punto">Clave</span>
          </div>
          <div class="conector-puntos {pasoWizard > 2 ? 'activo' : ''}">
            <span class="dot">•</span><span class="dot">•</span><span class="dot">•</span>
          </div>
        {/if}

        <div class="paso-punto {pasoWizard === 3 ? 'actual' : pasoWizard > 3 ? 'completado' : ''}">
          <span class="num-punto">{modoEdicion ? 2 : 3}</span>
          <span class="lbl-punto">Vistas</span>
        </div>
        <div class="conector-puntos {pasoWizard > 3 ? 'activo' : ''}">
          <span class="dot">•</span><span class="dot">•</span><span class="dot">•</span>
        </div>

        <div class="paso-punto {pasoWizard === 4 ? 'actual' : pasoWizard > 4 ? 'completado' : ''}">
          <span class="num-punto">{modoEdicion ? 3 : 4}</span>
          <span class="lbl-punto">Ficha</span>
        </div>
      </div>

      {#if errorModal}
        <p class="error error-wizard">{errorModal}</p>
      {/if}

      <div class="dialog-actions wizard-acciones">
        <button type="button" class="btn btn-ghost btn-cancelar-izq" onclick={() => (modalAbierto = false)}>Cancelar</button>

        <div class="wizard-acciones-derecha">
          {#if pasoWizard > 1}
            <button type="button" class="btn btn-secondary" onclick={anteriorPaso}>Anterior</button>
          {/if}

          {#if pasoWizard < 4}
            <button type="button" class="btn btn-primary" onclick={siguientePaso}>
              Siguiente
            </button>
          {:else}
            <button type="button" class="btn btn-primary" disabled={guardando} onclick={guardar}>
              {guardando ? 'Guardando…' : modoEdicion ? 'Guardar Cambios' : 'Crear Usuario'}
            </button>
          {/if}
        </div>
      </div>
    </div>
  </div>
</Dialogo>

<!-- MODAL DE DETALLE DE USUARIO Y CAMBIO DE CONTRASEÑA -->
<Dialogo abierto={modalDetalleAbierto} ancho="520px">
  {#if usuarioSeleccionado}
    <div class="cabecera-detalle">
      <div class="avatar-gran" style="background: {obtenerColorArcoiris(usuarioSeleccionado.nombre)}">
        {obtenerIniciales(usuarioSeleccionado.nombre)}
      </div>
      <div class="info-cabecera-usuario">
        <h3>{usuarioSeleccionado.nombre}</h3>
        <span class="badge-username">@{usuarioSeleccionado.usuario || usuarioSeleccionado.nombre.toLowerCase().replace(/\s+/g, '')}</span>
        <span class="tag" class:tag-accent={usuarioSeleccionado.rol === 'dueño'} class:tag-neutral={usuarioSeleccionado.rol === 'cajero'}>
          Rol: {usuarioSeleccionado.rol}
        </span>
      </div>
    </div>

    <div class="seccion-detalle">
      <div class="bloque-info">
        <div class="campo-det">
          <span class="lbl">ID Usuario</span>
          <span class="val">#{usuarioSeleccionado.id_usuario}</span>
        </div>
        <div class="campo-det">
          <span class="lbl">Email</span>
          <span class="val">{usuarioSeleccionado.email || 'Sin email'}</span>
        </div>
        <div class="campo-det">
          <span class="lbl">Comercio</span>
          <span class="val">{sesion.negocio || 'Negocio'}</span>
        </div>
        <div class="campo-det">
          <span class="lbl">Fecha de Alta</span>
          <span class="val">{formatearFecha(usuarioSeleccionado.fecha_alta)}</span>
        </div>
      </div>

      <!-- VISTAS Y PERMISOS -->
      <div class="bloque-permisos-det">
        <div class="encabezado-sub">
          <h5>Vistas Autorizadas ({usuarioSeleccionado.rol === 'dueño' ? 'Todas' : (usuarioSeleccionado.permisos?.length || 2)})</h5>
          {#if usuarioSeleccionado.rol !== 'dueño'}
            <button class="btn-link-sm" onclick={() => abrirEditarDesdeDetalle(usuarioSeleccionado)}>
              Editar permisos
            </button>
          {/if}
        </div>
        <div class="lista-permisos-tarjetas">
          {#if usuarioSeleccionado.rol === 'dueño'}
            <div class="tarjeta-permiso-mini toda">
              <strong>Acceso Completo</strong> — El dueño puede ingresar a todos los módulos.
            </div>
          {:else if Array.isArray(usuarioSeleccionado.permisos) && usuarioSeleccionado.permisos.length > 0}
            {#each usuarioSeleccionado.permisos as p}
              {@const mod = MODULOS_DISPONIBLES.find((m) => m.id === p)}
              <div class="tarjeta-permiso-mini">
                <strong>{mod ? mod.nombre : p}</strong>
                <span>{mod ? mod.desc : ''}</span>
              </div>
            {/each}
          {:else}
            <div class="tarjeta-permiso-mini">
              <strong>Vender y Stock</strong> — Permisos básicos por defecto.
            </div>
          {/if}
        </div>
      </div>

      <!-- CAMBIO DE CONTRASEÑA -->
      {#if usuarioSeleccionado.rol !== 'dueño'}
        <div class="bloque-pass">
          <h5>Cambiar Contraseña</h5>
          <form onsubmit={(e) => { e.preventDefault(); actualizarContrasena(); }}>
            <div class="grilla-pass">
              <div class="field">
                <label for="p_nueva">Nueva Contraseña</label>
                <input id="p_nueva" class="input" type="password" placeholder="••••••••" bind:value={passNueva} minlength="8" required />
              </div>
              <div class="field">
                <label for="p_conf">Confirmar Contraseña</label>
                <input id="p_conf" class="input" type="password" placeholder="••••••••" bind:value={passConfirmar} minlength="8" required />
              </div>
            </div>
            {#if errorPass}
              <p class="error error-pass">{errorPass}</p>
            {/if}
            <button type="submit" class="btn btn-secondary btn-block-sm" disabled={cambiandoPass || !passNueva}>
              {cambiandoPass ? 'Actualizando…' : 'Actualizar Contraseña'}
            </button>
          </form>
        </div>
      {/if}
    </div>

    <div class="dialog-actions">
      <button class="btn btn-ghost" onclick={() => (modalDetalleAbierto = false)}>Cerrar</button>
      {#if usuarioSeleccionado.rol !== 'dueño'}
        <button class="btn btn-secondary" onclick={() => abrirEditarDesdeDetalle(usuarioSeleccionado)}>
          Editar Usuario
        </button>
        <button class="btn btn-peligro-fill" onclick={() => pedirConfirmacion(usuarioSeleccionado.id_usuario, usuarioSeleccionado.nombre)}>
          Eliminar
        </button>
      {/if}
    </div>
  {/if}
</Dialogo>

<!-- Confirmación de eliminar -->
<Dialogo abierto={confirmandoId > 0} ancho="360px">
  <span class="dialog-title">Eliminar usuario</span>
  <p class="dialog-body">
    ¿Seguro que querés eliminar a <strong>{confirmandoNombre}</strong>? Esta acción no se puede deshacer.
  </p>
  <div class="dialog-actions">
    <button class="btn btn-secondary" onclick={() => (confirmandoId = 0)}>Cancelar</button>
    <button class="btn btn-peligro-fill" disabled={eliminando} onclick={confirmarEliminar}>
      {eliminando ? 'Eliminando…' : 'Eliminar'}
    </button>
  </div>
</Dialogo>

<style>
  .pagina { max-width: 960px; }
  .cabecera {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    margin-bottom: var(--space-5);
  }
  .cabecera h2 { margin: var(--space-2) 0 4px; }
  .cabecera p { margin: 0; font-size: 13px; }
  .error { color: var(--color-peligro, #d32f2f); font-size: 13px; }
  .error-wizard { margin-top: 4px; text-align: center; }
  .error-pass { margin-top: 6px; font-size: 12px; }

  .vacio {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-6) 0;
  }

  .fila-usuario {
    cursor: pointer;
    transition: background 0.12s ease;
  }
  .fila-usuario:hover {
    background: color-mix(in srgb, var(--color-accent) 6%, transparent);
  }

  .celda-usuario {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .avatar-mini {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    color: #fff;
    font-weight: 700;
    font-size: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 2px 4px rgba(0,0,0,0.12);
  }

  .badge-username {
    font-family: monospace;
    font-size: 12px;
    color: var(--color-accent-700);
    background: var(--color-accent-100);
    padding: 2px 8px;
    border-radius: 6px;
    font-weight: 600;
  }

  .acciones-fila {
    display: flex;
    gap: var(--space-1);
    justify-content: flex-end;
  }
  .btn-sm { font-size: 12px; min-height: 30px; padding: 4px 10px; }
  .btn-peligro { color: var(--color-peligro); }
  .btn-peligro:hover { background: color-mix(in srgb, var(--color-peligro) 10%, transparent); }
  .btn-peligro-fill {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    min-height: 38px;
    padding: 8px 16px;
    border: 1px solid var(--color-peligro);
    border-radius: var(--radius-md);
    background: var(--color-peligro);
    font-family: var(--font-body);
    font-size: 14px;
    color: #fff;
    cursor: pointer;
    transition: background 0.12s ease;
  }
  .btn-peligro-fill:hover { background: color-mix(in srgb, var(--color-peligro) 85%, #000); }
  .btn-peligro-fill:disabled { opacity: 0.45; pointer-events: none; }

  /* WIZARD DE DIMENSIÓN FIJA */
  .wizard-contenedor-fijo {
    display: flex;
    flex-direction: column;
    height: 480px;
    box-sizing: border-box;
  }

  .paso-cuerpo-fijo {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow-y: auto;
    padding-right: 4px;
    margin-top: 4px;
  }

  .paso-contenido {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .paso-contenido h4 { margin: 0; font-size: 18px; color: var(--color-text); }
  .desc-paso { font-size: 13px; margin: 0 0 var(--space-2); }

  .campos-modal {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .label-con-badge {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .badge-opcional {
    font-size: 11px;
    color: color-mix(in srgb, var(--color-text) 50%, transparent);
    font-style: italic;
  }

  /* ESTILOS DE ERROR EN CAMPOS DE ENTRADA */
  .input-error {
    border-color: var(--color-peligro, #d32f2f) !important;
    background-color: color-mix(in srgb, var(--color-peligro, #d32f2f) 4%, #ffffff) !important;
    box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-peligro, #d32f2f) 25%, transparent) !important;
  }

  .slot-leyenda {
    min-height: 18px;
    margin-top: 2px;
    display: flex;
    align-items: center;
  }

  .leyenda-error {
    font-size: 11.5px;
    color: var(--color-peligro, #d32f2f);
    font-weight: 600;
  }

  .leyenda-ayuda {
    font-size: 11px;
    color: color-mix(in srgb, var(--color-text) 50%, transparent);
  }

  /* PERMISOS GRID */
  .cabecera-permisos {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
  }
  .acciones-rapidas-permisos {
    display: flex;
    gap: 10px;
  }
  .btn-link-sm {
    border: none;
    background: transparent;
    color: var(--color-accent-700);
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
    padding: 0;
  }
  .btn-link-sm:hover { text-decoration: underline; }

  .grilla-permisos {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    gap: 8px;
    max-height: 310px;
    overflow-y: auto;
    padding-right: 4px;
  }
  .tarjeta-permiso {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    padding: 10px 12px;
    border: 1px solid var(--color-divider);
    border-radius: var(--radius-md);
    background: var(--color-bg, #f2f4f4);
    cursor: pointer;
    transition: all 0.15s ease;
  }
  .tarjeta-permiso:hover {
    border-color: var(--color-accent-400);
  }
  .tarjeta-permiso.marcada {
    border-color: var(--color-accent);
    background: var(--color-accent-100);
  }
  .tarjeta-permiso input { margin-top: 3px; cursor: pointer; }
  .permiso-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .permiso-nombre { font-size: 13px; font-weight: 600; }
  .permiso-desc { font-size: 11px; color: color-mix(in srgb, var(--color-text) 60%, transparent); line-height: 1.2; }

  /* FICHA RESUMEN */
  .ficha-resumen {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    padding: var(--space-4);
    background: var(--color-bg, #f2f4f4);
    border-radius: var(--radius-md);
    border: 1px solid var(--color-divider);
  }
  .ficha-fila {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 13.5px;
  }
  .ficha-etiqueta { font-weight: 600; color: color-mix(in srgb, var(--color-text) 65%, transparent); }
  .ficha-valor { font-weight: 600; }
  .ficha-valor.destacado { color: var(--color-accent-700); font-size: 15px; }

  .ficha-bloque {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin-top: 4px;
  }
  .lista-badges-permisos {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }
  .lista-badges-permisos.grande { gap: 6px; }
  .badge-permiso {
    font-size: 11px;
    padding: 2px 8px;
    border-radius: 6px;
    background: color-mix(in srgb, var(--color-text) 8%, transparent);
    color: var(--color-text);
  }
  .badge-permiso.todo {
    background: var(--color-accent-100);
    color: var(--color-accent-700);
    font-weight: 600;
  }

  /* WIZARD PIE Y BARRA DE PUNTOS */
  .wizard-pie {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding-top: 10px;
    border-top: 1px solid var(--color-divider);
  }

  .barra-pasos-puntos {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    margin-bottom: 2px;
  }

  .paso-punto {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
  }

  .num-punto {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    background: var(--color-bg-subtle, #e5e7e9);
    color: #888888;
    border: 1.5px solid #cccccc;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 13px;
    font-weight: 600;
    transition: all 0.2s ease;
  }

  .paso-punto.actual .num-punto {
    background: var(--color-accent, #2e6e73);
    color: #ffffff;
    border-color: var(--color-accent, #2e6e73);
    box-shadow: 0 0 0 3px var(--color-accent-100, #e0f2f1);
    transform: scale(1.08);
    font-weight: 700;
  }

  .paso-punto.completado .num-punto {
    background: var(--color-accent-100, #e0f2f1);
    color: var(--color-accent, #2e6e73);
    border-color: var(--color-accent-400, #2e6e73);
    font-weight: 700;
  }

  .lbl-punto {
    font-size: 11px;
    color: #888888;
    font-weight: 500;
  }
  .paso-punto.actual .lbl-punto {
    color: var(--color-accent, #2e6e73);
    font-weight: 700;
  }
  .paso-punto.completado .lbl-punto {
    color: var(--color-accent-700, #2e6e73);
  }

  .conector-puntos {
    display: flex;
    align-items: center;
    gap: 3px;
    color: #cccccc;
    font-size: 11px;
    margin-bottom: 14px;
  }

  .conector-puntos.activo {
    color: var(--color-accent, #2e6e73);
    font-weight: bold;
  }

  .wizard-acciones {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 0;
    width: 100%;
  }

  .btn-cancelar-izq {
    color: var(--color-text-muted, #666);
  }

  .wizard-acciones-derecha {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  /* DETALLE MODAL STYLES */
  .cabecera-detalle {
    display: flex;
    align-items: center;
    gap: 16px;
    margin-bottom: var(--space-4);
    padding-bottom: var(--space-3);
    border-bottom: 1px solid var(--color-divider);
  }
  .avatar-gran {
    width: 52px;
    height: 52px;
    border-radius: 50%;
    color: #fff;
    font-size: 20px;
    font-weight: 800;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 4px 8px rgba(0,0,0,0.15);
  }
  .info-cabecera-usuario h3 { margin: 0 0 4px; font-size: 20px; }

  .seccion-detalle {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }
  .bloque-info {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
    background: var(--color-bg, #f2f4f4);
    padding: 12px 14px;
    border-radius: var(--radius-md);
  }
  .campo-det {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .campo-det .lbl { font-size: 11px; color: color-mix(in srgb, var(--color-text) 60%, transparent); font-weight: 600; }
  .campo-det .val { font-size: 13.5px; font-weight: 600; }

  .bloque-permisos-det, .bloque-pass {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .encabezado-sub {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .encabezado-sub h5, .bloque-pass h5 { margin: 0; font-size: 14px; font-weight: 700; }

  .lista-permisos-tarjetas {
    display: flex;
    flex-direction: column;
    gap: 6px;
    max-height: 150px;
    overflow-y: auto;
  }
  .tarjeta-permiso-mini {
    display: flex;
    flex-direction: column;
    padding: 8px 10px;
    border-radius: var(--radius-sm);
    background: var(--color-bg, #f8f9f8);
    border: 1px solid var(--color-divider);
    font-size: 12px;
  }
  .tarjeta-permiso-mini.toda {
    background: var(--color-accent-100);
    border-color: var(--color-accent-400);
  }
  .tarjeta-permiso-mini strong { font-size: 13px; color: var(--color-accent-700); }
  .tarjeta-permiso-mini span { font-size: 11px; color: color-mix(in srgb, var(--color-text) 60%, transparent); }

  .grilla-pass {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
    margin-bottom: 8px;
  }
  .btn-block-sm {
    width: 100%;
    min-height: 36px;
    font-size: 13px;
    margin-top: 4px;
  }
</style>
