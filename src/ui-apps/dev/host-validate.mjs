import path from "node:path";
import { pathToFileURL } from "node:url";

import { chromium } from "playwright";

const hostFile = path.resolve("./dev/basic-host.html");
const hostUrl = pathToFileURL(hostFile).toString();

const browser = await chromium.launch({ headless: true });
const page = await browser.newPage();
await page.goto(hostUrl, { waitUntil: "load" });

await page.waitForFunction(() => {
  const el = document.getElementById("status");
  return !!el && (el.textContent?.startsWith("PASS") || el.textContent?.startsWith("FAIL"));
}, { timeout: 20000 });

const status = await page.locator("#status").innerText();
const log = await page.locator("#log").innerText();

console.log("STATUS:", status);
console.log("LOG_START");
console.log(log);
console.log("LOG_END");

await browser.close();

if (!status.startsWith("PASS")) {
  process.exit(1);
}
