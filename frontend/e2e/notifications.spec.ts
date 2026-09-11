import { expect, test } from "@playwright/test";

test("enables Web Push through the mounted notification inbox and live API", async ({ page }) => {
  await page.addInitScript(() => {
    Object.defineProperty(window, "PushManager", { configurable: true, value: class PushManager {} });
    Object.defineProperty(window, "Notification", {
      configurable: true,
      value: { requestPermission: async () => "granted" },
    });
    Object.defineProperty(ServiceWorkerContainer.prototype, "ready", {
      configurable: true,
      get: () => Promise.resolve({
        pushManager: {
          subscribe: async () => ({
            toJSON: () => ({
              endpoint: `https://push.example/${crypto.randomUUID()}`,
              expirationTime: null,
              keys: { p256dh: "browser-p256dh", auth: "browser-auth" },
            }),
          }),
        },
      }),
    });
  });

  await page.goto("/");
  await page.getByRole("button", { name: "Đăng nhập bằng Google" }).click();
  await expect(page.getByRole("heading", { name: "Ví của tôi" })).toBeVisible();
  await page.getByRole("button", { name: "Thông báo" }).click();
  await page.getByRole("button", { name: "Bật Web Push" }).click();

  await expect(page.getByText("Web Push đã bật")).toBeVisible();
});
