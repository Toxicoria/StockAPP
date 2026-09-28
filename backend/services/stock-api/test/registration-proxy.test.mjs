import assert from 'node:assert/strict';
import { createServer } from 'node:http';
import { after, before, test } from 'node:test';

let upstream;
let gateway;
let baseUrl;
let createApp;
const received = [];

function listen(server) {
  return new Promise((resolve, reject) => {
    const onError = (error) => reject(error);
    server.once('error', onError);
    server.listen(0, '127.0.0.1', () => {
      server.off('error', onError);
      resolve(server);
    });
  });
}

before(async () => {
  process.env.DB_PASSWORD = 'test-only-password';
  process.env.JWT_SECRET = 'test-only-jwt-secret-with-enough-entropy';
  ({ createApp } = await import('../dist/app.js'));

  upstream = createServer((req, res) => {
    received.push({ method: req.method, url: req.url, authorization: req.headers.authorization });
    res.writeHead(req.method === 'POST' ? 201 : 200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify(req.method === 'POST' ? { id_negocio: 1 } : { registrado: false }));
  });
  await listen(upstream);
  const upstreamAddress = upstream.address();
  process.env.GO_INTERNO = `http://127.0.0.1:${upstreamAddress.port}`;

  gateway = await listen(createServer(createApp()));
  const gatewayAddress = gateway.address();
  baseUrl = `http://127.0.0.1:${gatewayAddress.port}`;
});

after(async () => {
  await Promise.all([
    new Promise((resolve, reject) => gateway.close((err) => (err ? reject(err) : resolve()))),
    new Promise((resolve, reject) => upstream.close((err) => (err ? reject(err) : resolve()))),
  ]);
});

test('GET /api/registro is forwarded without an access token', async () => {
  const response = await fetch(`${baseUrl}/api/registro`);

  assert.equal(response.status, 200);
  assert.deepEqual(await response.json(), { registrado: false });
  assert.deepEqual(received.at(-1), { method: 'GET', url: '/api/registro', authorization: undefined });
});

test('POST /api/registro is forwarded without an access token', async () => {
  const response = await fetch(`${baseUrl}/api/registro`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ nombre_negocio: 'Prueba' }),
  });

  assert.equal(response.status, 201);
  assert.deepEqual(await response.json(), { id_negocio: 1 });
  assert.deepEqual(received.at(-1), { method: 'POST', url: '/api/registro', authorization: undefined });
});

test('other proxied routes remain protected', async () => {
  const callsBefore = received.length;
  const response = await fetch(`${baseUrl}/api/stock`);

  assert.equal(response.status, 401);
  assert.equal(received.length, callsBefore);
});

test('login is limited after ten attempts from the same client', async () => {
  for (let attempt = 1; attempt <= 10; attempt += 1) {
    const response = await fetch(`${baseUrl}/api/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: '{}',
    });
    assert.equal(response.status, 400);
  }

  const blocked = await fetch(`${baseUrl}/api/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: '{}',
  });
  assert.equal(blocked.status, 429);
  assert.match(blocked.headers.get('ratelimit') ?? '', /r=0/);
  assert.deepEqual(await blocked.json(), {
    error: 'demasiados intentos de ingreso; esperá 15 minutos',
  });
});
