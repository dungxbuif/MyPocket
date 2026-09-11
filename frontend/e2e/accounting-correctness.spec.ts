import { expect, test, type Page } from '@playwright/test';

async function setup(page: Page) {
  const owner = `accounting-${crypto.randomUUID()}`;
  await page.goto(`http://127.0.0.1:18173/api/v1/auth/google/callback?subject=${owner}&email=${owner}@example.com&email_verified=true&name=Accounting`);
  await expect(page.getByRole('heading', { name: 'Ví của tôi' })).toBeVisible();
  const csrf = (await page.context().cookies()).find(c => c.name === 'mypocket_csrf')!.value;
  const headers = { 'X-CSRF-Token': csrf };
  const categories = (await (await page.request.get('http://127.0.0.1:18173/api/v1/categories')).json()).categories;
  const post = async (path: string, data: object) => {
    const transaction = data as { type?: string };
    const defaults = path === 'transactions' && ['income', 'expense'].includes(transaction.type ?? '')
      ? { category_id: categories.find((category: { kind: string }) => category.kind === transaction.type).id } : {};
    const response = await page.request.post(`http://127.0.0.1:18173/api/v1/${path}`, { headers: { ...headers, 'Idempotency-Key': crypto.randomUUID() }, data: { ...defaults, ...data } });
    expect(response.status(), await response.text()).toBe(201);
    return response.json();
  };
  const a = (await post('wallets', { name: 'Ví nguồn kiểm chứng', type: 'cash' })).wallet;
  const b = (await post('wallets', { name: 'Ví nhận kiểm chứng', type: 'bank' })).wallet;
  const get = async (path: string) => {
    const response = await page.request.get(`http://127.0.0.1:18173/api/v1/${path}`);
    expect(response.status()).toBe(200);
    return response.json();
  };
  return { a, b, post, get, headers };
}

test('opening an empty wallet renders an empty state without crashing and nonempty detail shows the real amount', async ({ page }) => {
  const errors: string[] = [];
  page.on('pageerror', error => errors.push(error.message));
  const { a, post } = await setup(page);
  await page.reload();
  await page.getByRole('button', { name: /Ví nguồn kiểm chứng/ }).click();
  const detail = page.getByLabel('Chi tiết Ví nguồn kiểm chứng');
  await expect(detail.getByText('Chưa có giao dịch')).toBeVisible();
  await detail.getByRole('button', { name: 'Đóng' }).click();
  await post('transactions', { type: 'income', source_wallet_id: a.id, amount_vnd: 123456, occurred_at: '2026-09-09T05:00:00Z', note: 'Thu kiểm chứng chi tiết' });
  await page.reload();
  await page.getByRole('button', { name: /Ví nguồn kiểm chứng/ }).click();
  await expect(detail.getByText('Thu kiểm chứng chi tiết')).toBeVisible();
  await expect(detail.getByText('123.456 đ')).toHaveCount(2);
  expect(errors).toEqual([]);
});

test('transfer create edit archive moves both balances exactly and never counts as income or expense', async ({ page }) => {
  const { a, b, post, get, headers } = await setup(page);
  await post('transactions', { type: 'income', source_wallet_id: a.id, amount_vnd: 1000000, occurred_at: new Date().toISOString(), note: 'Vốn nguồn' });
  await page.reload();
  await page.getByRole('button', { name: 'Thêm giao dịch', exact: true }).click();
  await page.getByRole('button', { name: 'Chuyển ví', exact: true }).click();
  await page.getByLabel('Số tiền', { exact: true }).fill('250000');
  await page.getByLabel('Ví nguồn', { exact: true }).selectOption(a.id);
  await page.getByLabel('Ví đích', { exact: true }).selectOption(a.id);
  await expect(page.getByRole('button', { name: 'Lưu', exact: true })).toBeDisabled();
  await page.getByLabel('Ví đích', { exact: true }).selectOption(b.id);
  await page.getByLabel('Ghi chú', { exact: true }).fill('Chuyển kiểm chứng');
  await page.getByRole('button', { name: 'Lưu', exact: true }).click();
  await expect(page.getByRole('dialog', { name: 'Thêm Giao Dịch' })).toHaveCount(0);
  const balances = async () => {
    const wallets = (await get('wallets')).wallets;
    return [wallets.find((w: { id: string }) => w.id === a.id).balance_vnd, wallets.find((w: { id: string }) => w.id === b.id).balance_vnd];
  };
  await expect.poll(balances).toEqual([750000, 250000]);
  await page.getByRole('button', { name: 'Sổ giao dịch', exact: true }).click();
  await expect(page.getByRole('button', { name: /Chuyển kiểm chứng/ })).toContainText('250.000 đ');
  const transfers = (await get('transactions')).transactions.filter((t: { type: string }) => t.type === 'transfer');
  expect(transfers).toHaveLength(1);
  const denied = await page.request.post('http://127.0.0.1:18173/api/v1/transactions', { headers: { ...headers, 'Idempotency-Key': crypto.randomUUID() }, data: { type: 'transfer', source_wallet_id: a.id, destination_wallet_id: a.id, amount_vnd: 100, occurred_at: new Date().toISOString() } });
  expect(denied.status()).toBe(400);
  expect(await balances()).toEqual([750000, 250000]);
  await page.getByText('Chuyển kiểm chứng', { exact: true }).click();
  await page.getByLabel('Số tiền', { exact: true }).fill('400000');
  await page.getByRole('button', { name: 'Lưu thay đổi', exact: true }).click();
  await expect.poll(balances).toEqual([600000, 400000]);
  expect((await get('reports/cash-flow')).report.summary).toMatchObject({ income_vnd: 1000000, expense_vnd: 0, net_income_vnd: 1000000 });
  await page.getByText('Chuyển kiểm chứng', { exact: true }).click();
  await page.getByRole('button', { name: 'Lưu trữ', exact: true }).click();
  await expect.poll(balances).toEqual([1000000, 0]);
  await page.reload();
  expect((await get('transactions')).transactions.filter((t: { type: string }) => t.type === 'transfer')).toEqual([]);
});

