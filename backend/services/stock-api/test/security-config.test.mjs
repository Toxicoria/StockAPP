import assert from 'node:assert/strict';
import { test } from 'node:test';
import { requiredEnv } from '../dist/config/env.js';
import { jwtSecret } from '../dist/services/tokenService.js';

test('required environment variables reject missing values', () => {
  delete process.env.REQUIRED_ENV_TEST;
  assert.throws(() => requiredEnv('REQUIRED_ENV_TEST'), /missing required environment variable/);
});

test('JWT_SECRET rejects weak values', () => {
  process.env.JWT_SECRET = 'short';
  assert.throws(() => jwtSecret(), /at least 32 characters/);
});

test('JWT_SECRET accepts a non-empty 32-character secret', () => {
  process.env.JWT_SECRET = '12345678901234567890123456789012';
  assert.equal(jwtSecret(), process.env.JWT_SECRET);
});
