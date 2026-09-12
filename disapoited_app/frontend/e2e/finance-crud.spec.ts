import { expect, test } from "@playwright/test";

test("mobile finance CRUD works through the live API", async ({ page }) => {
  const suffix = Date.now().toString().slice(-6);
  const walletName = `Ví E2E ${suffix}`;
  const categoryName = `Cafe E2E ${suffix}`;
  const originalNote = `Lương E2E ${suffix}`;
  const updatedNote = `Lương sửa ${suffix}`;

  await page.goto("/");
  await page.getByRole("button", { name: "Đăng nhập bằng Google" }).click();
  await expect(page.getByRole("heading", { name: "Ví của tôi" })).toBeVisible();

  await page.getByRole("button", { name: "Xem tất cả" }).click();
  await expect(page.getByRole("dialog", { name: "Ví Của Tôi" })).toBeVisible();
  await page.getByRole("button", { name: "Thêm ví", exact: true }).click();
  await page.getByLabel("Tên ví mới").fill(walletName);
  await page.getByRole("button", { name: "Tạo ví" }).click();
  await expect(page.getByRole("dialog").locator("strong").filter({ hasText: walletName })).toBeVisible();

  await page.getByLabel("Tên nhóm mới").fill(categoryName);
  await page.getByRole("button", { name: "Tạo nhóm" }).click();
  await expect(page.getByLabel(`Tên nhóm ${categoryName}`)).toBeVisible();
  await page.getByRole("button", { name: "Đóng" }).click();

  await page.getByRole("button", { name: "Thêm giao dịch" }).click();
  await page.getByRole("button", { name: "Thu" }).click();
  await page.getByLabel("Số tiền").fill("700000");
  await page.getByLabel("Ví nguồn").selectOption({ label: walletName });
  await page.getByLabel("Ghi chú").fill(originalNote);
  await page.getByRole("button", { name: "Lưu" }).click();

  await page.getByRole("button", { name: "Sổ giao dịch" }).click();
  await expect(page.getByText(originalNote)).toBeVisible();
  await page.getByText(originalNote).click();
  await expect(page.getByRole("dialog", { name: "Sửa Giao Dịch" })).toBeVisible();
  await page.getByLabel("Ghi chú").fill(updatedNote);
  await page.getByRole("button", { name: "Lưu thay đổi" }).click();
  await expect(page.getByRole("dialog", { name: "Sửa Giao Dịch" })).toHaveCount(0);

  await expect(page.getByText(updatedNote)).toBeVisible();
  await page.getByText(updatedNote).click();
  await page.getByRole("button", { name: "Lưu trữ" }).click();
  await expect(page.getByText(updatedNote)).toHaveCount(0);
});
