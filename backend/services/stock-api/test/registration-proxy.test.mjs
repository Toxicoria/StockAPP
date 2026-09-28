import assert from 'node:assert/strict';
import { createServer } from 'node:http';
import { after, before, test } from 'node:test';
import { createApp } from '../dist/app.js';

let upstream;
let gateway;
let baseUrl;
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
