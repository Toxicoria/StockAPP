// stock-api only touches the auth tables (usuarios, refresh_tokens); the
// rest of the schema belongs to stock-operations.
import pg from 'pg';
import { env, requiredEnv } from '../config/env.js';

const connectionString = process.env.DATABASE_URL?.trim();

export const pool = new pg.Pool(
  connectionString
    ? { connectionString }
    : {
        host: env('DB_HOST', 'localhost'),
        port: Number(env('DB_PORT', '5432')),
        user: env('DB_USER', 'admin_dev'),
        password: requiredEnv('DB_PASSWORD'),
        database: env('DB_NAME', 'stock_db'),
      },
);
