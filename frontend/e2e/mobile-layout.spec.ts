import { expect, test } from '@playwright/test';

// WebKit worker-owned requests bypass page.route. These layout-only fixtures
// need deterministic API responses; the real PWA/offline suites retain workers.
test.use({ serviceWorkers: 'block' });

for (const count of [0, 100]) {
  test(`large report labels preserve navigation with ${count} wallets`, async ({ page }) => {
    await page.route('**/api/v1/wallets', async route => {
      if (route.request().method() !== 'GET') return route.continue();
      await route.fulfill({ json: { wallets: Array.from({ length: count }, (_, i) => ({
        id: `fixture-${i}`, name: `Wallet ${i} with a deliberately long name for mobile layout`, type: 'cash', balance_vnd: 0,
        include_in_total: true, is_default_ai: false, version: 1,
      })) } });
    });
    // Fixed historical data keeps chart width independent of the date and shared fixtures.
    await page.route('**/api/v1/reports/daily', route => route.fulfill({ json: { report: {
      summary: { income_vnd: 0, expense_vnd: 864197523, net_income_vnd: -864197523,
        generated_at: '2020-01-07T12:00:00Z', timezone: 'Asia/Ho_Chi_Minh',
        from: '2020-01-01', to: '2020-01-07', data_version: 1 },
      daily: Array.from({ length: 7 }, (_, i) => ({ date: `2020-01-0${i + 1}`,
        income_vnd: 0, expense_vnd: 123456789, net_income_vnd: -123456789,
        cumulative_net_vnd: -123456789 * (i + 1) })),
    } } }));
    await page.goto('/');
    await page.getByRole('button', { name: 'Đăng nhập bằng Google' }).click();
    await expect(page.getByRole('heading', { name: 'Ví của tôi' })).toBeVisible();
    await expect(page.getByTitle('123.456.789 đ', { exact: true })).toHaveCount(7);
    await expect(page.locator('.pwa-install-prompt')).toBeVisible();
    await expect.poll(() => page.evaluate(() =>
      document.documentElement.scrollWidth - document.documentElement.clientWidth
    )).toBeLessThanOrEqual(1);
    // Use real hit testing with the install overlay still present.
    await page.getByRole('button', { name: 'Tài khoản', exact: true }).click();
    await expect(page.getByRole('button', { name: 'Đăng xuất', exact: true })).toBeVisible();
  });
}
