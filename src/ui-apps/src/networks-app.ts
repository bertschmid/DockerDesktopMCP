import type { App } from "@modelcontextprotocol/ext-apps";

import { escapeHtml } from "./shared/formatting";
import { applyBaseStyles, connectApp, createAndConnectApp } from "./shared/theme";

interface NetworkRow {
  id: string;
  name: string;
  driver: string;
  scope: string;
  subnet?: string;
}

function extractRows(payload: unknown): NetworkRow[] {
  const data = payload as { networks?: NetworkRow[] } | NetworkRow[] | undefined;
  if (Array.isArray(data)) {
    return data;
  }
  return data?.networks ?? [];
}

function renderTopology(rows: NetworkRow[]): string {
  const height = Math.max(100, rows.length * 90);
  const regions = rows
    .map((n, i) => {
      const y = 10 + i * 85;
      return `
        <rect x="10" y="${y}" width="560" height="70" rx="8" fill="rgba(59,130,246,0.12)" stroke="#3b82f6" />
        <text x="24" y="${y + 24}" fill="currentColor" font-size="12">${escapeHtml(n.name)}</text>
        <text x="24" y="${y + 44}" fill="currentColor" font-size="11">${escapeHtml(n.driver)} | ${escapeHtml(n.subnet ?? "-")}</text>
      `;
    })
    .join("");
  return `<svg width="100%" height="${height}" viewBox="0 0 580 ${height}">${regions}</svg>`;
}

function render(rows: NetworkRow[]): void {
  const root = document.getElementById("root");
  if (!root) {
    return;
  }

  const tableRows = rows
    .map((row) => `
      <tr>
        <td><code>${escapeHtml(row.id)}</code></td>
        <td>${escapeHtml(row.name)}</td>
        <td>${escapeHtml(row.driver)}</td>
        <td>${escapeHtml(row.scope)}</td>
        <td>${escapeHtml(row.subnet ?? "-")}</td>
      </tr>
    `)
    .join("");

  root.innerHTML = `
    <div class="card"><h2>Networks</h2><div class="muted">Total networks: ${rows.length}</div></div>
    <div class="card">${renderTopology(rows)}</div>
    <div class="card" style="overflow:auto;">
      <table>
        <thead><tr><th>ID</th><th>Name</th><th>Driver</th><th>Scope</th><th>Subnet</th></tr></thead>
        <tbody>${tableRows}</tbody>
      </table>
    </div>
  `;
}

function wireToolHandlers(app: App): void {
  app.ontoolinput = (params: unknown) => {
    const p = params as { structuredContent?: unknown };
    render(extractRows(p.structuredContent));
  };

  app.ontoolresult = (result: unknown) => {
    const r = result as { structuredContent?: unknown };
    render(extractRows(r.structuredContent));
  };
}

applyBaseStyles();
const app = createAndConnectApp("docker-networks");
wireToolHandlers(app);
await connectApp(app);
