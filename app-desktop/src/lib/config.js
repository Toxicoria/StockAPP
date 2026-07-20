// Base de la API. En desarrollo (.env.development) apunta directo al backend
// en :8080; en producción se usa el sidecar de Tailscale en :9090.
export const BASE_API = import.meta.env.VITE_API_URL ?? 'http://localhost:9090';
