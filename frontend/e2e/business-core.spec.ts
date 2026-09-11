import { expect, test, type Page } from "@playwright/test";

const API_BASE = "http://127.0.0.1:18173/api/v1";

interface WalletSummary {
  id: string;
  name: string;
  balance_vnd: number;
}

interface CashFlowSummary {
  income_vnd: number;
  expense_vnd: number;
  net_income_vnd: number;
}

type ApiEnvelope<T> = { [key: string]: T };

async function loginViaFixture(page: Page, owner: string) {
  await page.goto(`${API_BASE.replace("/api/v1", "")}/api/v1/auth/google/callback?subject=${owner}&email=${owner}@example.com&email_verified=true&name=${encodeURIComponent(owner)}`);
  await expect(page.getByRole("heading", { name: "Ví của tôi" })).toBeVisible();
}

async function openWalletManager(page: Page) {
  await page.getByRole("button", { name: "Xem tất cả" }).click();
  await expect(page.getByRole("dialog", { name: "Ví Của Tôi" })).toBeVisible();
}

async function closeWalletManager(page: Page) {
  await page.getByRole("button", { name: "Đóng" }).first().click();
  await expect(page.getByRole("dialog", { name: "Ví Của Tôi" })).toHaveCount(0);
}

async function createWallet(page: Page, name: string) {
  await page.getByRole("button", { name: "Thêm ví", exact: true }).click();
  await page.getByLabel("Tên ví mới").fill(name);
  await page.getByLabel("Loại ví").selectOption("cash");
  await page.getByRole("button", { name: "Tạo ví" }).click();
  await expect(page.getByRole("dialog").locator("strong").filter({ hasText: name })).toBeVisible();
}

async function renameWallet(page: Page, previousName: string, nextName: string) {
  await page.getByRole("button", { name: "Sửa" }).click();
  const row = page.locator(".manager-row").filter({ has: page.getByLabel(`Tên ví ${previousName}`) });
  await expect(row.getByLabel(`Tên ví ${previousName}`)).toBeVisible();
  await row.getByLabel(`Tên ví ${previousName}`).fill(nextName);
  await row.getByRole("button", { name: "Lưu", exact: true }).click();
  await expect(page.getByLabel(`Tên ví ${nextName}`)).toHaveValue(nextName);
}

async function archiveWalletByName(page: Page, walletName: string) {
  if (await page.getByRole("button", { name: "Sửa" }).count() > 0) {
    await page.getByRole("button", { name: "Sửa" }).click();
  }
  const row = page.locator(".manager-row").filter({ has: page.getByLabel(`Tên ví ${walletName}`) });
  await row.getByRole("button", { name: "Ẩn" }).click();
  await expect(row).toHaveCount(0);
}

function randomName(prefix: string) {
  return `${prefix} ${Date.now().toString().slice(-6)}-${crypto.randomUUID().slice(0, 4)}`;
}

async function addTransaction(page: Page, typeLabel: "Khoản thu" | "Khoản chi" | "Vay/nợ" | "Chuyển ví", params: {
  amount: number;
  sourceWalletName: string;
  note: string;
  destinationWalletName?: string;
}) {
  await page.getByRole("button", { name: "Thêm giao dịch", exact: true }).click();
  await page.getByRole("button", { name: typeLabel }).click();
  await page.getByLabel("Số tiền").fill(String(params.amount));
  await page.getByLabel("Ví nguồn").selectOption({ label: params.sourceWalletName });
  if (typeLabel === "Chuyển ví") {
    expect(params.destinationWalletName).toBeTruthy();
    await page.getByLabel("Ví đích").selectOption({ label: params.destinationWalletName! });
  }
  await page.getByLabel("Ghi chú").fill(params.note);
  await page.getByRole("button", { name: "Lưu", exact: true }).click();
  await expect(page.getByRole("dialog", { name: "Thêm Giao Dịch" })).toHaveCount(0);
}

async function openTransactionByNote(page: Page, note: string) {
  await page.getByRole("button", { name: "Sổ giao dịch", exact: true }).click();
  await page.getByText(note).click();
  await expect(page.getByRole("dialog", { name: "Sửa Giao Dịch" })).toBeVisible();
}

async function editTransaction(page: Page, previousNote: string, next: { note?: string; amount?: number }) {
  await openTransactionByNote(page, previousNote);
  if (next.amount !== undefined) await page.getByLabel("Số tiền", { exact: true }).fill(String(next.amount));
  if (next.note) await page.getByLabel("Ghi chú").fill(next.note);
  await page.getByRole("button", { name: "Lưu thay đổi", exact: true }).click();
  await expect(page.getByRole("dialog", { name: "Sửa Giao Dịch" })).toHaveCount(0);
}

