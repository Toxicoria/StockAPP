export const estadoConfirmacion = $state({
  cerrarSesion: false,
  cerrarApp: false,
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
}

export async function confirmarCerrarApp() {
  estadoConfirmacion.cerrarApp = false;
  if (typeof window !== 'undefined' && '__TAURI_INTERNALS__' in window) {
    try {
      const { getCurrentWindow } = await import('@tauri-apps/api/window');
      const appWindow = getCurrentWindow();
      await appWindow.destroy();
    } catch {
      // Si falla destroy, intentar cerrar
      window.close();
    }
  }
}
