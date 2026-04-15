import type { App } from "@modelcontextprotocol/ext-apps";

import { escapeHtml } from "./shared/formatting";
import { applyBaseStyles, connectApp, createAndConnectApp } from "./shared/theme";

interface ComposeServiceRow {
  name: string;
  state: string;
  status?: string;
  project?: string;
}

function extractRows(payload: unknown): ComposeServiceRow[] {
  const data = payload as { services?: ComposeServiceRow[] } | ComposeServiceRow[] | undefined;
  if (Array.isArray(data)) {
    return data;
  }
  return data?.services ?? [];
}

function render(rows: ComposeServiceRow[]): void {
  const root = document.getElementById("root");
  if (!root) {
    return;
  }

  const byProject = new Map<string, ComposeServiceRow[]>();
  for (const row of rows) {
    const p = row.project || "default";
    byProject.set(p, [...(byProject.get(p) ?? []), row]);
  }

  const projects = Array.from(byProject.entries())
    .map(([project, services]) => {
      const chips = services
        .map((s) => {
          const cls = `chip chip-${(s.state || "unknown").toLowerCase()}`;
          return `<span class="${cls}">${escapeHtml(s.name)}: ${escapeHtml(s.state || "unknown")}</span>`;
        })
        .join(" ");
      return `<div class="card"><h3>${escapeHtml(project)}</h3><div>${chips || "-"}</div></div>`;
    })
    .join("");

  root.innerHTML = `
    <div class="card"><h2>Compose Services</h2><div class="muted">Services: ${rows.length}</div></div>
    ${projects || "<div class=\"card muted\">No compose services found. Ensure project_dir is set.</div>"}
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
const app = createAndConnectApp("docker-compose");
wireToolHandlers(app);
await connectApp(app);
