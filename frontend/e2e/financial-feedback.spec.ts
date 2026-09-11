import { expect, test, type Page } from '@playwright/test';

// Fault injection is scoped to this spec; actual PWA/offline tests retain workers.
test.use({ serviceWorkers: 'block' });
const api = 'http://127.0.0.1:18173/api/v1';

async function ledger(page: Page) {
  const owner = `feedback-${crypto.randomUUID()}`;
  await page.goto(`${api}/auth/google/callback?subject=${owner}&email=${owner}@example.com&email_verified=true&name=Feedback`);
  await expect(page.getByRole('heading', { name: 'Ví của tôi' })).toBeVisible();
  const csrf = (await page.context().cookies()).find(c => c.name === 'mypocket_csrf')!.value;
  const post = async (path: string, data: object) => {
    const response = await page.request.post(`${api}/${path}`, { headers: { 'X-CSRF-Token': csrf, 'Idempotency-Key': crypto.randomUUID() }, data });
    expect(response.status(), await response.text()).toBe(201);
    return response.json();
  };
  const wallet = (await post('wallets', { name: 'Ví kiểm chứng F1', type: 'cash' })).wallet;
  const category = (await post('categories', { name: 'Chi F1', kind: 'expense' })).category;
  await post('transactions', { type: 'expense', source_wallet_id: wallet.id, category_id: category.id, amount_vnd: 410000, occurred_at: new Date().toISOString(), note: 'Chi thực 410 nghìn' });
  return { wallet, category, post };
}

test('overlapping budgets retain independent exact totals when choosing and reloading', async ({ page }, info) => {
  const { post, category } = await ledger(page);
  const a = (await post('budgets', { name: 'Hạn mức A', period_type: 'monthly', amount_vnd: 500000, category_ids: [category.id] })).budget;
  const b = (await post('budgets', { name: 'Hạn mức B', period_type: 'monthly', amount_vnd: 200000, category_ids: [category.id] })).budget;
  await page.reload();
  await page.getByRole('button', { name: 'Ngân sách', exact: true }).click();
  const summary = page.getByRole('region', { name: 'Ngân sách đã chọn' });
  const select = page.getByRole('combobox', { name: 'Ngân sách hiển thị' });
  await select.selectOption(a.id);
  await expect(summary).toContainText('Hạn mức A');
  await expect(summary).toContainText('90.000 đ');
  await expect(summary).not.toContainText('820.000');
  await select.selectOption(b.id);
  await expect(summary).toContainText('Hạn mức B');
  await expect(summary).toContainText('Vượt ngân sách');
  await expect(summary).toContainText('210.000 đ');
  await expect(summary).not.toContainText('820.000');
  await page.screenshot({ path: `/tmp/mypocket-so-tien-f1-budget-${info.project.name}.png`, fullPage: true });
  await page.reload();
  await page.getByRole('button', { name: 'Ngân sách', exact: true }).click();
  await select.selectOption(b.id);
  await expect(summary).toContainText('210.000 đ');
});

test('wallet detail failure is safe and explicit retry returns the real ledger', async ({ page }) => {
  const { wallet } = await ledger(page);
  await page.reload();
  let attempts = 0;
  await page.route(`**/api/v1/wallets/${wallet.id}/detail`, async route => {
    if (++attempts === 1) await route.fulfill({ status: 503, contentType: 'application/json', body: JSON.stringify({ error: { code: 'INTERNAL_RETRYABLE', message: 'PRIVATE_WALLET_DETAIL' }, correlation_id: 'req_wallet_f1' }) });
    else await route.continue();
  });
  await page.getByRole('button', { name: /Ví kiểm chứng F1/ }).click();
  const error = page.getByRole('alert').filter({ hasText: 'req_wallet_f1' });
  await expect(error).toBeVisible();
  await expect(page.getByText('PRIVATE_WALLET_DETAIL')).toHaveCount(0);
  await error.getByRole('button').click();
  const detail = page.getByLabel('Chi tiết Ví kiểm chứng F1');
  await expect(detail).toContainText('Chi thực 410 nghìn');
  expect(attempts).toBe(2);
});

test('a failed planning read does not erase successful wallet and transaction data', async ({ page }) => {
  await ledger(page);
  await page.route('**/api/v1/budgets', route => route.fulfill({ status: 503, contentType: 'application/json', body: JSON.stringify({ error: { code: 'INTERNAL_RETRYABLE' }, correlation_id: 'req_planning_f1' }) }));
  await page.reload();
  await expect(page.getByRole('button', { name: /Ví kiểm chứng F1/ })).toBeVisible();
  await page.getByRole('button', { name: 'Ngân sách', exact: true }).click();
  await expect(page.getByRole('alert').filter({ hasText: 'req_planning_f1' })).toBeVisible();
  await page.getByRole('button', { name: 'Sổ giao dịch', exact: true }).click();
  await expect(page.getByText('Chi thực 410 nghìn', { exact: true })).toBeVisible();
});
