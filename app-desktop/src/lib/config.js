// Base de la API. En desarrollo (.env.development) apunta directo al gateway Express
// en :3000; en producción se usa el sidecar de Tailscale en :9090.
export const BASE_API = import.meta.env.VITE_API_URL ?? 'http://localhost:3000';
