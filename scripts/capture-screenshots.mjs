import { chromium } from "playwright";

const baseURL = process.env.BASE_URL ?? "http://127.0.0.1:8080";
const browser = await chromium.launch({ headless: true });
const page = await browser.newPage({
  viewport: { width: 1440, height: 1000 },
  deviceScaleFactor: 1,
});

const pageErrors = [];
const consoleErrors = [];
page.on("pageerror", (error) => pageErrors.push(error));
page.on("console", (message) => {
  if (message.type() === "error") consoleErrors.push(message.text());
});

try {
  await page.goto(baseURL, { waitUntil: "networkidle" });
  await page.locator("#pagesCount").filter({ hasText: "4" }).waitFor();
  await page.screenshot({ path: "docs/screenshots/01-dashboard.png", fullPage: true });

  await page.getByRole("button", { name: "Seiten & Namespaces" }).click();
  await page.locator("#pageTree .page-row").first().waitFor();

  // Expand the namespace hierarchy before selecting a nested page.
  const collapsedNamespaces = page.locator(
    '#pageTree .folder-row .tree-toggle[aria-label="Namespace aufklappen"]',
  );
  while (await collapsedNamespaces.count()) {
    await collapsedNamespaces.first().click();
  }
  await page.screenshot({ path: "docs/screenshots/02-pages.png", fullPage: true });

  const row = page.locator("#pageTree .page-row").filter({ hasText: "SAP Betriebshandbuch" });
  await row.waitFor();
  await row.locator('input[type="checkbox"]').check();
  await page.getByText("1 ausgewählt", { exact: true }).waitFor();
  await page.screenshot({ path: "docs/screenshots/03-selection.png", fullPage: true });

  const reviewButton = page.getByRole("button", { name: "Prüfung starten →" });
  await reviewButton.waitFor();
  await reviewButton.click();
  await page.getByRole("heading", { name: "2 · Übernahme prüfen" }).waitFor();

  const detailRow = page.locator("#pageTree .page-row").filter({ hasText: "SAP Betriebshandbuch" });
  await detailRow.click();
  const detailHeading = page.locator("#detail > h2");
  await detailHeading.waitFor();
  const detailTitle = (await detailHeading.textContent())?.trim();
  if (detailTitle !== "SAP Betriebshandbuch") {
    throw new Error(`Unexpected detail title: ${detailTitle ?? "missing"}`);
  }
  await page.getByRole("button", { name: "Migration", exact: true }).click();
  await page.screenshot({ path: "docs/screenshots/04-migration-preview.png", fullPage: true });

  const exportButton = page.getByRole("button", { name: "Export vorbereiten →" });
  await exportButton.waitFor();
  await exportButton.click();
  await page.getByRole("heading", { name: "3 · Export" }).waitFor();
  await page.screenshot({ path: "docs/screenshots/05-export.png", fullPage: true });

  await page.getByRole("button", { name: "Plugin-Inventar" }).click();
  await page.locator("#pluginRows tr").first().waitFor();
  await page.screenshot({ path: "docs/screenshots/06-plugins.png", fullPage: true });

  if (pageErrors.length || consoleErrors.length) {
    const details = [
      ...pageErrors.map((error) => error.stack ?? error.message),
      ...consoleErrors,
    ].join("\n");
    throw new Error(`Browser errors detected:\n${details}`);
  }
} catch (error) {
  await page.screenshot({ path: "test-runtime/ui-failure.png", fullPage: true });
  if (pageErrors.length || consoleErrors.length) {
    console.error("Browser errors:", [
      ...pageErrors.map((item) => item.stack ?? item.message),
      ...consoleErrors,
    ].join("\n"));
  }
  throw error;
} finally {
  await browser.close();
}
