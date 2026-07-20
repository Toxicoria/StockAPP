export function env(key: string, fallback: string): string {
  return process.env[key] ?? fallback;
}
