// stock-api only touches the auth tables (usuarios, refresh_tokens); the
// rest of the schema belongs to stock-operations.
import pg from 'pg';
import { env } from '../config/env.js';

export const pool = new pg.Pool({
  host: env('DB_HOST', 'localhost'),
  port: Number(env('DB_PORT', '5432')),
  user: env('DB_USER', 'admin_dev'),
  password: env('DB_PASSWORD', 'password_dev'),
  database: env('DB_NAME', 'stock_db'),
});
