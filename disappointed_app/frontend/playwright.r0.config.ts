import { defineConfig } from '@playwright/test';
import base from './playwright.config';

export default defineConfig({
  ...base,
  outputDir: 'test-results/r0',
  use: { ...base.use, baseURL: 'http://127.0.0.1:4187' },
  webServer: (base.webServer as Array<{ command: string; url: string }>).map(server => ({
    ...server,
    command: server.command.replaceAll('55433/mypocket_e2e', '64739/mypocket_r0_e2e').replaceAll('4173', '4187'),
    url: server.url.replaceAll('4173', '4187'),
    reuseExistingServer: false,
  })),
});
