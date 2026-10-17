const { defineConfig } = require('@playwright/test');

module.exports = defineConfig({
  testDir: './journeys',
  timeout: 150_000,
  expect: { timeout: 15_000 },
  workers: 1,
  retries: 0,
  outputDir: 'artifacts/playwright',
  reporter: [['list'], ['html', { outputFolder: 'artifacts/playwright-report', open: 'never' }]],
  use: {
    browserName: 'chromium',
    viewport: { width: 1440, height: 1100 },
    launchOptions: {
      executablePath: process.env.CHROMIUM_PATH || '/usr/bin/chromium',
      args: ['--no-sandbox'],
    },
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },
});
