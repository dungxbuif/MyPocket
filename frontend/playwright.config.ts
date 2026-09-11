import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  use: { baseURL: 'http://127.0.0.1:4173', trace: 'retain-on-failure' },
  webServer: [
    {
      command: 'curl -sf -X PUT http://127.0.0.1:4567/mypocket >/dev/null || true; cd ../backend && DATABASE_URL="postgres://mypocket:mypocket@127.0.0.1:55433/mypocket_e2e?sslmode=disable" go run ./cmd/migrate && APP_ENV=test DATABASE_URL="postgres://mypocket:mypocket@127.0.0.1:55433/mypocket_e2e?sslmode=disable" PUBLIC_WEB_URL="http://127.0.0.1:4173" COOKIE_SECRET="change-this-development-cookie-secret-32-bytes" CSRF_SECRET="change-this-development-csrf-secret-32-bytes" API_KEY_HASH_SECRET="change-this-test-api-key-hash-secret-32-bytes" S3_ENDPOINT="http://127.0.0.1:4567" S3_BUCKET="mypocket" S3_ACCESS_KEY="test" S3_SECRET_KEY="test" HTTP_ADDR="127.0.0.1:18173" OAUTH_FIXTURE_MODE=true go run ./cmd/api',
      url: 'http://127.0.0.1:18173/api/v1/health/live',
      reuseExistingServer: false,
      timeout: 30_000,
    },
    {
      command: 'VITE_API_BASE_URL="http://127.0.0.1:18173" VITE_WEB_PUSH_PUBLIC_KEY="test-public-key" npm run build && npm run preview -- --host 0.0.0.0 --port 4173',
      url: 'http://127.0.0.1:4173',
      reuseExistingServer: false,
      timeout: 30_000,
    },
  ],
  projects: [
    { name: 'mobile', use: { ...devices['Pixel 5'] } },
    { name: 'desktop', use: { ...devices['Desktop Chrome'] } },
    { name: 'webkit-mobile', use: { ...devices['iPhone 13'] } },
  ],
});
