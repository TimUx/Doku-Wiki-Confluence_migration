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
  await page.screenshot({ path: "docs/screenshots/dashboard.png", fullPage: true });

  await page.getByRole("button", { name: "Seiten & Namespaces" }).click();
  await page.locator("#pageRows tr").first().waitFor();
  await page.locator('#pageRows tr[data-id="betrieb:sap:handbuch"]').click();
  await page.locator("#detail > h2").filter({ hasText: "Überschrift" }).waitFor();
  await page.screenshot({ path: "docs/screenshots/page-detail.png", fullPage: true });

  await page.getByRole("button", { name: "Plugin-Inventar" }).click();
  await page.locator("#pluginRows tr").first().waitFor();
  await page.screenshot({ path: "docs/screenshots/plugins.png", fullPage: true });
} finally {
  await browser.close();
}
