import type { Response } from 'express';

const clientes = new Set<Response>();

// Heartbeat cada 30 segundos para mantener vivas las conexiones a través de túneles y proxies
let intervaloHeartbeat: NodeJS.Timeout | null = null;

function asegurarHeartbeat() {
  if (intervaloHeartbeat || clientes.size === 0) return;

  intervaloHeartbeat = setInterval(() => {
    if (clientes.size === 0 && intervaloHeartbeat) {
      clearInterval(intervaloHeartbeat);
      intervaloHeartbeat = null;
      return;
    }

    for (const res of clientes) {
      try {
        res.write(': keepalive\n\n');
      } catch {
        clientes.delete(res);
      }
    }
  }, 30_000);
}

/**
 * Registra un nuevo cliente para streaming de eventos del servidor (SSE).
 */
export function registrarClienteSSE(res: Response) {
  res.writeHead(200, {
    'Content-Type': 'text/event-stream',
    'Cache-Control': 'no-cache, no-transform',
    'Connection': 'keep-alive',
    'X-Accel-Buffering': 'no',
    'Access-Control-Allow-Origin': '*',
  });

  res.write(': conexion establecida con canal de actualizaciones\n\n');
  clientes.add(res);
  asegurarHeartbeat();

  console.log(`[sse] cliente desktop conectado (total activos: ${clientes.size})`);

  res.on('close', () => {
    clientes.delete(res);
    console.log(`[sse] cliente desktop desconectado (total activos: ${clientes.size})`);
    if (clientes.size === 0 && intervaloHeartbeat) {
      clearInterval(intervaloHeartbeat);
      intervaloHeartbeat = null;
    }
  });
}

/**
 * Emite un evento a todas las terminales de escritorio activas
 * notificando que hay una actualización disponible o modificada.
 */
export function broadcastActualizacion(datos?: Record<string, any>) {
  if (clientes.size === 0) return;

  const payload = JSON.stringify(datos ?? {
    tipo: 'actualizacion_disponible',
    timestamp: Date.now(),
  });

  const mensaje = `event: actualizacion\ndata: ${payload}\n\n`;

  console.log(`[sse] emitiendo broadcast de actualizacion a ${clientes.size} terminales`);

  for (const res of clientes) {
    try {
      res.write(mensaje);
    } catch {
      clientes.delete(res);
    }
  }
}

/**
 * Retorna la cantidad de terminales de escritorio conectadas actualmente.
 */
export function contarClientesConectados(): number {
  return clientes.size;
}
