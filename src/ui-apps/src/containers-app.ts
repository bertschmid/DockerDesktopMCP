import type { App } from "@modelcontextprotocol/ext-apps";

import { escapeHtml, formatDate } from "./shared/formatting";
import { applyBaseStyles, connectApp, createAndConnectApp } from "./shared/theme";

interface ContainerRow {
  id: string;
  names: string[];
  image: string;
  state: string;
  status: string;
  ports: string[];
  created: string;
}

function extractRows(payload: unknown): ContainerRow[] {
  const data = payload as { containers?: ContainerRow[] } | ContainerRow[] | undefined;
  if (Array.isArray(data)) {
    return data;
  }
  return data?.containers ?? [];
}

function actionButton(toolName: string, id: string, label: string): string {
  return `<button data-tool="${toolName}" data-id="${escapeHtml(id)}">${escapeHtml(label)}</button>`;
}

function render(rows: ContainerRow[]): void {
  const root = document.getElementById("root");
  if (!root) {
    return;
  }

  const running = rows.filter((r) => r.state === "running").length;
  const paused = rows.filter((r) => r.state === "paused").length;
  const stopped = rows.length - running - paused;

  const tableRows = rows
    .map((row) => {
      const names = row.names.map((n) => `<code>${escapeHtml(n)}</code>`).join(" ");
      const ports = row.ports.map((p) => `<code>${escapeHtml(p)}</code>`).join(" ");
      const stateClass = `chip chip-${escapeHtml(row.state.toLowerCase())}`;
      return `
        <tr>
          <td><code>${escapeHtml(row.id)}</code></td>
          <td>${names}</td>
          <td><code>${escapeHtml(row.image)}</code></td>
          <td><span class="${stateClass}">${escapeHtml(row.state)}</span></td>
          <td>${escapeHtml(row.status)}</td>
          <td>${ports}</td>
          <td>${escapeHtml(formatDate(row.created))}</td>
          <td class="actions">
            ${actionButton("docker_container_start", row.id, "Start")}
            ${actionButton("docker_container_stop", row.id, "Stop")}
            ${actionButton("docker_container_restart", row.id, "Restart")}
          </td>
        </tr>
      `;
    })
    .join("");

  root.innerHTML = `
    <div class="card">
      <h2>Containers</h2>
      <div class="muted">Running: ${running} | Paused: ${paused} | Stopped: ${stopped}</div>
    </div>
    <div class="card" style="overflow:auto;">
      <table>
        <thead>
          <tr>
            <th>ID</th><th>Names</th><th>Image</th><th>State</th><th>Status</th><th>Ports</th><th>Created</th><th>Actions</th>
          </tr>
        </thead>
        <tbody>${tableRows}</tbody>
      </table>
    </div>
  `;

  root.querySelectorAll<HTMLButtonElement>("button[data-tool]").forEach((btn) => {
    btn.addEventListener("click", () => {
      const tool = btn.dataset.tool ?? "";
      const id = btn.dataset.id ?? "";
      window.parent.postMessage(
        {
          type: "tool",
          payload: {
            toolName: tool,
            params: { id }
          }
        },
        globalThis.location.origin
      );
    });
  });
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
const app = createAndConnectApp("docker-containers");
wireToolHandlers(app);
await connectApp(app);
