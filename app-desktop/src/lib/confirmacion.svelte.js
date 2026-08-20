export const estadoConfirmacion = $state({
  cerrarSesion: false,
  cerrarApp: false,
});

export const permitiendoCierre = $state({
  activo: false,
});

export function solicitarCerrarSesion() {
  estadoConfirmacion.cerrarSesion = true;
}

export function cancelarCerrarSesion() {
  estadoConfirmacion.cerrarSesion = false;
}

export function solicitarCerrarApp() {
  estadoConfirmacion.cerrarApp = true;
}

export function cancelarCerrarApp() {
  estadoConfirmacion.cerrarApp = false;
  permitiendoCierre.activo = false;
}

export async function confirmarCerrarApp() {
  estadoConfirmacion.cerrarApp = false;
  permitiendoCierre.activo = true;
  if (typeof window !== 'undefined' && '__TAURI_INTERNALS__' in window) {
    try {
      const { getCurrentWindow } = await import('@tauri-apps/api/window');
      const appWindow = getCurrentWindow();
      await appWindow.destroy();
    } catch (e) {
      console.error('Error destruyendo ventana Tauri:', e);
      try {
        const { getCurrentWindow } = await import('@tauri-apps/api/window');
        await getCurrentWindow().close();
      } catch {
        window.close();
      }
    }
  } else {
    window.close();
  }
}
