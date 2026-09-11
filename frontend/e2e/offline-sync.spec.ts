import { expect, test } from "@playwright/test";
import { randomUUID } from "node:crypto";

for (const loseFirstResponse of [false, true]) {
test.describe(loseFirstResponse ? "transport failure injection" : "normal reconnect", () => {
  // Service-worker-owned requests bypass page.route in WebKit. This scenario
  // tests the API response boundary; normal reconnect retains the real worker.
  if (loseFirstResponse) test.use({ serviceWorkers: "block" });
test(`offline transaction syncs once after reconnect and survives reload${loseFirstResponse ? " after losing committed response" : ""}`, async ({ page, context }) => {
  const suffix = randomUUID().slice(0, 8);
  const walletName = `Ví offline ${suffix}`;
  const categoryName = `Nhóm offline ${suffix}`;
  const note = `Offline sync ${suffix}`;

  await page.goto("/");
  await page.getByRole("button", { name: "Đăng nhập bằng Google" }).click();
  await expect(page.getByRole("heading", { name: "Ví của tôi" })).toBeVisible();

  await page.getByRole("button", { name: "Xem tất cả" }).click();
  await page.getByRole("button", { name: "Thêm ví", exact: true }).click();
  await page.getByLabel("Tên ví mới").fill(walletName);
  await page.getByRole("button", { name: "Tạo ví" }).click();
  await expect(page.getByRole("dialog").locator("strong").filter({ hasText: walletName })).toBeVisible();
  await page.getByLabel("Tên nhóm mới").fill(categoryName);
  await page.getByRole("button", { name: "Tạo nhóm" }).click();
  await expect(page.getByLabel(`Tên nhóm ${categoryName}`)).toBeVisible();
  await page.getByRole("button", { name: "Đóng" }).click();

  await context.setOffline(true);
  await page.evaluate(() => window.dispatchEvent(new Event("offline")));
  await expect(page.getByText("Offline", { exact: true })).toBeVisible();
  await page.getByRole("button", { name: "Thêm giao dịch" }).click();
  await page.getByLabel("Số tiền").fill("123000");
  await page.getByLabel("Ví nguồn").selectOption({ label: walletName });
  await page.getByLabel("Nhóm").selectOption({ label: categoryName });
  await page.getByLabel("Ghi chú").fill(note);
  await page.getByRole("button", { name: "Lưu" }).click();
  await expect(page.getByRole("dialog", { name: "Thêm Giao Dịch" })).toHaveCount(0);
  await expect(page.getByText("1 chờ đồng bộ")).toBeVisible();
  await page.getByRole("button", { name: "Sổ giao dịch" }).click();
  await expect(page.getByText(note)).toBeVisible();

  let committedResponseLost = false;
  if (loseFirstResponse) {
    await page.route("**/api/v1/sync/mutations", async (route) => {
      const response = await route.fetch();
      expect(response.status()).toBe(200);
      if (!committedResponseLost) {
        committedResponseLost = true;
        await route.abort("failed");
      } else {
        await route.fulfill({ response });
      }
    });
  }
  await context.setOffline(false);
  await page.evaluate(() => window.dispatchEvent(new Event("online")));
  if (loseFirstResponse) {
    await expect.poll(() => committedResponseLost).toBe(true);
    // Allow the response error and refresh to settle before triggering replay.
    await expect(page.getByText("1 chờ đồng bộ")).toBeVisible();
    await page.reload();
  }
  await expect(page.getByText("1 chờ đồng bộ")).toHaveCount(0);
  await page.reload();
  await page.getByRole("button", { name: "Sổ giao dịch" }).click();
  await expect(page.getByText(note)).toBeVisible();
  const transactions = await page.request.get("http://127.0.0.1:18173/api/v1/transactions");
  expect(transactions.ok()).toBe(true);
  const matching = ((await transactions.json()).transactions ?? []).filter((tx: { note: string }) => tx.note === note);
  expect(matching).toHaveLength(1);
  const wallets = await page.request.get("http://127.0.0.1:18173/api/v1/wallets");
  const savedWallet = ((await wallets.json()).wallets ?? []).find((wallet: { name: string }) => wallet.name === walletName);
  expect(savedWallet.balance_vnd).toBe(-123000);
});
});
}

