export function env(key: string, fallback: string): string {
  return process.env[key] ?? fallback;
}

export function requiredEnv(key: string): string {
  const value = process.env[key]?.trim();
  if (!value) throw new Error(`missing required environment variable: ${key}`);
  return value;
}
