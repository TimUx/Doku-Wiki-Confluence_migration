import { chromium } from "playwright";

const baseURL = process.env.BASE_URL ?? "http://127.0.0.1:8080";
const browser = await chromium.launch({ headless: true });
const page = await browser.newPage({
  viewport: { width: 1440, height: 1000 },
  deviceScaleFactor: 1,
});

try {
  await page.goto(baseURL, { waitUntil: "networkidle" });
  await page.locator("#pagesCount").filter({ hasText: "4" }).waitFor();
  await page.screenshot({ path: "docs/screenshots/01-dashboard.png", fullPage: true });

  await page.getByRole("button", { name: "Seiten & Namespaces" }).click();
  await page.locator("#pageRows tr").first().waitFor();
  await page.screenshot({ path: "docs/screenshots/02-pages.png", fullPage: true });

  const row = page.locator("#pageRows tr").filter({ hasText: "SAP Betriebshandbuch" });
  await row.locator('input[type="checkbox"]').check();
  await page.screenshot({ path: "docs/screenshots/03-selection.png", fullPage: true });

  const detailRow = page.locator("#pageRows tr").filter({ hasText: "SAP Betriebshandbuch" });
  await detailRow.click();
  const detailHeading = page.locator("#detail > h2");
  await detailHeading.waitFor();
  const detailTitle = (await detailHeading.textContent())?.trim();
  if (detailTitle !== "SAP Betriebshandbuch") {
    throw new Error(`Unexpected detail title: ${detailTitle ?? "missing"}`);
  }
  await page.getByRole("button", { name: "Migration" }).click();
  await page.screenshot({ path: "docs/screenshots/04-migration-preview.png", fullPage: true });

  await page.getByRole("button", { name: "Export vorbereiten →" }).click();
  await page.locator("#wizardBody").filter({ hasText: "3 · Export" }).waitFor();
  await page.screenshot({ path: "docs/screenshots/05-export.png", fullPage: true });

  await page.getByRole("button", { name: "Plugin-Inventar" }).click();
  await page.locator("#pluginRows tr").first().waitFor();
  await page.screenshot({ path: "docs/screenshots/06-plugins.png", fullPage: true });
} catch (error) {
  await page.screenshot({ path: "test-runtime/ui-failure.png", fullPage: true });
  throw error;
} finally {
  await browser.close();
}
