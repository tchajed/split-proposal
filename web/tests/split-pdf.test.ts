import { test, expect } from '@playwright/test';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

test.describe('PDF splitting', () => {
	test('loads wasm module', async ({ page }) => {
		await page.goto('/');

		// Wait for WASM to be ready (drop zone becomes enabled)
		await expect(page.locator('.drop-zone:not(.disabled)')).toBeVisible({ timeout: 10000 });
	});

	test('drag-drop PDF and verify download links', async ({ page }) => {
		await page.goto('/');

		// Wait for WASM to be ready (drop zone becomes enabled)
		await expect(page.locator('.drop-zone:not(.disabled)')).toBeVisible({ timeout: 10000 });

		// Upload the sample PDF via file input
		const samplePdfPath = path.resolve(__dirname, '../../sample/main.pdf');
		const fileInput = page.locator('input[type="file"]');
		await fileInput.setInputFiles(samplePdfPath);

		// Wait for processing to complete - results container should appear
		await expect(page.locator('.bg-green-50')).toBeVisible({ timeout: 15000 });

		// Check for main download link (the zip file)
		const mainDownloadLink = page.locator('a.btn-success');
		await expect(mainDownloadLink).toBeVisible();
		await expect(mainDownloadLink).toHaveAttribute('download', 'main.zip');

		// Check for submit-summary.pdf download link
		const summaryLink = page.locator('a', { hasText: 'submit-summary.pdf' });
		await expect(summaryLink).toBeVisible();

		// Check for submit-mentoring-plan.pdf download link
		const mentoringLink = page.locator('a', { hasText: 'submit-mentoring-plan.pdf' });
		await expect(mentoringLink).toBeVisible();
	});

	test('timestamp checkbox adds date to download filename', async ({ page }) => {
		await page.goto('/');

		// Wait for WASM to be ready
		await expect(page.locator('.drop-zone:not(.disabled)')).toBeVisible({ timeout: 10000 });

		// Upload the sample PDF
		const samplePdfPath = path.resolve(__dirname, '../../sample/main.pdf');
		const fileInput = page.locator('input[type="file"]');
		await fileInput.setInputFiles(samplePdfPath);

		// Wait for processing to complete
		await expect(page.locator('.bg-green-50')).toBeVisible({ timeout: 15000 });

		// Verify initial download name doesn't have timestamp
		const mainDownloadLink = page.locator('a.btn-success');
		await expect(mainDownloadLink).toHaveAttribute('download', 'main.zip');

		// Check the "Add timestamp" checkbox
		const timestampCheckbox = page.getByLabel('Add timestamp');
		await timestampCheckbox.check();

		// Verify download name now includes a date (format: yyyy-MM-dd HH_mm)
		const downloadAttr = await mainDownloadLink.getAttribute('download');
		expect(downloadAttr).toMatch(/main \d{4}-\d{2}-\d{2} \d{2}_\d{2}\.zip/);
	});

	test('can change download filename', async ({ page }) => {
		await page.goto('/');

		// Wait for WASM to be ready
		await expect(page.locator('.drop-zone:not(.disabled)')).toBeVisible({ timeout: 10000 });

		// Upload the sample PDF
		const samplePdfPath = path.resolve(__dirname, '../../sample/main.pdf');
		const fileInput = page.locator('input[type="file"]');
		await fileInput.setInputFiles(samplePdfPath);

		// Wait for processing to complete
		await expect(page.locator('.bg-green-50')).toBeVisible({ timeout: 15000 });

		// Verify initial download name
		const mainDownloadLink = page.locator('a.btn-success');
		await expect(mainDownloadLink).toHaveAttribute('download', 'main.zip');

		// Change the filename
		const filenameInput = page.locator('#zip-name');
		await filenameInput.fill('my-proposal');

		// Verify download name is updated
		await expect(mainDownloadLink).toHaveAttribute('download', 'my-proposal.zip');
	});
});
