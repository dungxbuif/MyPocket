import { createServer } from 'node:http';
import { createHash } from 'node:crypto';
import { pathToFileURL } from 'node:url';

export function createFaultProxy({ upstream }) {
  const origin = new URL(upstream);
  if (origin.protocol !== 'http:' || !['127.0.0.1', '[::1]'].includes(origin.hostname) ||
      origin.username || origin.password || origin.search || origin.hash || origin.pathname !== '/') {
    throw new Error('Fault proxy upstream must be an explicit credential-free loopback HTTP origin');
  }
  let mode = 'pass', walletID = '';
  let blocked = 0, forwarded = 0, dropped = 0;
  let keys = new Set(), upstreamStatuses = [];
  const json = (response, status, value) => {
    response.writeHead(status, { 'content-type': 'application/json' });
    response.end(JSON.stringify(value));
  };
  return createServer(async (request, response) => {
    try {
      const route = new URL(request.url, 'http://local.invalid');
      const chunks = []; let size = 0;
      for await (const chunk of request) {
        size += chunk.length;
        if (size > 1024 * 1024) { json(response, 413, { code: 'LOCAL_TEST_BODY_TOO_LARGE' }); return; }
        chunks.push(chunk);
      }
      const body = Buffer.concat(chunks);
      if (route.pathname === '/__fault') {
        if (request.headers['x-mp-local-test'] !== 'offline-e2e') { json(response, 403, { code: 'LOCAL_TEST_CONTROL_REQUIRED' }); return; }
        if (request.method === 'POST') {
          const input = JSON.parse(body.toString());
          if (!['pass', 'reject503', 'reject422', 'reject409', 'commit-then-drop'].includes(input.mode) ||
              typeof input.walletID !== 'string' || !/^[a-fA-F0-9-]{36}$/.test(input.walletID)) {
            json(response, 400, { code: 'INVALID_LOCAL_TEST_MODE' }); return;
          }
          mode = input.mode; walletID = input.walletID;
          if (input.reset === true) { blocked = 0; forwarded = 0; dropped = 0; keys = new Set(); upstreamStatuses = []; }
        } else if (request.method !== 'GET') { json(response, 405, {}); return; }
        json(response, 200, { mode, blocked, forwarded, dropped, keys: [...keys], upstreamStatuses }); return;
      }
      let scoped = false;
      if (request.method === 'POST' && ['/api/v1/transactions', '/api/v1/transactions/transfer'].includes(route.pathname)) {
        try {
          const input = JSON.parse(body.toString());
          const source = route.pathname.endsWith('/transfer') ? input.source_wallet_id : input.wallet_id;
          scoped = source === walletID && walletID !== '';
        } catch {}
      }
      if (scoped) {
        const key = request.headers['idempotency-key'];
        if (typeof key === 'string') keys.add(createHash('sha256').update(key).digest('hex'));
        if (mode === 'reject503' || mode === 'reject422' || mode === 'reject409') {
          blocked++;
          const status = mode === 'reject503' ? 503 : mode === 'reject409' ? 409 : 422;
          json(response, status, { title: 'Local E2E fault', status, code: 'LOCAL_TEST_FAULT', detail: 'Controlled synthetic test failure' }); return;
        }
      }
      const target = new URL(origin.origin);
      target.pathname = route.pathname; target.search = route.search;
      const headers = { ...request.headers };
      for (const name of ['host', 'connection', 'content-length', 'transfer-encoding']) delete headers[name];
      const result = await fetch(target, { method: request.method, headers,
        body: ['GET', 'HEAD'].includes(request.method) ? undefined : body, redirect: 'manual' });
      const data = Buffer.from(await result.arrayBuffer());
      if (scoped) { forwarded++; upstreamStatuses.push(result.status); }
      if (scoped && mode === 'commit-then-drop' && result.ok) { dropped++; response.destroy(); return; }
      const responseHeaders = {};
      for (const [name, value] of result.headers) {
        if (!['connection', 'transfer-encoding', 'content-length', 'content-encoding'].includes(name)) responseHeaders[name] = value;
      }
      responseHeaders['content-length'] = String(data.length);
      response.writeHead(result.status, responseHeaders); response.end(data);
    } catch {
      if (!response.headersSent) json(response, 502, { code: 'LOCAL_TEST_PROXY_FAILURE' });
      else response.destroy();
    }
  });
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  const port = Number(process.env.MP_FAULT_PROXY_PORT ?? 18081);
  if (!Number.isInteger(port) || port < 1024 || port > 65535) throw new Error('Invalid test port');
  const server = createFaultProxy({ upstream: process.env.MP_FAULT_UPSTREAM ?? 'http://127.0.0.1:8081' });
  server.listen(port, '127.0.0.1', () => process.stdout.write(`Local offline fault proxy ready on 127.0.0.1:${port}\n`));
  process.on('SIGTERM', () => { server.closeAllConnections(); server.close(() => process.exit(0)); });
}
