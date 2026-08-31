import { expect, test } from "@playwright/test";

test("mobile budget CRUD shows threshold progress through the live API", async ({ page }) => {
  const suffix = Date.now().toString().slice(-6);
  const walletName = `Ví budget ${suffix}`;
  const categoryName = `Budget cafe ${suffix}`;
  const note = `Chi budget ${suffix}`;
  const budgetName = `Ngân sách cafe ${suffix}`;

  await page.goto("/");
  await page.getByRole("button", { name: "Đăng nhập bằng Google" }).click();
  await expect(page.getByText("fixture@example.com")).toBeVisible();

  await page.getByRole("button", { name: "Xem tất cả" }).click();
  await page.getByLabel("Tên ví mới").fill(walletName);
  await page.getByRole("button", { name: "Tạo ví" }).click();
  await expect(page.getByLabel(`Tên ví ${walletName}`)).toBeVisible();
  await page.getByLabel("Tên nhóm mới").fill(categoryName);
  await page.getByRole("button", { name: "Tạo nhóm" }).click();
  await expect(page.getByLabel(`Tên nhóm ${categoryName}`)).toBeVisible();
  await page.getByRole("button", { name: "Đóng" }).click();

  await page.getByRole("button", { name: "Thêm giao dịch" }).click();
  await page.getByLabel("Số tiền").fill("410000");
  await page.getByLabel("Ví nguồn").selectOption({ label: walletName });
  await page.getByLabel("Nhóm").selectOption({ label: categoryName });
  await page.getByLabel("Ghi chú").fill(note);
  await page.getByRole("button", { name: "Lưu", exact: true }).click();
  await expect(page.getByRole("dialog", { name: "Thêm Giao Dịch" })).toHaveCount(0);

  await page.getByRole("button", { name: "Ngân sách" }).click();
  await page.getByRole("button", { name: "Tạo", exact: true }).click();
  await page.getByLabel("Tên ngân sách").fill(budgetName);
  await page.getByLabel("Số tiền ngân sách").fill("500000");
  await page.getByRole("button", { name: "Tất cả nhóm chi" }).click();
  await page.getByRole("button", { name: categoryName }).click();
  await page.getByRole("button", { name: "Lưu", exact: true }).click();

  const budgetRow = page.getByRole("button", { name: new RegExp(budgetName) });
  await expect(budgetRow).toBeVisible();
  await expect(budgetRow.getByText("Đã chạm 80%")).toBeVisible();
  await budgetRow.click();
  await page.getByLabel("Số tiền ngân sách").fill("600000");
  await page.getByRole("button", { name: "Lưu", exact: true }).click();
  await expect(page.getByRole("button", { name: new RegExp(`${budgetName}.*600\\.000`) })).toBeVisible();
  await page.getByRole("button", { name: new RegExp(budgetName) }).click();
  await page.getByRole("button", { name: "Lưu trữ" }).click();
  await expect(page.getByText(budgetName)).toHaveCount(0);
});