test("stale offline edit creates a conflict inbox entry that can keep server", async ({ page, context, browser }) => {
  const suffix = randomUUID().slice(0, 8);
  const walletName = `Ví conflict ${suffix}`;
  const categoryName = `Nhóm conflict ${suffix}`;
  const baseNote = `Base conflict ${suffix}`;
  const serverNote = `Server conflict ${suffix}`;
  const offlineNote = `Offline conflict ${suffix}`;

  await page.goto("/");
  await page.getByRole("button", { name: "Đăng nhập bằng Google" }).click();
  await expect(page.getByRole("heading", { name: "Ví của tôi" })).toBeVisible();

  await page.getByRole("button", { name: "Xem tất cả" }).click();
  await page.getByRole("button", { name: "Thêm ví", exact: true }).click();
  await page.getByLabel("Tên ví mới").fill(walletName);
  await page.getByRole("button", { name: "Tạo ví" }).click();
  await expect(page.getByRole("dialog").locator("strong").filter({ hasText: walletName })).toBeVisible();
  await page.getByLabel("Tên nhóm mới").fill(categoryName);
  await page.getByRole("button", { name: "Tạo nhóm" }).click();
  await expect(page.getByLabel(`Tên nhóm ${categoryName}`)).toBeVisible();
  await page.getByRole("button", { name: "Đóng" }).click();

  await page.getByRole("button", { name: "Thêm giao dịch" }).click();
  await page.getByLabel("Số tiền").fill("91000");
  await page.getByLabel("Ví nguồn").selectOption({ label: walletName });
  await page.getByLabel("Nhóm").selectOption({ label: categoryName });
  await page.getByLabel("Ghi chú").fill(baseNote);
  await page.getByRole("button", { name: "Lưu" }).click();
  await expect(page.getByRole("dialog", { name: "Thêm Giao Dịch" })).toHaveCount(0);
  await page.getByRole("button", { name: "Sổ giao dịch" }).click();
  await expect(page.getByText(baseNote)).toBeVisible();

  const secondContext = await browser.newContext();
  const secondPage = await secondContext.newPage();
  await secondPage.goto("/");
  await secondPage.getByRole("button", { name: "Đăng nhập bằng Google" }).click();
  await expect(secondPage.getByRole("heading", { name: "Ví của tôi" })).toBeVisible();
  await secondPage.getByRole("button", { name: "Sổ giao dịch" }).click();
  await secondPage.getByText(baseNote).click();
  await secondPage.getByLabel("Ghi chú").fill(serverNote);
  await secondPage.getByRole("button", { name: "Lưu thay đổi" }).click();
  await expect(secondPage.getByRole("dialog", { name: "Sửa Giao Dịch" })).toHaveCount(0);
  await secondContext.close();

  await context.setOffline(true);
  await page.evaluate(() => window.dispatchEvent(new Event("offline")));
  await page.getByText(baseNote).click();
  await page.getByLabel("Ghi chú").fill(offlineNote);
  await page.getByRole("button", { name: "Lưu thay đổi" }).click();
  await expect(page.getByRole("dialog", { name: "Sửa Giao Dịch" })).toHaveCount(0);
  await expect(page.getByText("1 chờ đồng bộ")).toBeVisible();

  await context.setOffline(false);
  await page.evaluate(() => window.dispatchEvent(new Event("online")));
  await expect(page.getByLabel("Xung đột đồng bộ")).toBeVisible();
  await expect(page.getByText(`Offline: ${offlineNote}`)).toBeVisible();
  await page.getByRole("button", { name: "Giữ server" }).click();
  await expect(page.getByLabel("Xung đột đồng bộ")).toHaveCount(0);
  await page.getByRole("button", { name: "Sổ giao dịch" }).click();
  await expect(page.getByText(serverNote)).toBeVisible();
});
