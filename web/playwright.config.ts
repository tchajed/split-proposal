import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
	testDir: './tests',
	globalSetup: './tests/global-setup.ts',
	fullyParallel: true,
	forbidOnly: !!process.env.CI,
	retries: process.env.CI ? 2 : 0,
	workers: process.env.CI ? 1 : undefined,
	reporter: 'html',
	use: {
		baseURL: 'http://localhost:4173',
		trace: 'on-first-retry'
	},
	projects: [
		{
			name: 'chromium',
			use: { ...devices['Desktop Chrome'] }
		},
		{
			name: 'webkit',
			use: { ...devices['Desktop Safari'] }
		}
	],
	webServer: {
		command: 'bun run build && bun run preview',
		url: 'http://localhost:4173',
		reuseExistingServer: !process.env.CI
	}
});
