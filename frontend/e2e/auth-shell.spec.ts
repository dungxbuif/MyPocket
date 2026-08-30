import { expect, test } from "@playwright/test";

test("fixture login persists after browser reload and logout returns to login", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("button", { name: "Đăng nhập bằng Google" })).toBeVisible();

  await page.getByRole("button", { name: "Đăng nhập bằng Google" }).click();
  await expect(page.getByText("fixture@example.com")).toBeVisible();

  await page.reload();
  await expect(page.getByText("fixture@example.com")).toBeVisible();

  await page.getByRole("button", { name: "Tài khoản" }).click();
  await page.getByRole("button", { name: "Đăng xuất" }).click();

  await expect(page.getByRole("button", { name: "Đăng nhập bằng Google" })).toBeVisible();
  await expect(page.getByText("fixture@example.com")).toHaveCount(0);
});

test("forbidden auth response renders a safe forbidden state", async ({ page }) => {
  await page.route("**/api/v1/me", async (route) => {
    await route.fulfill({
      status: 403,
      contentType: "application/json",
      body: JSON.stringify({
        status: "error",
        error: { code: "FORBIDDEN", message: "Forbidden" },
        correlation_id: "req_forbidden_e2e",
      }),
    });
  });

  await page.goto("/");

  await expect(page.getByText("Không có quyền truy cập")).toBeVisible();
  await expect(page.getByText("req_forbidden_e2e")).toBeVisible();
  await expect(page.getByRole("button", { name: "Đăng xuất" })).toBeVisible();
});
