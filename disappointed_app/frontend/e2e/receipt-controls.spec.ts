import { devices, expect, test, webkit, type Page, type BrowserContext } from '@playwright/test';

// Synthetic 1px PNG; never use a user's private receipts in test artifacts.
const receipt = { name: 'receipt-fixture.png', mimeType: 'image/png', buffer: Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+aX1kAAAAASUVORK5CYII=', 'base64') };

async function openAddWithWallet(page: Page) {
  const wallet = `Receipt ${crypto.randomUUID().slice(0, 8)}`;
  await page.goto('/');
  await page.getByRole('button', { name: 'Đăng nhập bằng Google' }).click();
  await expect(page.getByRole('heading', { name: 'Ví của tôi' })).toBeVisible();
  await page.getByRole('button', { name: 'Để sau', exact: true }).click();
  await page.getByRole('button', { name: 'Xem tất cả', exact: true }).click();
  await page.getByRole('button', { name: 'Thêm ví', exact: true }).click();
  await page.getByLabel('Tên ví mới').fill(wallet);
  await page.getByRole('button', { name: 'Tạo ví', exact: true }).click();
  await expect(page.getByRole("dialog").locator("strong").filter({ hasText: wallet })).toBeVisible();
  await page.getByRole('button', { name: 'Đóng', exact: true }).click();
  await page.getByRole('button', { name: 'Thêm giao dịch', exact: true }).click();
  await page.getByLabel('Ví nguồn').selectOption({ label: wallet });
}

test('both image buttons open the same picker and preserve selection across details', async ({ page }, testInfo) => {
  await openAddWithWallet(page);
  const firstPicker = page.waitForEvent('filechooser');
  await page.getByRole('button', { name: 'Đính kèm ảnh', exact: true }).click();
  await (await firstPicker).setFiles(receipt);
  await expect(page.getByRole('status')).toContainText(receipt.name);
  await page.getByRole('button', { name: 'Thêm chi tiết', exact: true }).click();
  for (const label of ['Với', 'Đặt vị trí', 'Chọn sự kiện', 'Đặt nhắc nhở']) {
    await expect(page.getByRole('button', { name: new RegExp(label + ' Chưa hỗ trợ') })).toBeDisabled();
  }
  const detailButton = page.getByRole('button', { name: 'Đổi ảnh', exact: true });
  await detailButton.focus();
  const secondPicker = page.waitForEvent('filechooser');
  await page.keyboard.press('Enter');
  await (await secondPicker).setFiles({ ...receipt, name: 'replacement.png' });
  await page.getByRole('button', { name: 'Ẩn chi tiết', exact: true }).click();
  await expect(page.getByRole('status')).toContainText('replacement.png');
  await page.getByRole('status').scrollIntoViewIfNeeded();
  await page.screenshot({ path: testInfo.outputPath('receipt-selected.png') });
  await page.getByRole('button', { name: 'Bỏ ảnh', exact: true }).click();
  await expect(page.getByRole('status')).toHaveCount(0);
  const thirdPicker = page.waitForEvent('filechooser');
  await page.getByRole('button', { name: 'Đính kèm ảnh', exact: true }).click();
  await (await thirdPicker).setFiles(receipt);
  await expect(page.getByRole('status')).toContainText(receipt.name);
});

type ReceiptScenario = { page: Page; context: BrowserContext; browserName: string; baseURL?: string };

test('offline receipt bytes persist against the saved transaction without uploading', async ({ page, context, browserName, baseURL }) => {
  await verifyOfflineReceipt({ page, context, browserName, baseURL }, 'emulated');
});

test('HTTP outage retains offline receipt bytes against the saved transaction', async ({ page, context, browserName, baseURL }) => {
  await verifyOfflineReceipt({ page, context, browserName, baseURL }, 'http-outage');
});

async function verifyOfflineReceipt({ page: defaultPage, context: defaultContext, browserName, baseURL }: ReceiptScenario, network: 'emulated' | 'http-outage') {
  // Ephemeral WebKit contexts reject all IndexedDB Blob/File writes, even in a
  // minimal fixture. Use an isolated disposable on-disk profile for this storage
  // test only; do not change the application's storage format to fit emulation.
  const persistent = browserName === 'webkit' ? await webkit.launchPersistentContext('', {
    ...devices['iPhone 13'], baseURL,
    // Routing does not intercept service-worker-controlled requests. Only the
    // separate HTTP-outage scenario blocks workers; original PWA proof is kept.
    serviceWorkers: network === 'http-outage' ? 'block' : 'allow',
  }) : null;
  const context = persistent ?? defaultContext;
  const page = persistent ? await persistent.newPage() : defaultPage;
  try {
  await openAddWithWallet(page);
  const me = await page.request.get('http://127.0.0.1:18173/api/v1/me');
  expect(me.ok()).toBeTruthy();
  const { user } = await me.json();
  await page.getByLabel('Số tiền').fill('23000');
  await page.getByLabel('Ghi chú').fill('Offline receipt fixture');
  const picker = page.waitForEvent('filechooser');
  await page.getByRole('button', { name: 'Đính kèm ảnh', exact: true }).click();
  await (await picker).setFiles(receipt);
  if (network === 'emulated') {
    await context.setOffline(true);
  } else {
    // Keep WebKit's local Blob reader working while HTTP is genuinely aborted.
    // This is a separate signal/transport simulation, not physical airplane mode.
    let blockedHealthRequests = 0;
    await context.route(/^https?:\/\//, route => {
      if (route.request().url() === 'http://127.0.0.1:18173/api/v1/health/live') blockedHealthRequests++;
      return route.abort('internetdisconnected');
    });
    await page.evaluate(() => Object.defineProperty(navigator, 'onLine', { configurable: true, get: () => false }));
    expect(await page.evaluate(async () => {
      try { await fetch('http://127.0.0.1:18173/api/v1/health/live', { cache: 'no-store' }); return false; }
      catch { return true; }
    })).toBe(true);
    expect(blockedHealthRequests).toBeGreaterThan(0);
  }
  await page.evaluate(() => window.dispatchEvent(new Event('offline')));
  await expect(page.getByText('Offline', { exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'Lưu', exact: true }).click();
  await expect(page.getByRole('dialog', { name: 'Thêm Giao Dịch' })).toHaveCount(0);
  await expect(page.getByText('1 chờ đồng bộ')).toBeVisible();
  // WebKit can persist an IndexedDB Blob while offline but throws NotReadableError
  // when its bytes are read before transport is restored. Restore transport only
  // for the storage read; route all HTTP requests away so queued data cannot sync.
  if (network === 'emulated' && browserName === 'webkit') {
    await context.setOffline(false);
    await context.route(/^https?:\/\//, route => route.abort('internetdisconnected'));
  }
  const saved = await page.evaluate(async userID => {
    const db = await new Promise<IDBDatabase>((resolve, reject) => {
      const request = indexedDB.open(`mypocket.offline.v1:user:${encodeURIComponent(userID)}`);
      request.onsuccess = () => resolve(request.result);
      request.onerror = () => reject(request.error);
    });
    try {
      const readAll = <T,>(store: string) => new Promise<T[]>((resolve, reject) => {
        const request = db.transaction(store).objectStore(store).getAll();
        request.onsuccess = () => resolve(request.result);
        request.onerror = () => reject(request.error);
      });
      const receipts = await readAll<{ transaction_id: string; filename: string; file: Blob }>('receipts');
      const pending = await readAll<{ entity_id: string; entity_type: string; state: string }>('outbox');
      const transactions = await readAll<{ id: string; amount_vnd: number }>('transactions');
      return {
        receipts: await Promise.all(receipts.map(async item => ({ transactionID: item.transaction_id, filename: item.filename, bytes: Array.from(new Uint8Array(await item.file.arrayBuffer())) }))),
        pending: pending.filter(item => item.entity_type === 'transaction' && item.state === 'pending'),
        transactions,
      };
    } finally { db.close(); }
  }, user.id as string);
  expect(saved.receipts).toHaveLength(1);
  expect(saved.pending).toHaveLength(1);
  expect(saved.receipts[0]).toEqual({ transactionID: saved.pending[0].entity_id, filename: receipt.name, bytes: Array.from(receipt.buffer) });
  expect(saved.transactions.find(item => item.id === saved.receipts[0].transactionID)).toMatchObject({ amount_vnd: 23000 });
  } finally { await persistent?.close(); }
}
