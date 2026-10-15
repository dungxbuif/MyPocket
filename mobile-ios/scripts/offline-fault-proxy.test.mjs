import test from 'node:test';
import assert from 'node:assert/strict';
import { createServer } from 'node:http';
import { createHash } from 'node:crypto';
import { createFaultProxy } from './offline-fault-proxy.mjs';

const walletID = '00000000-0000-4000-8000-000000000001';
const otherID = '00000000-0000-4000-8000-000000000002';
async function serve(t, server) {
  await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
  t.after(() => new Promise(resolve => { server.closeAllConnections(); server.close(resolve); }));
  return `http://127.0.0.1:${server.address().port}`;
}
async function control(base, mode, reset = false) {
  const response = await fetch(base + '/__fault', { method: 'POST', headers: { 'content-type': 'application/json', 'x-mp-local-test': 'offline-e2e' }, body: JSON.stringify({ mode, walletID, reset }) });
  assert.equal(response.status, 200);
}
async function stats(base) {
  const response = await fetch(base + '/__fault', { headers: { 'x-mp-local-test': 'offline-e2e' } });
  return response.json();
}
test('refuses remote or credential-bearing upstream before binding', () => {
  for (const upstream of ['https://example.com', 'http://user:pass@127.0.0.1:8081', 'http://127.0.0.1:8081?secret=x']) {
    assert.throws(() => createFaultProxy({ upstream }));
  }
});
test('scoped503 blocks only target creates and stats never expose keys or payload', async t => {
  let forwarded = 0;
  const upstream = await serve(t, createServer((req, res) => { forwarded++; res.writeHead(201, { 'content-type': 'application/json' }); res.end('{"data":{"id":"fixture"}}'); }));
  const base = await serve(t, createFaultProxy({ upstream }));
  assert.equal((await fetch(base + '/__fault')).status, 403);
  await control(base, 'reject503', true);
  const key = 'synthetic-private-key';
  const create = id => fetch(base + '/api/v1/transactions', { method: 'POST', headers: { 'content-type': 'application/json', 'idempotency-key': key }, body: JSON.stringify({ wallet_id: id, note: 'synthetic-private-note', amount: 12300 }) });
  assert.equal((await create(walletID)).status, 503);
  assert.equal((await create(otherID)).status, 201);
  assert.equal(forwarded, 1);
  const result = await stats(base);
  assert.equal(result.blocked, 1);
  assert.deepEqual(result.keys, [createHash('sha256').update(key).digest('hex')]);
  assert.equal(JSON.stringify(result).includes(key), false);
  assert.equal(JSON.stringify(result).includes('synthetic-private-note'), false);
});
test('commit then drop forwards request before losing response and pass keeps key history', async t => {
  let commits = 0;
  const upstream = await serve(t, createServer((req, res) => { commits++; res.writeHead(commits === 1 ? 201 : 200, { 'content-type': 'application/json' }); res.end('{"data":{"id":"same-row"}}'); }));
  const base = await serve(t, createFaultProxy({ upstream }));
  await control(base, 'commit-then-drop', true);
  const send = () => fetch(base + '/api/v1/transactions', { method: 'POST', headers: { 'content-type': 'application/json', 'idempotency-key': 'stable-key' }, body: JSON.stringify({ wallet_id: walletID }) });
  await assert.rejects(send);
  assert.equal(commits, 1);
  await control(base, 'pass');
  assert.equal((await send()).status, 200);
  const result = await stats(base);
  assert.equal(result.dropped, 1);
  assert.deepEqual(result.upstreamStatuses, [201, 200]);
  assert.equal(result.keys.length, 1);
});

test('transfer faults match source wallet only and leave other transfers untouched', async t => {
  let forwarded = 0;
  const upstream = await serve(t, createServer((req, res) => { forwarded++; res.writeHead(201, { 'content-type': 'application/json' }); res.end('{"data":[]}'); }));
  const base = await serve(t, createFaultProxy({ upstream }));
  await control(base, 'reject503', true);
  const send = (source, destination) => fetch(base + '/api/v1/transactions/transfer', {
    method: 'POST', headers: { 'content-type': 'application/json', 'idempotency-key': 'transfer-stable-key' },
    body: JSON.stringify({ source_wallet_id: source, destination_wallet_id: destination, amount: 12500 })
  });
  assert.equal((await send(walletID, otherID)).status, 503);
  assert.equal((await send(otherID, walletID)).status, 201);
  assert.equal(forwarded, 1);
  const result = await stats(base);
  assert.equal(result.blocked, 1); assert.equal(result.keys.length, 1);
  assert.equal(JSON.stringify(result).includes('transfer-stable-key'), false);
});

test('scoped conflict mode returns409 without forwarding a financial write', async t => {
  let forwarded = 0;
  const upstream = await serve(t, createServer((req, res) => { forwarded++; res.writeHead(201); res.end('{}'); }));
  const base = await serve(t, createFaultProxy({ upstream }));
  await control(base, 'reject409', true);
  const result = await fetch(base + '/api/v1/transactions', { method: 'POST',
    headers: { 'content-type': 'application/json', 'idempotency-key': 'conflict-key' },
    body: JSON.stringify({ wallet_id: walletID }) });
  assert.equal(result.status, 409); assert.equal(forwarded, 0);
  const observed = await stats(base); assert.equal(observed.blocked, 1); assert.equal(observed.keys.length, 1);
});
