import { expect, test } from '@playwright/test';

test('mobile shell registers service worker and reloads offline', async ({ page, context }) => {
  await page.goto('/');
  await page.getByRole("button", { name: "Đăng nhập bằng Google" }).click();
  await expect(page.getByRole('heading', { name: 'Ví của tôi' })).toBeVisible();
  await expect(page.locator('link[rel="manifest"]')).toHaveAttribute('href', /manifest\.webmanifest/);
  await expect.poll(() => page.evaluate(async () => (await navigator.serviceWorker.getRegistrations()).length)).toBeGreaterThan(0);
  await page.evaluate(() => navigator.serviceWorker.ready);
  await page.reload();
  await expect(page.getByRole('heading', { name: 'Ví của tôi' })).toBeVisible();
  await expect.poll(() => page.evaluate(() => Boolean(navigator.serviceWorker.controller))).toBe(true);
  await expect.poll(() => page.evaluate(async () => Boolean(await caches.match("/")))).toBe(true);
  await expect.poll(() => page.evaluate(() => localStorage.getItem("mypocket.current-user.v1"))).not.toBeNull();
  // Playwright WebKit aborts an offline `page.reload()` before the service worker
  // can serve its cache. Keep the browser transport available only on WebKit and
  // model the navigator offline transition across this reload; Chromium still
  // exercises native context offline mode below.
  if (test.info().project.name === 'webkit-mobile') {
    await page.addInitScript(() => Object.defineProperty(navigator, 'onLine', { configurable: true, get: () => false }));
    await page.evaluate(() => {
      Object.defineProperty(navigator, 'onLine', { configurable: true, get: () => false });
      window.dispatchEvent(new Event('offline'));
    });
  } else {
    await context.setOffline(true);
  }
  await expect.poll(() => page.evaluate(() => navigator.onLine)).toBe(false);
  await page.reload();
  await expect.poll(() => page.evaluate(() => localStorage.getItem("mypocket.current-user.v1"))).not.toBeNull();
  await expect(page.getByRole('heading', { name: 'Ví của tôi' })).toBeVisible();
});
