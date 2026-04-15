import type { App } from "@modelcontextprotocol/ext-apps";

import { parseHumanBytes } from "./shared/formatting";
import { renderDonut, type PieSlice } from "./shared/svg";
import { applyBaseStyles, connectApp, createAndConnectApp } from "./shared/theme";

interface DiskUsage {
  images?: { total_size?: string; count?: number };
  containers?: { rw_size?: string; count?: number };
  volumes?: { total_size?: string; count?: number };
  build_cache?: { total_size?: string; count?: number };
  total_reclaimable?: string;
}

function extractData(payload: unknown): DiskUsage {
  const data = payload as { disk_usage?: DiskUsage } | DiskUsage | undefined;
  if (!data) {
    return {};
  }
  if ("disk_usage" in data && data.disk_usage) {
    return data.disk_usage;
  }
  return data as DiskUsage;
}

function render(usage: DiskUsage): void {
  const root = document.getElementById("root");
  if (!root) {
    return;
  }

  const slices: PieSlice[] = [
    { label: "Images", value: parseHumanBytes(usage.images?.total_size ?? "0 B"), color: "#3b82f6" },
    { label: "Containers", value: parseHumanBytes(usage.containers?.rw_size ?? "0 B"), color: "#10b981" },
    { label: "Volumes", value: parseHumanBytes(usage.volumes?.total_size ?? "0 B"), color: "#f59e0b" },
    { label: "Build Cache", value: parseHumanBytes(usage.build_cache?.total_size ?? "0 B"), color: "#ef4444" }
  ];

  const legend = slices
    .map((s, i) => {
      const raw = [usage.images?.total_size, usage.containers?.rw_size, usage.volumes?.total_size, usage.build_cache?.total_size][i] ?? "0 B";
      return `<div><span style="display:inline-block;width:10px;height:10px;background:${s.color};border-radius:999px;margin-right:6px;"></span>${s.label}: ${raw}</div>`;
    })
    .join("");

  root.innerHTML = `
    <div class="card"><h2>Disk Usage</h2><div class="muted">Total reclaimable: ${usage.total_reclaimable ?? "0 B"}</div></div>
    <div class="card" style="display:flex;gap:16px;align-items:center;flex-wrap:wrap;">
      ${renderDonut(slices)}
      <div>${legend}</div>
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
const app = createAndConnectApp("docker-disk-usage");
wireToolHandlers(app);
await connectApp(app);
