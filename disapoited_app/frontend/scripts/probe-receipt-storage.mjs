// Synthetic data only. No MyPocket API, production profile or user files.
// Exit zero means all diagnostic cases ran; inspect each case's result.
import { createServer } from 'node:http';
import { webkit } from '@playwright/test';

const server = createServer((_request, response) => {
  response.setHeader('Content-Type', 'text/html');
  response.end('<!doctype html><input type="file">');
});
await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
const browser = await webkit.launch();
try {
  for (const mode of ['ephemeral', 'temporaryPersistent']) {
    const context = mode === 'ephemeral' ? await browser.newContext() : await webkit.launchPersistentContext('');
    try {
      const page = await context.newPage();
      await page.goto(`http://127.0.0.1:${server.address().port}`);
      for (const offline of [false, true]) {
      await context.setOffline(offline);
      for (const resetInput of [false, true]) {
      for (const strategy of ['nativeFile', 'memoryBlob', 'arrayBuffer']) {
        await page.locator('input').setInputFiles({ name: 'probe.png', mimeType: 'image/png', buffer: Buffer.from([1, 2, 3]) });
        const result = await page.evaluate(async ({ strategy, resetInput }) => {
          let db;
          let stage = 'open';
          try {
            db = await new Promise((resolve, reject) => {
              const request = indexedDB.open('receipt-probe', 1);
              request.onupgradeneeded = () => request.result.createObjectStore('receipts');
              request.onsuccess = () => resolve(request.result);
              request.onerror = () => reject(request.error);
            });
            const file = document.querySelector('input').files[0];
            if (resetInput) document.querySelector('input').value = '';
            stage = 'source-read';
            const bytes = await file.arrayBuffer();
            const value = strategy === 'nativeFile' ? file : strategy === 'memoryBlob' ? new Blob([bytes], { type: 'image/png' }) : bytes;
            stage = 'write';
            await new Promise((resolve, reject) => {
              const transaction = db.transaction('receipts', 'readwrite');
              transaction.oncomplete = resolve;
              transaction.onabort = () => reject(transaction.error);
              const request = transaction.objectStore('receipts').put(value, strategy);
              request.onerror = () => reject(request.error);
            });
            stage = 'stored-read';
            const stored = await new Promise((resolve, reject) => {
              const request = db.transaction('receipts').objectStore('receipts').get(strategy);
              request.onsuccess = () => resolve(request.result);
              request.onerror = () => reject(request.error);
            });
            const actual = Array.from(new Uint8Array(stored instanceof Blob ? await stored.arrayBuffer() : stored));
            if (JSON.stringify(actual) !== '[1,2,3]') throw new Error('Stored bytes differ from fixture');
            return { result: 'pass', stage, bytes: actual };
          } catch (error) {
            return { result: 'fail', stage, name: error?.name, error: error?.message };
          } finally { db?.close(); }
        }, { strategy, resetInput });
        console.log(JSON.stringify({ engine: 'webkit', version: browser.version(), mode, strategy, resetInput, offline, ...result }));
      }
      }
      }
      // Read the same online-written record without rewriting it. This isolates
      // offline emulation from input reset, file selection and persistence.
      if (mode === 'temporaryPersistent') {
        for (const offline of [false, true, false]) {
          await context.setOffline(offline);
          const result = await page.evaluate(async () => {
            let db;
            try {
              db = await new Promise((resolve, reject) => {
                const request = indexedDB.open('receipt-probe', 1);
                request.onsuccess = () => resolve(request.result);
                request.onerror = () => reject(request.error);
              });
              const stored = await new Promise((resolve, reject) => {
                const request = db.transaction('receipts').objectStore('receipts').get('nativeFile');
                request.onsuccess = () => resolve(request.result);
                request.onerror = () => reject(request.error);
              });
              const bytes = Array.from(new Uint8Array(await stored.arrayBuffer()));
              if (JSON.stringify(bytes) !== '[1,2,3]') throw new Error('Stored bytes differ');
              return { result: 'pass', bytes };
            } catch (error) {
              return { result: 'fail', name: error?.name, error: error?.message };
            } finally { db?.close(); }
          });
          console.log(JSON.stringify({ mode, scenario: 'same-record-offline-roundtrip', offline, ...result }));
        }
      }
    } finally { await context.close(); }
  }
} finally {
  await browser.close();
  server.closeAllConnections();
  await new Promise(resolve => server.close(resolve));
}
