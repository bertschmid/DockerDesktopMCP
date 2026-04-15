import type { App } from "@modelcontextprotocol/ext-apps";

import { escapeHtml } from "./shared/formatting";
import { applyBaseStyles, connectApp, createAndConnectApp } from "./shared/theme";

interface VolumeRow {
  name: string;
  driver: string;
  mountpoint: string;
  created?: string;
  scope: string;
}

function extractRows(payload: unknown): VolumeRow[] {
  const data = payload as { volumes?: VolumeRow[] } | VolumeRow[] | undefined;
  if (Array.isArray(data)) {
    return data;
  }
  return data?.volumes ?? [];
}

function render(rows: VolumeRow[]): void {
  const root = document.getElementById("root");
  if (!root) {
    return;
  }

  const tableRows = rows
    .map((row) => {
      return `
        <tr>
          <td><code>${escapeHtml(row.name)}</code></td>
          <td><span class="chip">${escapeHtml(row.driver)}</span></td>
          <td><span class="chip">${escapeHtml(row.scope)}</span></td>
          <td><code>${escapeHtml(row.mountpoint)}</code></td>
          <td>${escapeHtml(row.created ?? "-")}</td>
        </tr>
      `;
    })
    .join("");

  root.innerHTML = `
    <div class="card">
      <h2>Volumes</h2>
      <div class="muted">Total volumes: ${rows.length}</div>
    </div>
    <div class="card" style="overflow:auto;">
      <table>
        <thead><tr><th>Name</th><th>Driver</th><th>Scope</th><th>Mountpoint</th><th>Created</th></tr></thead>
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
const app = createAndConnectApp("docker-volumes");
wireToolHandlers(app);
await connectApp(app);
