export function renderMobileHtml(): string {
  return `<!DOCTYPE html>
<html lang="es">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no" />
  <title>StockAPP · Control en Vivo</title>
  <meta name="theme-color" content="#2e6e73" />
  <meta name="apple-mobile-web-app-capable" content="yes" />
  <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent" />
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
  <style>
    :root {
      --primary: #2e6e73;
      --primary-dark: #1f4e52;
      --primary-light: #44898f;
      --primary-subtle: #eaf4f5;
      --accent: #2e6e73;
      --bg: #f4f6f8;
      --surface: #ffffff;
      --text: #1e293b;
      --text-muted: #64748b;
      --border: #e2e8f0;
      --success: #16a34a;
      --warning: #d97706;
      --danger: #dc2626;
      --radius: 14px;
      --radius-sm: 8px;
      --shadow: 0 4px 12px -2px rgba(0, 0, 0, 0.05), 0 2px 6px -1px rgba(0, 0, 0, 0.03);
    }

    * {
      box-sizing: border-box;
      margin: 0;
      padding: 0;
      -webkit-tap-highlight-color: transparent;
    }

    body {
      font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      background-color: var(--bg);
      color: var(--text);
      min-height: 100vh;
      padding-bottom: 40px;
    }

    /* HEADER */
    .header {
      background: linear-gradient(135deg, var(--primary-dark) 0%, var(--primary) 100%);
      color: #ffffff;
      padding: 16px 20px 20px;
      position: sticky;
      top: 0;
      z-index: 50;
      box-shadow: 0 4px 16px rgba(31, 78, 82, 0.2);
    }

    .header-top {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 8px;
    }

    .badge-live {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      background: rgba(255, 255, 255, 0.15);
      backdrop-filter: blur(8px);
      padding: 4px 10px;
      border-radius: 20px;
      font-size: 11.5px;
      font-weight: 600;
      letter-spacing: 0.4px;
      text-transform: uppercase;
    }

    .live-dot {
      width: 7px;
      height: 7px;
      background: #22c55e;
      border-radius: 50%;
      box-shadow: 0 0 0 0 rgba(34, 197, 94, 0.7);
      animation: pulse-green 1.8s infinite;
    }

    @keyframes pulse-green {
      0% { box-shadow: 0 0 0 0 rgba(34, 197, 94, 0.7); }
      70% { box-shadow: 0 0 0 8px rgba(34, 197, 94, 0); }
      100% { box-shadow: 0 0 0 0 rgba(34, 197, 94, 0); }
    }

    .header-actions {
      display: flex;
      gap: 8px;
    }

    .btn-icon {
      background: rgba(255, 255, 255, 0.12);
      border: 1px solid rgba(255, 255, 255, 0.2);
      color: #ffffff;
      width: 36px;
      height: 36px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      cursor: pointer;
      transition: background 0.15s ease;
    }

    .btn-icon:active {
      background: rgba(255, 255, 255, 0.25);
    }

    .rotar {
      animation: giro 0.8s linear;
    }

    @keyframes giro {
      from { transform: rotate(0deg); }
      to { transform: rotate(360deg); }
    }

    .header-negocio {
      font-size: 20px;
      font-weight: 700;
      line-height: 1.2;
    }

    .header-sub {
      font-size: 12.5px;
      opacity: 0.85;
      margin-top: 2px;
    }

    /* CONTENEDOR PRINCIPAL */
    .container {
      max-width: 500px;
      margin: 0 auto;
      padding: 16px;
      display: flex;
      flex-direction: column;
      gap: 14px;
    }

    /* TARJETAS */
    .card {
      background: var(--surface);
      border-radius: var(--radius);
      border: 1px solid var(--border);
      padding: 16px;
      box-shadow: var(--shadow);
    }

    /* CARD VENTAS HOY */
    .card-destacada {
      background: linear-gradient(135deg, #ffffff 0%, #f9fbfb 100%);
      border-left: 4px solid var(--primary);
    }

    .total-label {
      font-size: 12.5px;
      font-weight: 600;
      color: var(--text-muted);
      text-transform: uppercase;
      letter-spacing: 0.5px;
      margin-bottom: 4px;
    }

    .total-monto {
      font-size: 32px;
      font-weight: 800;
      color: var(--primary-dark);
      line-height: 1.1;
      margin-bottom: 10px;
    }

    .grid-metricas {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 8px;
      padding-top: 10px;
      border-top: 1px solid var(--border);
    }

    .metrica-item span {
      font-size: 11.5px;
      color: var(--text-muted);
      display: block;
    }

    .metrica-item strong {
      font-size: 16px;
      font-weight: 700;
      color: var(--text);
    }

    /* CARD CAJERO EN TURNO */
    .card-cajero {
      display: flex;
      align-items: center;
      gap: 12px;
      background: #ffffff;
    }

    .cajero-avatar {
      width: 44px;
      height: 44px;
      border-radius: 50%;
      background: var(--primary-subtle);
      border: 1.5px solid var(--primary-light);
      display: flex;
      align-items: center;
      justify-content: center;
      color: var(--primary-dark);
      font-weight: 700;
      font-size: 16px;
      flex-shrink: 0;
    }

    .cajero-info h4 {
      font-size: 14px;
      font-weight: 700;
      color: var(--text);
    }

    .cajero-info p {
      font-size: 12px;
      color: var(--text-muted);
      margin-top: 1px;
    }

    .cajero-estado {
      margin-left: auto;
      text-align: right;
    }

    .badge-cajero {
      display: inline-block;
      font-size: 10.5px;
      font-weight: 600;
      padding: 3px 8px;
      border-radius: 12px;
      background: #dcfce7;
      color: #15803d;
    }

    /* MEDIOS DE PAGO */
    .seccion-tit {
      font-size: 13px;
      font-weight: 700;
      color: var(--text-muted);
      text-transform: uppercase;
      letter-spacing: 0.5px;
      margin-bottom: 12px;
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .fila-pago {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 8px;
      font-size: 13px;
    }

    .pago-nombre {
      display: flex;
      align-items: center;
      gap: 6px;
      font-weight: 500;
      text-transform: capitalize;
    }

    .pago-monto {
      font-weight: 700;
    }

    .barra-progreso-bg {
      width: 100%;
      height: 6px;
      background: var(--border);
      border-radius: 4px;
      overflow: hidden;
      margin-bottom: 12px;
    }

    .barra-progreso-fill {
      height: 100%;
      background: var(--primary);
      border-radius: 4px;
    }

    /* FEED DE ÚLTIMAS VENTAS */
    .lista-ventas {
      display: flex;
      flex-direction: column;
      divide-y: 1px solid var(--border);
    }

    .item-venta {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 10px 0;
      border-bottom: 1px solid var(--border);
    }

    .item-venta:last-child {
      border-bottom: none;
      padding-bottom: 0;
    }

    .item-venta:first-child {
      padding-top: 0;
    }

    .venta-info {
      display: flex;
      flex-direction: column;
      gap: 2px;
    }

    .venta-id {
      font-size: 13.5px;
      font-weight: 600;
      color: var(--text);
    }

    .venta-meta {
      font-size: 11.5px;
      color: var(--text-muted);
    }

    .venta-monto {
      font-size: 14.5px;
      font-weight: 700;
      color: var(--primary-dark);
      text-align: right;
    }

    .tag-metodo {
      display: inline-block;
      font-size: 10px;
      font-weight: 600;
      padding: 2px 6px;
      border-radius: 4px;
      background: var(--primary-subtle);
      color: var(--primary-dark);
      text-transform: uppercase;
    }

    /* ALERTAS DE STOCK */
    .alerta-item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 9px 0;
      border-bottom: 1px solid var(--border);
      font-size: 12.5px;
    }

    .alerta-item:last-child { border-bottom: none; padding-bottom: 0; }
    .alerta-item:first-child { padding-top: 0; }

    .alerta-desc {
      font-weight: 500;
      color: var(--text);
      max-width: 65%;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .alerta-stock {
      font-weight: 700;
      color: var(--danger);
      background: #fee2e2;
      padding: 2px 8px;
      border-radius: 10px;
      font-size: 11px;
    }

    /* EQUIPOS CONECTADOS */
    .pill-equipos {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 6px;
      font-size: 12px;
      color: var(--text-muted);
      padding: 8px 12px;
      background: #ffffff;
      border: 1px dashed var(--border);
      border-radius: var(--radius-sm);
    }

    /* FOOTER */
    .footer-seguridad {
      text-align: center;
      font-size: 11.5px;
      color: var(--text-muted);
      margin-top: 10px;
      display: flex;
      flex-direction: column;
      gap: 8px;
      align-items: center;
    }

    .btn-salir {
      background: none;
      border: none;
      color: var(--danger);
      font-size: 12px;
      font-weight: 600;
      cursor: pointer;
      padding: 6px 12px;
    }

    /* CARGANDO & ERROR */
    .cargando-pantalla {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      min-height: 60vh;
      gap: 12px;
      text-align: center;
      padding: 20px;
    }

    .spinner {
      width: 36px;
      height: 36px;
      border: 3px solid rgba(46, 110, 115, 0.2);
      border-top-color: var(--primary);
      border-radius: 50%;
      animation: giro 0.8s linear infinite;
    }

    .alerta-banner {
      background: #fef2f2;
      border: 1px solid #fecaca;
      color: #b91c1c;
      padding: 12px;
      border-radius: var(--radius-sm);
      font-size: 13px;
      text-align: center;
    }
  </style>
</head>
<body>
  <div id="app">
    <div class="cargando-pantalla">
      <div class="spinner"></div>
      <p style="font-size: 14px; color: var(--text-muted);">Conectando con StockAPP...</p>
    </div>
  </div>

  <script>
    const STORAGE_TOKEN = 'stockapp_mobile_token';
    const STORAGE_DEVICE = 'stockapp_mobile_device_id';
    let autoRefreshTimer = null;

    function getDeviceId() {
      let id = localStorage.getItem(STORAGE_DEVICE);
      if (!id) {
        id = 'dev-mob-' + Math.random().toString(36).substring(2, 9);
        localStorage.setItem(STORAGE_DEVICE, id);
      }
      return id;
    }

    function detectarNombreCelular() {
      const ua = navigator.userAgent || '';
      if (/iPhone/.test(ua)) return 'iPhone (Dueño)';
      if (/iPad/.test(ua)) return 'iPad (Dueño)';
      if (/Android/.test(ua)) return 'Celular Android (Dueño)';
      return 'Celular Dueño';
    }

    function formatPesos(monto) {
      return new Intl.NumberFormat('es-AR', {
        style: 'currency',
        currency: 'ARS',
        maximumFractionDigits: 0,
      }).format(monto || 0);
    }

    function formatHora(fechaIso) {
      if (!fechaIso) return '';
      const d = new Date(fechaIso);
      return d.toLocaleTimeString('es-AR', { hour: '2-digit', minute: '2-digit' });
    }

    function calcularHaceCuanto(fechaIso) {
      if (!fechaIso) return '';
      const diffMin = Math.round((Date.now() - new Date(fechaIso).getTime()) / 60000);
      if (diffMin < 1) return 'Hace instantes';
      if (diffMin === 1) return 'Hace 1 minuto';
      if (diffMin < 60) return \`Hace \${diffMin} min\`;
      const horas = Math.floor(diffMin / 60);
      return \`Hace \${horas}h\`;
    }

    async function inicializar() {
      const params = new URLSearchParams(window.location.search);
      const pairingToken = params.get('token');

      // Si viene un token en la URL (escaneo de QR nuevo), canjearlo
      if (pairingToken) {
        try {
          const res = await fetch('/api/movil/canjear', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              token: pairingToken,
              device_id: getDeviceId(),
              nombre_dispositivo: detectarNombreCelular(),
              tipo_dispositivo: 'mobile',
            }),
          });
          const datos = await res.json();
          if (!res.ok) {
            mostrarError(datos.error || 'No se pudo autorizar el acceso móvil.');
            return;
          }
          localStorage.setItem(STORAGE_TOKEN, datos.access_token);
          // Limpiar el parámetro de la URL sin recargar
          window.history.replaceState({}, document.title, window.location.pathname);
        } catch (e) {
          mostrarError('Error de conexión al canjear el código QR.');
          return;
        }
      }

      // Cargar datos del resumen
      await cargarResumen();

      // Configurar auto-refresco cada 15 segundos
      if (autoRefreshTimer) clearInterval(autoRefreshTimer);
      autoRefreshTimer = setInterval(cargarResumen, 15000);
    }

    async function cargarResumen() {
      const token = localStorage.getItem(STORAGE_TOKEN);
      if (!token) {
        mostrarError('No hay una sesión activa. Escaneá el código QR desde la computadora del negocio.');
        return;
      }

      const btnRefresh = document.getElementById('btn-refresh');
      if (btnRefresh) btnRefresh.classList.add('rotar');

      try {
        const res = await fetch('/api/movil/resumen', {
          headers: { Authorization: 'Bearer ' + token },
        });

        if (res.status === 401 || res.status === 403) {
          localStorage.removeItem(STORAGE_TOKEN);
          mostrarError('La sesión ha expirado o el celular fue desvinculado. Volvé a escanear el código QR en la computadora.');
          return;
        }

        const data = await res.json();
        renderizar(data);
      } catch (e) {
        console.warn('Error refrescando:', e);
      } finally {
        if (btnRefresh) {
          setTimeout(() => btnRefresh.classList.remove('rotar'), 800);
        }
      }
    }

    function renderizar(data) {
      const hoy = data.hoy || {};
      const cajero = data.cajero_activo;
      const pagos = data.pagos || [];
      const ventas = data.ultimas_ventas || [];
      const stock = data.stock_critico || [];
      const totalEquipos = data.total_equipos_activos || 1;

      // Calcular porcentajes de pago
      const totalPagos = pagos.reduce((acc, p) => acc + p.total, 0) || 1;

      const app = document.getElementById('app');
      app.innerHTML = \`
        <header class="header">
          <div class="header-top">
            <div class="badge-live">
              <span class="live-dot"></span>
              En Vivo
            </div>
            <div class="header-actions">
              <button id="btn-refresh" class="btn-icon" onclick="cargarResumen()" title="Actualizar">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M21 2v6h-6"></path>
                  <path d="M3 12a9 9 0 0 1 15-6.7L21 8"></path>
                  <path d="M3 22v-6h6"></path>
                  <path d="M21 12a9 9 0 0 1-15 6.7L3 16"></path>
                </svg>
              </button>
            </div>
          </div>
          <h1 class="header-negocio">StockAPP</h1>
          <p class="header-sub">Supervisión en tiempo real · Solo Lectura</p>
        </header>

        <main class="container">
          <!-- TARJETA: VENTAS DE HOY -->
          <div class="card card-destacada">
            <div class="total-label">Ventas Totales Hoy</div>
            <div class="total-monto">\${formatPesos(hoy.total)}</div>
            <div class="grid-metricas">
              <div class="metrica-item">
                <span>Operaciones</span>
                <strong>\${hoy.cantidad || 0} tickets</strong>
              </div>
              <div class="metrica-item">
                <span>Ticket Promedio</span>
                <strong>\${formatPesos(hoy.promedio)}</strong>
              </div>
            </div>
          </div>

          <!-- TARJETA: CAJERO ATENDIENDO EN TURNO -->
          <div class="card card-cajero">
            <div class="cajero-avatar">
              \${(cajero?.nombre || 'PC')[0].toUpperCase()}
            </div>
            <div class="cajero-info">
              <h4>\${cajero ? cajero.nombre : 'Mostrador Principal'}</h4>
              <p>\${cajero ? ('Última venta ' + calcularHaceCuanto(cajero.ultima_venta_hora)) : 'Sin actividad registrada hoy'}</p>
            </div>
            <div class="cajero-estado">
              <span class="badge-cajero">Atendiendo</span>
            </div>
          </div>

          <!-- TARJETA: DESGLOSE DE MEDIOS DE PAGO -->
          <div class="card">
            <div class="seccion-tit">
              <span>Medios de Pago</span>
              <span style="font-size: 11px; text-transform: none; color: var(--text-muted);">Hoy</span>
            </div>
            \${pagos.length === 0 ? '<p style="font-size: 12.5px; color: var(--text-muted);">Aún no hay cobros registrados hoy.</p>' : ''}
            \${pagos.map(p => {
              const pct = Math.round((p.total / totalPagos) * 100);
              return \`
                <div class="fila-pago">
                  <span class="pago-nombre">\${p.metodo_pago} (\${p.cantidad})</span>
                  <span class="pago-monto">\${formatPesos(p.total)} (\${pct}%)</span>
                </div>
                <div class="barra-progreso-bg">
                  <div class="barra-progreso-fill" style="width: \${pct}%"></div>
                </div>
              \`;
            }).join('')}
          </div>

          <!-- TARJETA: ALERTAS DE STOCK CRÍTICO -->
          \${stock.length > 0 ? \`
            <div class="card" style="border-left: 4px solid var(--danger);">
              <div class="seccion-tit" style="color: var(--danger);">
                <span>Reposición Urgente (\${stock.length})</span>
              </div>
              \${stock.map(s => \`
                <div class="alerta-item">
                  <span class="alerta-desc" title="\${s.descripcion}">\${s.descripcion}</span>
                  <span class="alerta-stock">Quedan \${s.cantidad_disponible} (mín \${s.stock_minimo})</span>
                </div>
              \`).join('')}
            </div>
          \` : ''}

          <!-- TARJETA: ÚLTIMAS VENTAS EMITIDAS -->
          <div class="card">
            <div class="seccion-tit">
              <span>Últimos Tickets Emitidos</span>
              <span style="font-size: 11px; text-transform: none; color: var(--text-muted);">Recientes</span>
            </div>
            \${ventas.length === 0 ? '<p style="font-size: 12.5px; color: var(--text-muted);">Sin tickets registrados aún.</p>' : ''}
            <div class="lista-ventas">
              \${ventas.map(v => \`
                <div class="item-venta">
                  <div class="venta-info">
                    <span class="venta-id">Ticket #\${v.id_venta}</span>
                    <span class="venta-meta">\${formatHora(v.fecha_hora)} · Por \${v.cajero}</span>
                  </div>
                  <div>
                    <div class="venta-monto">\${formatPesos(v.total_venta)}</div>
                    <div style="text-align: right; margin-top: 2px;">
                      <span class="tag-metodo">\${v.metodo_pago}</span>
                    </div>
                  </div>
                </div>
              \`).join('')}
            </div>
          </div>

          <!-- INDICADOR DE EQUIPOS ACTIVOS -->
          <div class="pill-equipos">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect><line x1="8" y1="21" x2="16" y2="21"></line><line x1="12" y1="17" x2="12" y2="21"></line></svg>
            <span>\${totalEquipos} de 4 equipos conectados simultáneamente</span>
          </div>

          <!-- FOOTER -->
          <div class="footer-seguridad">
            <span>🔒 Modo Supervisión Exclusivo para Dueño · Solo Lectura</span>
            <button class="btn-salir" onclick="cerrarSesion()">Cerrar sesión en este celular</button>
          </div>
        </main>
      \`;
    }

    function mostrarError(mensaje) {
      const app = document.getElementById('app');
      app.innerHTML = \`
        <div class="cargando-pantalla">
          <div class="alerta-banner">
            <strong>Acceso no disponible</strong>
            <p style="margin-top: 6px;">\${mensaje}</p>
          </div>
          <button style="margin-top: 16px; padding: 10px 18px; border-radius: 8px; background: var(--primary); color: #fff; border: none; font-weight: 600;" onclick="location.href='/movil'">Reintentar</button>
        </div>
      \`;
    }

    function cerrarSesion() {
      if (confirm('¿Cerrar sesión en este celular? Para volver a entrar deberás escanear el QR desde la computadora.')) {
        localStorage.removeItem(STORAGE_TOKEN);
        location.reload();
      }
    }

    window.addEventListener('DOMContentLoaded', inicializar);
  </script>
</body>
</html>`;
}
