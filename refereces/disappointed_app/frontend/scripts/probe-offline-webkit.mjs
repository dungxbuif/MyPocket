// Read-only diagnostic: disposable browser contexts and loopback fixture server.
// Does not use MyPocket API, user data, or production services.
import { createServer } from 'node:http';
import { readFile } from 'node:fs/promises';
import { chromium, webkit } from '@playwright/test';

const appWorker = await readFile(new URL('../public/sw.js', import.meta.url), 'utf8');
const minimalWorker = `
self.addEventListener('install', event => { event.waitUntil(caches.open('probe').then(cache => cache.add('/'))); self.skipWaiting(); });
self.addEventListener('activate', event => event.waitUntil(self.clients.claim()));
self.addEventListener('fetch', event => { event.respondWith(caches.match('/')); });
`;
for (const [engine, launcher] of Object.entries({ chromium, webkit })) {
  const browser = await launcher.launch();
  try {
    for (const [strategy, worker] of Object.entries({ minimalCacheOnly: minimalWorker, currentAppWorker: appWorker })) {
      for (const reload of ['automation', 'pageInitiated', 'serverUnavailable']) {
        const server = createServer((request, response) => {
          response.setHeader('Content-Type', request.url === '/sw.js' ? 'text/javascript' : 'text/html');
          response.end(request.url === '/sw.js' ? worker : '<!doctype html><title>Offline probe</title><h1>Cached fixture</h1>');
        });
        await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
        const address = `http://127.0.0.1:${server.address().port}`;
        const context = await browser.newContext();
        const page = await context.newPage();
        const errors = [];
        page.on('pageerror', error => errors.push(error.message));
        let before;
        try {
          await page.goto(address);
          await page.evaluate(async () => { await navigator.serviceWorker.register('/sw.js'); await navigator.serviceWorker.ready; });
          await page.waitForFunction(() => Boolean(navigator.serviceWorker.controller));
          before = await page.evaluate(async () => ({ controller: Boolean(navigator.serviceWorker.controller), cached: Boolean(await caches.match('/')) }));
          if (reload === 'serverUnavailable') {
            server.closeAllConnections();
            await new Promise(resolve => server.close(resolve));
          } else await context.setOffline(true);
          if (reload !== 'pageInitiated') await page.reload({ timeout: 5000 });
          else await Promise.all([
            page.waitForEvent('load', { timeout: 5000 }),
            page.evaluate(() => { setTimeout(() => location.reload(), 0); }),
          ]);
          const heading = await page.locator('h1').textContent({ timeout: 5000 });
          console.log(JSON.stringify({ engine, version: browser.version(), strategy, reload, before, result: heading === 'Cached fixture' ? 'pass' : 'wrong-content', errors }));
        } catch (error) {
          console.log(JSON.stringify({ engine, version: browser.version(), strategy, reload, before, result: 'fail', error: error.message, errors }));
        } finally {
          await context.close();
          if (server.listening) {
            server.closeAllConnections();
            await new Promise(resolve => server.close(resolve));
          }
        }
      }
    }
  } finally { await browser.close(); }
}
