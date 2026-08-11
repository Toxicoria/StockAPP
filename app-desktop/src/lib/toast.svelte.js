// Sistema de notificaciones tipo Toast con Runas de Svelte 5.
/** @type {Array<{ id: number, mensaje: string, tipo: string }>} */
export const toasts = $state([]);

/**
 * Muestra una notificación emergente (Toast) en la pantalla.
 * @param {string} mensaje
 * @param {'exito' | 'error' | 'info'} [tipo='exito']
 * @param {number} [duracion=3200]
 */
export function agregarToast(mensaje, tipo = 'exito', duracion = 3200) {
  const id = Date.now() + Math.random();
  const nuevo = { id, mensaje, tipo };
  toasts.push(nuevo);

  setTimeout(() => {
    removerToast(id);
  }, duracion);
}

/** @param {number} id */
export function removerToast(id) {
  const index = toasts.findIndex((t) => t.id === id);
  if (index !== -1) {
    toasts.splice(index, 1);
  }
}
