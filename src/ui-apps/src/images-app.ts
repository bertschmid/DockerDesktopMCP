import type { App } from "@modelcontextprotocol/ext-apps";

import { escapeHtml, parseHumanBytes } from "./shared/formatting";
import { applyBaseStyles, connectApp, createAndConnectApp } from "./shared/theme";

interface ImageRow {
  id: string;
  repository: string[];
  tags: string[];
  size: string;
  created: string;
}

function extractRows(payload: unknown): ImageRow[] {
  const data = payload as { images?: ImageRow[] } | ImageRow[] | undefined;
  if (Array.isArray(data)) {
    return data;
  }
  return data?.images ?? [];
}

function render(rows: ImageRow[]): void {
  const root = document.getElementById("root");
  if (!root) {
    return;
  }

  const sorted = [...rows].sort((a, b) => parseHumanBytes(b.size) - parseHumanBytes(a.size));
  const max = Math.max(1, ...sorted.map((r) => parseHumanBytes(r.size)));

  const tableRows = sorted
    .map((row) => {
      const sizeBytes = parseHumanBytes(row.size);
      const width = Math.max(4, Math.round((sizeBytes / max) * 100));
      return `
        <tr>
          <td><code>${escapeHtml(row.id)}</code></td>
          <td>${row.repository.map((r) => `<code>${escapeHtml(r)}</code>`).join(" ")}</td>
          <td>${row.tags.map((t) => `<code>${escapeHtml(t)}</code>`).join(" ")}</td>
          <td>
            <div>${escapeHtml(row.size)}</div>
            <div style="height:8px;border-radius:4px;background:rgba(127,127,127,.25);margin-top:4px;">
              <div style="height:8px;border-radius:4px;background:#3b82f6;width:${width}%;"></div>
            </div>
          </td>
          <td>${escapeHtml(new Date(row.created).toLocaleString())}</td>
        </tr>
      `;
    })
    .join("");

  root.innerHTML = `
    <div class="card">
      <h2>Images</h2>
      <div class="muted">Total images: ${sorted.length}</div>
    </div>
    <div class="card" style="overflow:auto;">
      <table>
        <thead><tr><th>ID</th><th>Repository</th><th>Tags</th><th>Size</th><th>Created</th></tr></thead>
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
const app = createAndConnectApp("docker-images");
wireToolHandlers(app);
await connectApp(app);
