import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: true,
  retries: 0,
  reporter: 'list',
  use: { baseURL: 'http://127.0.0.1:5173', trace: 'retain-on-failure', screenshot: 'only-on-failure' },
  projects: [
    { name: 'desktop', use: { ...devices['Desktop Chrome'], viewport: { width: 1440, height: 1000 }, ...(process.platform === 'win32' ? { channel: 'msedge' } : {}) } },
    { name: 'mobile', use: { ...devices['iPhone 13'], defaultBrowserType: 'chromium', ...(process.platform === 'win32' ? { channel: 'msedge' } : {}) } },
  ],
  webServer: process.env.FUNTIME_EXTERNAL_SERVER ? undefined : { command: 'node scripts/dev.mjs', url: 'http://127.0.0.1:5173', reuseExistingServer: !process.env.CI, timeout: 120_000 },
});