async function archiveTransaction(page: Page, note: string) {
  await openTransactionByNote(page, note);
  await page.getByRole("button", { name: "Lưu trữ", exact: true }).click();
  await expect(page.getByRole("dialog", { name: "Sửa Giao Dịch" })).toHaveCount(0);
}

async function loadWallets(page: Page): Promise<WalletSummary[]> {
  const response = await page.request.get(`${API_BASE}/wallets`);
  expect(response.status()).toBe(200);
  return (await response.json() as ApiEnvelope<{ wallets: WalletSummary[] }>).wallets;
}

async function walletBalance(page: Page, name: string) {
  const wallets = await loadWallets(page);
  const match = wallets.find((wallet) => wallet.name === name);
  expect(match).toBeTruthy();
  return match!.balance_vnd;
}

async function cashFlowSummary(page: Page): Promise<CashFlowSummary> {
  const response = await page.request.get(`${API_BASE}/reports/cash-flow`);
  expect(response.status()).toBe(200);
  return (await response.json() as ApiEnvelope<{ report: { summary: CashFlowSummary } }>).report.summary;
}

function expectTransferExcludedFromCashFlow(summary: CashFlowSummary) {
  expect(summary).toMatchObject({
    income_vnd: 1_000_000,
    expense_vnd: 0,
    net_income_vnd: 1_000_000,
  });
}

test("core finance workflow: tạo/sửa/ẩn ví", async ({ page }) => {
  const owner = randomName("owner-wallet-core");
  const originalName = randomName("Ví nghiệp vụ");
  const updatedName = `${originalName} - Đã đổi tên`;
  await loginViaFixture(page, owner);
  await openWalletManager(page);
  await createWallet(page, originalName);
  await expect(page.getByRole("dialog").locator("strong").filter({ hasText: originalName })).toBeVisible();
  await renameWallet(page, originalName, updatedName);
  await archiveWalletByName(page, updatedName);
  await closeWalletManager(page);
});

test("core transaction workflow: tạo/sửa/ẩn giao dịch và chuyển tiền giữa ví", async ({ page }) => {
  const owner = randomName("owner-txn-core");
  const walletA = randomName("Ví nguồn");
  const walletB = randomName("Ví đích");
  const txIncome = randomName("Thu ban đầu");
  const txExpense = randomName("Chi cập nhật");
  const txTransfer = randomName("Chuyển nghiệp vụ");
  const txTransferUpdated = `${txTransfer} - sửa`;
  await loginViaFixture(page, owner);

  await openWalletManager(page);
  await createWallet(page, walletA);
  await createWallet(page, walletB);
  await closeWalletManager(page);

  await addTransaction(page, "Khoản thu", {
    amount: 1_000_000,
    sourceWalletName: walletA,
    note: txIncome,
  });
  await addTransaction(page, "Khoản chi", {
    amount: 120_000,
    sourceWalletName: walletA,
    note: txExpense,
  });

  await editTransaction(page, txExpense, { note: txExpense, amount: 150_000 });
  await expect.poll(() => walletBalance(page, walletA)).toBe(850_000);

  await archiveTransaction(page, txExpense);
  await expect(page.getByText(txExpense)).toHaveCount(0);
  await expect.poll(() => walletBalance(page, walletA)).toBe(1_000_000);

  const reportBeforeTransfer = await cashFlowSummary(page);
  expectTransferExcludedFromCashFlow(reportBeforeTransfer);

  await addTransaction(page, "Chuyển ví", {
    amount: 250_000,
    sourceWalletName: walletA,
    destinationWalletName: walletB,
    note: txTransfer,
  });
  await expect.poll(async () => Promise.all([walletBalance(page, walletA), walletBalance(page, walletB)])).toEqual([750_000, 250_000]);
  expectTransferExcludedFromCashFlow(await cashFlowSummary(page));

  await editTransaction(page, txTransfer, { note: txTransferUpdated, amount: 300_000 });
  await expect.poll(async () => Promise.all([walletBalance(page, walletA), walletBalance(page, walletB)])).toEqual([700_000, 300_000]);
  expectTransferExcludedFromCashFlow(await cashFlowSummary(page));

  await archiveTransaction(page, txTransferUpdated);
  await expect.poll(async () => Promise.all([walletBalance(page, walletA), walletBalance(page, walletB)])).toEqual([1_000_000, 0]);
  expectTransferExcludedFromCashFlow(await cashFlowSummary(page));
});
