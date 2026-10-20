import { expect, test } from '@playwright/test';

test('expanded install help leaves the last account action reachable', async ({ page }, testInfo) => {
  await page.goto('/');
  await page.getByRole('button', { name: 'Đăng nhập bằng Google' }).click();
  await page.getByRole('button', { name: 'Tài khoản', exact: true }).click();
  await page.getByRole('button', { name: 'Cài ứng dụng', exact: true }).click();
  await expect(page.getByText('Safari: Chia sẻ → Thêm vào Màn hình chính')).toBeVisible();
  await page.getByRole('button', { name: 'Đăng xuất', exact: true }).scrollIntoViewIfNeeded();
  await page.screenshot({ path: testInfo.outputPath('account-install-help.png') });
  await page.getByRole('button', { name: 'Đăng xuất', exact: true }).click();
  await expect(page.getByRole('button', { name: 'Đăng nhập bằng Google' })).toBeVisible();
});

test('dismissed install help stays closed while navigating tabs', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('button', { name: 'Đăng nhập bằng Google' }).click();
  await page.getByRole('button', { name: 'Để sau', exact: true }).click();
  await page.getByRole('button', { name: 'Tài khoản', exact: true }).click();
  await expect(page.getByRole('dialog', { name: 'Cài MyPocket' })).toHaveCount(0);
  await page.getByRole('button', { name: 'Đăng xuất', exact: true }).click();
  await expect(page.getByRole('button', { name: 'Đăng nhập bằng Google' })).toBeVisible();
});
