// Formatos compartidos de la app (pesos argentinos, sin decimales en pantalla).
/** @param {number} n */
export function fmtPesos(n) {
  return '$ ' + Math.round(n).toLocaleString('es-AR');
}
