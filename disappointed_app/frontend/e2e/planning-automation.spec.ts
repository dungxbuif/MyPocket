import { expect, test } from "@playwright/test";

test("mobile budget CRUD shows threshold progress through the live API", async ({ page }) => {
  const suffix = Date.now().toString().slice(-6);
  const walletName = `Ví budget ${suffix}`;
  const categoryName = `Budget cafe ${suffix}`;
  const note = `Chi budget ${suffix}`;
  const budgetName = `Ngân sách cafe ${suffix}`;

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

  await page.getByRole("button", { name: "Ngân sách" }).click();
  await page.getByRole("button", { name: "Tạo", exact: true }).click();
  await page.getByLabel("Tên ngân sách").fill(budgetName);
  await page.getByLabel("Số tiền ngân sách").fill("500000");
  await page.getByRole("button", { name: "Tất cả nhóm chi" }).click();
  await page.getByRole("button", { name: categoryName }).click();
  await page.getByRole("button", { name: "Lưu", exact: true }).click();

  await page.getByRole("button", { name: "Thêm giao dịch" }).click();
  await page.getByLabel("Số tiền").fill("410000");
  await page.getByLabel("Ví nguồn").selectOption({ label: walletName });
  await page.getByLabel("Nhóm").selectOption({ label: categoryName });
  await page.getByLabel("Ngân sách giao dịch").selectOption({ label: budgetName });
  await page.getByLabel("Ghi chú").fill(note);
  await page.getByRole("button", { name: "Lưu", exact: true }).click();
  await expect(page.getByRole("dialog", { name: "Thêm Giao Dịch" })).toHaveCount(0);

  await page.reload();
  await page.getByRole("button", { name: "Ngân sách" }).click();
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

test("mobile event and debt planning links existing transactions", async ({ page }) => {
  const suffix = Date.now().toString().slice(-6);
  const walletName = `Ví plan ${suffix}`;
  const categoryName = `Plan cafe ${suffix}`;
  const note = `Khoản plan ${suffix}`;
  const eventName = `Sự kiện ${suffix}`;
  const counterparty = `Bạn ${suffix}`;
  const scheduleName = `Lịch ${suffix}`;

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
  await page.getByLabel("Số tiền").fill("200000");
  await page.getByLabel("Ví nguồn").selectOption({ label: walletName });
  await page.getByLabel("Nhóm").selectOption({ label: categoryName });
  await page.getByLabel("Ghi chú").fill(note);
  await page.getByRole("button", { name: "Lưu", exact: true }).click();
  await expect(page.getByRole("dialog", { name: "Thêm Giao Dịch" })).toHaveCount(0);

  await page.getByRole("button", { name: "Ngân sách" }).click();
  await page.getByRole("button", { name: "Tạo sự kiện" }).click();
  await page.getByLabel("Tên sự kiện").fill(eventName);
  await page.getByLabel("Giao dịch sự kiện").selectOption({ label: `${note} · 200.000 đ` });
  await page.getByRole("button", { name: "Lưu", exact: true }).click();
  await expect(page.getByRole("button", { name: new RegExp(`${eventName}[\\s\\S]*200\\.000`) })).toBeVisible();

  await page.getByRole("button", { name: "Tạo khoản nợ" }).click();
  await page.getByLabel("Đối tác").fill(counterparty);
  await page.getByLabel("Số tiền gốc").fill("500000");
  await page.getByLabel("Giao dịch trả nợ").selectOption({ label: `${note} · 200.000 đ` });
  await page.getByRole("button", { name: "Lưu", exact: true }).click();
  await expect(page.getByRole("button", { name: new RegExp(`${counterparty}[\\s\\S]*300\\.000`) })).toBeVisible();

  await page.getByRole("button", { name: "Tạo lịch" }).click();
  await page.getByLabel("Tên lịch lặp").fill(scheduleName);
  await page.getByLabel("Số tiền lịch lặp").fill("99000");
  await page.getByLabel("Ví lịch lặp").selectOption({ label: walletName });
  await page.getByLabel("Nhóm lịch lặp").selectOption({ label: categoryName });
  await page.getByLabel("Cách ghi lịch lặp").selectOption("auto_post");
  await page.getByLabel("Ngày kết thúc lịch lặp").fill("2026-12-31");
  await page.getByRole("button", { name: "Lưu", exact: true }).click();
  await expect(page.getByRole("button", { name: new RegExp(`${scheduleName}[\\s\\S]*99\\.000`) })).toBeVisible();
  await page.getByRole("button", { name: new RegExp(scheduleName) }).click();
  await page.getByLabel("Số tiền lịch lặp").fill("100000");
  await page.getByRole("button", { name: "Lưu thay đổi" }).click();
  await expect(page.getByRole("button", { name: new RegExp(`${scheduleName}[\\s\\S]*100\\.000`) })).toBeVisible();
  await page.getByRole("button", { name: new RegExp(scheduleName) }).click();
  await page.getByRole("button", { name: "Tạm dừng" }).click();
  await expect(page.getByRole("button", { name: new RegExp(scheduleName) })).toBeVisible();
});
