import { expect, test, type Page } from '@playwright/test';

// Route-based fault injection only; original PWA/service-worker tests unchanged.
test.use({ serviceWorkers: 'block' });

async function openSearchWithWallet(page: Page) {
  const wallet = `Search ${crypto.randomUUID().slice(0, 8)}`;
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
  await page.getByRole('button', { name: 'Tìm kiếm', exact: true }).click();
  return wallet;
}

test('search distinguishes API failure and retries the retained query against real data', async ({ page }, testInfo) => {
  const wallet = await openSearchWithWallet(page);
  let requests = 0;
  await page.route('**/api/v1/search?*', async route => {
    requests++;
    if (requests === 1) await route.fulfill({ status: 503, contentType: 'application/json', body: JSON.stringify({ error: { code: 'INTERNAL_FAILURE', message: 'PRIVATE_SEARCH_DETAIL' }, correlation_id: 'req_search_browser' }) });
    else await route.continue();
  });
  const panel = page.getByRole('region', { name: 'Tìm kiếm', exact: true });
  const input = page.getByLabel('Tìm kiếm giao dịch và ví');
  await input.fill(wallet);
  await expect(panel.getByRole('alert')).toContainText('req_search_browser');
  await expect(input).toHaveValue(wallet);
  await expect(panel.getByText('Không tìm thấy kết quả.')).toHaveCount(0);
  await expect(panel.getByText('PRIVATE_SEARCH_DETAIL')).toHaveCount(0);
  expect(requests).toBe(1);
  await panel.screenshot({ path: testInfo.outputPath('search-error.png') });
  await panel.getByRole('button', { name: 'Thử lại', exact: true }).click();
  await expect(panel.getByText(wallet, { exact: true })).toBeVisible();
  await expect(panel.getByRole('alert')).toHaveCount(0);
  await panel.getByRole('button', { name: 'Xóa từ khóa tìm kiếm' }).click();
  await expect(input).toHaveValue('');
  await expect(panel.getByText(wallet, { exact: true })).toHaveCount(0);
});

test('a late real search response cannot replace the newer empty query result', async ({ page }) => {
  const wallet = await openSearchWithWallet(page);
  let release!: () => void;
  const gate = new Promise<void>(resolve => { release = resolve; });
  let arrived!: () => void;
  const waiting = new Promise<void>(resolve => { arrived = resolve; });
  const missing = `No-match-${crypto.randomUUID()}`;
  await page.route('**/api/v1/search?*', async route => {
    if (new URL(route.request().url()).searchParams.get('q') === wallet) {
      const response = await route.fetch();
      arrived();
      await gate;
      await route.fulfill({ response });
    } else await route.continue();
  });
  const panel = page.getByRole('region', { name: 'Tìm kiếm', exact: true });
  const input = page.getByLabel('Tìm kiếm giao dịch và ví');
  try {
    await input.fill(wallet);
    await waiting;
    await expect(panel.getByRole('status')).toContainText('Đang tìm kiếm');
    await expect(panel.getByText('Không tìm thấy kết quả.')).toHaveCount(0);
    await input.fill(missing);
    await expect(panel.getByText('Không tìm thấy kết quả.')).toBeVisible();
    const delivered = page.waitForResponse(response => new URL(response.url()).pathname === '/api/v1/search' && new URL(response.url()).searchParams.get('q') === wallet);
    release();
    await (await delivered).finished();
    await expect(input).toHaveValue(missing);
    await expect(panel.getByText(wallet, { exact: true })).toHaveCount(0);
    await expect(panel.getByText('Không tìm thấy kết quả.')).toBeVisible();
  } finally { release(); }
});
