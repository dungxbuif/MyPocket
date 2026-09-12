import { expect, test } from '@playwright/test';

// One-time fixture keys must not be retained in traces/screenshots/videos.
// These fault-injection tests require route interception, which WebKit bypasses
// with a controlling worker. PWA/offline suites keep service workers enabled.
test.use({ trace: 'off', screenshot: 'off', video: 'off', serviceWorkers: 'block' });

test('rejected transaction preserves input and only saves after an explicit retry', async ({ page }) => {
  const suffix = crypto.randomUUID().slice(0, 8);
  const wallet = `Feedback ${suffix}`;
  const note = `Preserved ${suffix}`;
  await page.goto('/');
  await page.getByRole('button', { name: 'Đăng nhập bằng Google' }).click();
  await page.getByRole('button', { name: 'Để sau', exact: true }).click();
  await page.getByRole('button', { name: 'Xem tất cả' }).click();
  await page.getByRole('button', { name: 'Thêm ví', exact: true }).click();
  await page.getByLabel('Tên ví mới').fill(wallet);
  await page.getByRole('button', { name: 'Tạo ví' }).click();
  await expect(page.getByRole("dialog").locator("strong").filter({ hasText: wallet })).toBeVisible();
  await page.getByRole('button', { name: 'Đóng', exact: true }).click();
  let writes = 0;
  let rejectWrite = true;
  await page.route('**/api/v1/transactions', async route => {
    if (route.request().method() !== 'POST') return route.continue();
    writes++;
    if (!rejectWrite) return route.continue();
    return route.fulfill({ status: 422, json: { error: { code: 'VALIDATION', message: 'RAW_NOT_FOR_UI' }, correlation_id: 'req_browser_feedback' } });
  });
  await page.getByRole('button', { name: 'Thêm giao dịch' }).click();
  await page.getByRole('button', { name: 'Khoản thu', exact: true }).click();
  await page.getByLabel('Số tiền').fill('12000');
  await page.getByLabel('Ví nguồn').selectOption({ label: wallet });
  await page.getByLabel('Ghi chú').fill(note);
  await page.getByRole('button', { name: 'Lưu', exact: true }).click();
  await expect.poll(() => writes).toBe(1);
  await expect(page.getByRole('alert')).toContainText('req_browser_feedback');
  await expect(page.getByRole('alert')).not.toContainText('RAW_NOT_FOR_UI');
  await expect(page.getByLabel('Số tiền')).toHaveValue('12000');
  await expect(page.getByLabel('Ghi chú')).toHaveValue(note);
  expect(writes).toBe(1);
  rejectWrite = false;
  await page.getByRole('button', { name: 'Lưu', exact: true }).click();
  await expect(page.getByRole('dialog')).toHaveCount(0);
  expect(writes).toBe(2);
  await page.reload();
  await page.getByRole('button', { name: 'Sổ giao dịch', exact: true }).click();
  await expect(page.getByText(note, { exact: true })).toHaveCount(1);
});

test('confirmed API key create and revoke survive a failed list refresh', async ({ page }) => {
  const name = `Feedback key ${crypto.randomUUID().slice(0, 8)}`;
  await page.goto('/');
  await page.getByRole('button', { name: 'Đăng nhập bằng Google' }).click();
  await page.getByRole('button', { name: 'Để sau', exact: true }).click();
  await page.getByRole('button', { name: 'Tài khoản', exact: true }).click();
  await expect(page.getByRole('button', { name: 'Tạo key', exact: true })).toBeEnabled();
  let rejectRead = true;
  let rejectedReads = 0;
  await page.route('**/api/v1/api-keys', async route => {
    if (route.request().method() !== 'GET' || !rejectRead) return route.continue();
    rejectedReads++;
    return route.fulfill({ status: 503, json: { error: { code: 'UNAVAILABLE', message: 'RAW_NOT_FOR_UI' }, correlation_id: 'req_key_refresh' } });
  });
  await page.getByLabel('Tên API key').fill(name);
  await page.getByRole('button', { name: 'Tạo key', exact: true }).click();
  await expect(page.getByText('Chỉ hiển thị một lần', { exact: true })).toBeVisible();
  await expect.poll(() => rejectedReads).toBe(1);
  await expect(page.getByRole('alert')).toContainText('req_key_refresh');
  const row = page.locator('.api-key-row').filter({ has: page.getByText(name, { exact: true }) });
  await expect(row.getByRole('button', { name: 'Revoke', exact: true })).toBeEnabled();
  await row.getByRole('button', { name: 'Revoke', exact: true }).click();
  await expect(row.getByRole('button', { name: 'Revoke', exact: true })).toBeDisabled();
  await expect(row).toContainText('Đã revoke');
  await expect(page.getByText('Chỉ hiển thị một lần', { exact: true })).toHaveCount(0);
  rejectRead = false;
  await expect(page.getByRole('button', { name: 'Tải lại danh sách key' })).toBeEnabled();
  await page.getByRole('button', { name: 'Tải lại danh sách key' }).click();
  await expect(page.getByRole('alert')).toHaveCount(0);
  await expect(row).toContainText('Đã revoke');
});
