import type { App } from "@modelcontextprotocol/ext-apps";

import { applyBaseStyles, connectApp, createAndConnectApp } from "./shared/theme";

interface SystemInfo {
  server_version?: string;
  os?: string;
  os_type?: string;
  architecture?: string;
  kernel_version?: string;
  cpus?: number;
  memory?: string;
  containers_running?: number;
  containers_paused?: number;
  containers_stopped?: number;
  images?: number;
  storage_driver?: string;
}

function extractData(payload: unknown): SystemInfo {
  const data = payload as { info?: SystemInfo } | SystemInfo | undefined;
  if (!data) {
    return {};
  }
  if ("info" in data && data.info) {
    return data.info;
  }
  return data as SystemInfo;
}

function tile(label: string, value: string): string {
  return `<div class="card" style="min-width:180px;flex:1;"><div class="muted">${label}</div><div><strong>${value}</strong></div></div>`;
}

function render(info: SystemInfo): void {
  const root = document.getElementById("root");
  if (!root) {
    return;
  }

  root.innerHTML = `
    <div class="card"><h2>System Info</h2><div class="muted">Docker server overview</div></div>
    <div style="display:flex;flex-wrap:wrap;gap:10px;">
      ${tile("Docker", info.server_version ?? "-")}
      ${tile("OS", info.os ?? "-")}
      ${tile("Kernel", info.kernel_version ?? "-")}
      ${tile("Architecture", info.architecture ?? "-")}
      ${tile("CPUs", String(info.cpus ?? 0))}
      ${tile("Memory", info.memory ?? "-")}
      ${tile("Images", String(info.images ?? 0))}
      ${tile("Storage Driver", info.storage_driver ?? "-")}
      ${tile("Running", String(info.containers_running ?? 0))}
      ${tile("Paused", String(info.containers_paused ?? 0))}
      ${tile("Stopped", String(info.containers_stopped ?? 0))}
    </div>
  `;
}

function wireToolHandlers(app: App): void {
  app.ontoolinput = (params: unknown) => {
    const p = params as { structuredContent?: unknown };
    render(extractData(p.structuredContent));
  };

  app.ontoolresult = (result: unknown) => {
    const r = result as { structuredContent?: unknown };
    render(extractData(r.structuredContent));
  };
}

applyBaseStyles();
const app = createAndConnectApp("docker-system-info");
wireToolHandlers(app);
await connectApp(app);
