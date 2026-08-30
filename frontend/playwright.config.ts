import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  use: { baseURL: 'http://localhost:4173', trace: 'retain-on-failure' },
  webServer: [
    {
      command: 'cd ../backend && DATABASE_URL="postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable" go run ./cmd/migrate && APP_ENV=test DATABASE_URL="postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable" PUBLIC_WEB_URL="http://localhost:4173" COOKIE_SECRET="change-this-development-cookie-secret-32-bytes" CSRF_SECRET="change-this-development-csrf-secret-32-bytes" HTTP_ADDR=":18173" OAUTH_FIXTURE_MODE=true go run ./cmd/api',
      url: 'http://localhost:18173/api/v1/health/live',
      reuseExistingServer: true,
      timeout: 30_000,
    },
    {
      command: 'VITE_API_BASE_URL="http://localhost:18173" npm run build && npm run preview -- --host localhost --port 4173',
      url: 'http://localhost:4173',
      reuseExistingServer: true,
      timeout: 30_000,
    },
  ],
  projects: [{ name: 'mobile', use: { ...devices['Pixel 5'] } }],
});