test('report APIs and mounted report controls agree with a known ledger across timezone boundaries and exclusions', async ({ page }) => {
  const { a, b, post, get, headers } = await setup(page);
  const category = (await post('categories', { kind: 'expense', name: 'Chi kiểm chứng' })).category;
  const tx = (type: string, amount: number, date: string, extra: object = {}) => post('transactions', { type, amount_vnd: amount, source_wallet_id: a.id, occurred_at: date, ...extra });
  // Sep 1 midnight Vietnam is Aug 31 17:00Z; the transaction one second earlier is August.
  await tx('income', 999999, '2026-08-31T16:59:59Z');
  await tx('income', 1000000, '2026-08-31T17:00:00Z');
  await tx('expense', 120000, '2026-09-01T05:00:00Z', { category_id: category.id });
  await tx('expense', 80000, '2026-09-02T05:00:00Z', { category_id: category.id });
  await tx('expense', 70000, '2026-09-02T05:00:00Z', { excluded_from_reports: true });
  await tx('transfer', 250000, '2026-09-02T05:00:00Z', { destination_wallet_id: b.id });
  await tx('adjustment', 1600000, '2026-09-02T06:00:00Z', { target_balance_vnd: 1600000 });
  const archived = (await tx('expense', 50000, '2026-09-02T07:00:00Z')).transaction;
  expect((await page.request.post(`http://127.0.0.1:18173/api/v1/transactions/${archived.id}/archive`, { headers, data: { base_version: archived.version } })).status()).toBe(200);
  await tx('income', 123456, '2026-09-02T17:00:00Z'); // Sep 3, outside selected two days.
  const range = '?from=2026-09-01&to=2026-09-02';
  for (const kind of ['cash-flow', 'categories', 'daily', 'comparison', 'cumulative']) {
    const report = (await get(`reports/${kind}${range}`)).report;
    expect(report.summary).toMatchObject({ income_vnd: 1000000, expense_vnd: 200000, net_income_vnd: 800000, from: '2026-09-01', to: '2026-09-02', timezone: 'Asia/Ho_Chi_Minh' });
    expect((await get(`reports/${kind}${range}&wallet_id=${b.id}`)).report.summary).toMatchObject({ income_vnd: 0, expense_vnd: 0, net_income_vnd: 0 });
  }
  const categories = (await get(`reports/categories${range}`)).report.categories;
  expect(categories).toEqual([{ category_id: category.id, category_name: 'Chi kiểm chứng', amount_vnd: 200000, share_percent: 100 }]);
  const daily = (await get(`reports/daily${range}`)).report.daily;
  expect(daily).toEqual([
    { date: '2026-09-01', income_vnd: 1000000, expense_vnd: 120000, net_income_vnd: 880000, cumulative_net_vnd: 880000 },
    { date: '2026-09-02', income_vnd: 0, expense_vnd: 80000, net_income_vnd: -80000, cumulative_net_vnd: 800000 },
  ]);
  expect((await get(`reports/comparison${range}`)).report.prior).toMatchObject({ income_vnd: 999999, expense_vnd: 0 });
  expect((await get(`reports/insider${range}`)).report).toMatchObject({ spent_vnd: 200000, average_daily_vnd: 100000, selected_category: { name: 'Chi kiểm chứng', transaction_count: 2 } });
  const dashboard = (await get(`dashboard${range}`)).report;
  expect(dashboard.summary).toMatchObject({ income_vnd: 1000000, expense_vnd: 200000 });
  expect(dashboard.net_worth_vnd).toBe(1973456); // Live balances A=1723456, B=250000; report period is not a balance snapshot.
  await page.reload();
  await expect(page.getByRole('button', { name: 'Xem báo cáo', exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'Xem báo cáo', exact: true }).click();
  const panel = page.getByRole('region', { name: 'Báo cáo chi tiết', exact: true });
  await panel.getByLabel('Từ ngày', { exact: true }).fill('2026-09-01');
  await panel.getByLabel('Đến ngày', { exact: true }).fill('2026-09-02');
  for (const kind of ['cash-flow', 'categories', 'daily', 'comparison', 'cumulative']) {
    await panel.getByRole('combobox', { name: 'Loại báo cáo', exact: true }).selectOption(kind);
    await panel.getByRole('button', { name: 'Xem báo cáo', exact: true }).click();
    const summary = panel.getByRole('table', { name: 'Tổng hợp báo cáo' });
    await expect(summary).toContainText('1.000.000 đ');
    await expect(summary).toContainText('200.000 đ');
    await expect(summary).toContainText('800.000 đ');
  }
  await panel.getByRole('combobox', { name: 'Ví báo cáo', exact: true }).selectOption(b.id);
  await panel.getByRole('button', { name: 'Xem báo cáo', exact: true }).click();
  const summary = panel.getByRole('table', { name: 'Tổng hợp báo cáo' });
  await expect(summary).not.toContainText('1.000.000 đ');
  await expect(summary).toContainText('0 đ');
});
