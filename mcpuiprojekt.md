# MCP Apps Project Plan for DockerDesktopMCP

**Decision:** The implementation scope is strictly MCP Apps only.

The current goal is to add rich UI for Docker list-style tools so MCP-capable hosts can render
interactive tables, charts, and diagrams directly from tool calls. The required work is limited
to two parts:

1. `ui-apps` project under `src/` — builds self-contained HTML apps
2. Go backend — exposes MCP resources and returns `_meta.ui.resourceUri` plus `structuredContent`

No second dashboard surface is part of this plan.

---

## Required Scope

- Add a dedicated `ui-apps` project under `src/` for MCP Apps HTML UIs
- Add MCP `resources/list` and `resources/read` support in Go
- Attach UI resources only to the relevant list tools
- Keep plain text fallback output for non-UI hosts
- Test with an MCP Apps host such as `basic-host`

## Not In Scope

- Any additional dashboard surface outside the MCP Apps flow
- Duplicate tables or charts for the same Docker data
- A separate preview UI as part of this task

---

## Why This Scope Is Correct

The MCP Apps flow already provides the required UI delivery path:

- The model calls a tool such as `docker_container_list`
- The tool returns text output for compatibility
- The tool also returns `structuredContent`
- The tool also returns `_meta.ui.resourceUri`
- The host fetches the HTML app through `resources/read`
- The host renders the app in a sandboxed iframe

That is the actual product path for MCP UI. Building an extra local dashboard would duplicate the
same logic and create an unnecessary second UI surface to maintain.

---

## Target Tools

These tools should get MCP Apps UI because their outputs are structured and benefit from tables,
charts, or diagrams.

| Tool | Category | UI value |
|---|---|---|
| `docker_container_list` | Containers | High — status table, filters, badges |
| `docker_image_list` | Images | High — size ranking, repository/tag grouping |
| `docker_volume_list` | Volumes | Medium — clean list with scope/driver badges |
| `docker_network_list` | Networks | High — table + topology diagram |
| `docker_compose_ps` | Compose | High — grouped services by project and status |
| `docker_system_df` | System | High — disk usage chart |
| `docker_system_info` | System | Medium — KPI cards / summary panel |

Mutation tools such as `docker_container_start`, `docker_container_stop`, `docker_image_remove`,
and similar commands remain plain tools. The HTML apps may call them through UI actions where it
makes sense, but they do not need their own dedicated UI resources.

---

## Architecture

```
BUILD TIME
─────────────────────────────────────────────────────────────
ui-apps project under src/
  Vite + @modelcontextprotocol/ext-apps
        │
        │ npm run build
        ▼
  dist/containers.html
  dist/images.html
  dist/volumes.html
  dist/networks.html
  dist/compose.html
  dist/disk-usage.html
  dist/system-info.html
        │
        │ copied before go build
        ▼
backend/ui-apps-dist/
        │
        │ //go:embed
        ▼
Go MCP server binary

RUNTIME
─────────────────────────────────────────────────────────────
LLM / MCP Host
  │
  │ tools/call → docker_container_list
  ▼
Go MCP server
  │
  │ returns:
  │ - content (text fallback)
  │ - structuredContent
  │ - _meta.ui.resourceUri
  ▼
Host detects UI resource
  │
  │ resources/read → ui://docker-desktop/containers
  ▼
Embedded HTML app
  │
  ▼
Sandboxed iframe rendered by MCP host
```

---

## SDK Roles

| Layer | Role |
|---|---|
| `@modelcontextprotocol/ext-apps` | Runs inside each HTML app |
| Go backend | Registers resources, serves embedded HTML, returns structured tool output |

> Important: the HTML should not be generated dynamically in Go. The official MCP Apps flow is
> to pre-build the app HTML and serve it as a resource.

---

## Phase 1 — Create the `ui-apps` Project

This is the only frontend work required for the MCP Apps feature.

### 1.1 Project structure

```
ui-apps project under src/
├── package.json
├── vite.config.ts
├── tsconfig.json
├── containers.html
├── images.html
├── volumes.html
├── networks.html
├── compose.html
├── disk-usage.html
├── system-info.html
└── src/
    ├── containers-app.ts
    ├── images-app.ts
    ├── volumes-app.ts
    ├── networks-app.ts
    ├── compose-app.ts
    ├── disk-usage-app.ts
    ├── system-info-app.ts
    └── shared/
        ├── theme.ts
        ├── table.ts
        ├── svg.ts
        └── formatting.ts
```

### 1.2 Dependencies

```bash
cd src
cd ui-apps
npm install @modelcontextprotocol/ext-apps
npm install -D vite vite-plugin-singlefile typescript
```

Rules from the MCP Apps skill:

- Use `npm install`, not handwritten versions
- Use `vite-plugin-singlefile`
- Build self-contained HTML files with no external assets

### 1.3 Vite config

```ts
import { defineConfig } from "vite";
import { viteSingleFile } from "vite-plugin-singlefile";

export default defineConfig({
  plugins: [viteSingleFile()],
  build: {
    outDir: "dist",
    rollupOptions: {
      input: {
        containers: "containers.html",
        images: "images.html",
        volumes: "volumes.html",
        networks: "networks.html",
        compose: "compose.html",
        "disk-usage": "disk-usage.html",
        "system-info": "system-info.html",
      },
    },
  },
});
```

### 1.4 HTML entry pattern

Each UI gets its own HTML file:

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Docker Containers</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="./src/containers-app.ts"></script>
  </body>
</html>
```

### 1.5 App bootstrap pattern

Every app follows the same lifecycle:

```ts
import {
  App,
  PostMessageTransport,
  applyDocumentTheme,
  applyHostStyleVariables,
  applyHostFonts,
} from "@modelcontextprotocol/ext-apps";

const app = new App({ name: "docker-containers", version: "1.0.0" });

app.ontoolinput = (params) => {
  render(params.structuredContent);
};

app.ontoolresult = () => {
  updateAfterResult();
};

app.onhostcontextchanged = (ctx) => {
  if (ctx.theme) applyDocumentTheme(ctx.theme);
  if (ctx.styles?.variables) applyHostStyleVariables(ctx.styles.variables);
  if (ctx.styles?.css?.fonts) applyHostFonts(ctx.styles.css.fonts);
  if (ctx.safeAreaInsets) {
    const { top, right, bottom, left } = ctx.safeAreaInsets;
    document.body.style.padding = `${top}px ${right}px ${bottom}px ${left}px`;
  }
};

app.onteardown = async () => ({});

await app.connect(new PostMessageTransport());
```

Critical rule: register all handlers before `app.connect()`.

---

## Phase 2 — Build the Tool UIs

### Containers app

Purpose:
- status table
- coloured state badges
- optional action buttons for start / stop / restart

Expected `structuredContent` shape:

```json
{
  "containers": [
    {
      "id": "abc123",
      "names": ["/web"],
      "image": "nginx:latest",
      "state": "running",
      "status": "Up 2 hours",
      "ports": ["80/tcp"],
      "created": "2026-04-15T08:00:00Z"
    }
  ]
}
```

### Images app

Purpose:
- image table
- size bars
- repository/tag grouping

### Volumes app

Purpose:
- clean list with badges for driver and scope

### Networks app

Purpose:
- network table
- SVG topology diagram

### Compose app

Purpose:
- group services by compose project
- show service state summary per project

### Disk usage app

Purpose:
- SVG pie or donut chart for images / containers / volumes / build cache

### System info app

Purpose:
- KPI cards for CPU, memory, OS, kernel, Docker version

---

## Diagram and Visualization Rules

### Network topology

Use inline SVG. Networks are rendered as grouped regions, containers as nodes.

```svg
<svg width="600" height="300">
  <rect x="10" y="10" width="580" height="90" rx="8" />
  <text x="20" y="30">bridge</text>
  <circle cx="120" cy="60" r="18" />
  <text x="120" y="64" text-anchor="middle">nginx</text>
</svg>
```

### Disk usage chart

Use pure SVG or a small library that Vite can inline. Do not rely on a CDN.

### Status badges

Use CSS classes driven by host variables:

```css
.badge-running { background: var(--color-background-success); }
.badge-paused { background: var(--color-background-warning); }
.badge-stopped { background: var(--color-background-secondary); }
```

---

## Phase 3 — Go Backend MCP Apps Support

The backend work is the second required half.

### 3.1 Add resource methods

Add MCP support for:

- `resources/list`
- `resources/read`

### 3.2 Resource registry

Create `src/backend/internal/mcp/resources.go` with a static registry:

```go
var uiResources = []registeredResource{
    {"ui://docker-desktop/containers", "text/html;profile=mcp-app", "Containers List", "ui-apps-dist/containers.html"},
    {"ui://docker-desktop/images", "text/html;profile=mcp-app", "Images List", "ui-apps-dist/images.html"},
    {"ui://docker-desktop/volumes", "text/html;profile=mcp-app", "Volumes List", "ui-apps-dist/volumes.html"},
    {"ui://docker-desktop/networks", "text/html;profile=mcp-app", "Networks List", "ui-apps-dist/networks.html"},
    {"ui://docker-desktop/compose-services", "text/html;profile=mcp-app", "Compose Services", "ui-apps-dist/compose.html"},
    {"ui://docker-desktop/disk-usage", "text/html;profile=mcp-app", "Disk Usage", "ui-apps-dist/disk-usage.html"},
    {"ui://docker-desktop/system-info", "text/html;profile=mcp-app", "System Info", "ui-apps-dist/system-info.html"},
}
```

### 3.3 Use `go:embed`

Embed the built HTML files:

```go
//go:embed ../../../ui-apps-dist/*.html
var uiAppsFS embed.FS
```

### 3.4 `resources/read`

Serve the embedded HTML directly. No runtime templating.

### 3.5 Extend tool result model

`CallToolResult` needs:

```go
type CallToolResult struct {
    Content           []ContentItem  `json:"content"`
    StructuredContent map[string]any `json:"structuredContent,omitempty"`
    Meta              *ToolMeta      `json:"_meta,omitempty"`
    IsError           bool           `json:"isError,omitempty"`
}
```

### 3.6 Add `_meta.ui.resourceUri`

Relevant list tools should return both text fallback and UI metadata.

Example for `docker_container_list`:

```json
{
  "content": [{ "type": "text", "text": "[...]" }],
  "structuredContent": { "containers": [...] },
  "_meta": { "ui": { "resourceUri": "ui://docker-desktop/containers" } }
}
```

### 3.7 Files to change

| File | Change |
|---|---|
| `src/backend/internal/result/result.go` | add `StructuredContent` and `Meta` |
| `src/backend/internal/mcp/protocol.go` | add resource protocol types |
| `src/backend/internal/mcp/server.go` | route `resources/list` and `resources/read` |
| `src/backend/internal/mcp/resources.go` | new file for embedded UI resources |
| `src/backend/internal/docker/containers.go` | add `structuredContent` and `_meta` |
| `src/backend/internal/docker/images.go` | add `structuredContent` and `_meta` |
| `src/backend/internal/docker/volumes.go` | add `structuredContent` and `_meta` |
| `src/backend/internal/docker/networks.go` | add `structuredContent` and `_meta` |
| `src/backend/internal/docker/compose.go` | add `structuredContent` and `_meta` |
| `src/backend/internal/docker/system.go` | add `structuredContent` and `_meta` |

---

## Phase 4 — Dockerfile / Build Integration

The build must include the new UI project before the Go binary is compiled.

### Required Dockerfile change

Add a Node builder stage for the UI apps and copy the resulting `dist/` files into the Go build
context before `go build`.

Example direction:

```dockerfile
FROM node:22-alpine AS ui-apps-builder
WORKDIR /app
COPY ui-apps/package*.json ./
RUN npm ci
COPY ui-apps/ .
RUN npm run build

FROM golang:1.24-alpine AS go-builder
WORKDIR /app
COPY --from=ui-apps-builder /app/dist ./backend/ui-apps-dist
COPY src/backend/ ./backend/
RUN cd backend && go build -o /mcp-server ./...
```

---

## Common Mistakes to Avoid

1. Forgetting the text `content` fallback. UI is an enhancement, not a replacement.
2. Registering app handlers after `app.connect()`.
3. Omitting `vite-plugin-singlefile`.
4. Returning `_meta.ui.resourceUri` without registering the matching resource.
5. Loading fonts, scripts, or charts from a CDN inside the iframe.
6. Hardcoding styles instead of using host variables.
7. Generating HTML dynamically in Go.
8. Expanding scope beyond the MCP Apps path before it is finished.

---

## Execution Order

```
Step 1   Read Go + TS instructions

── Core frontend: ui-apps ────────────────────────────────────────────────────
Step 2   Create package.json
Step 3   Create vite.config.ts
Step 4   Create tsconfig.json
Step 5   Create HTML entry files
Step 6   Create shared helpers
Step 7   Implement containers app
Step 8   Implement images app
Step 9   Implement volumes app
Step 10  Implement networks app
Step 11  Implement compose app
Step 12  Implement disk-usage app
Step 13  Implement system-info app
Step 14  Build ui-apps and verify dist/*.html

── Backend integration ───────────────────────────────────────────────────────
Step 15  Extend result model
Step 16  Add MCP resource protocol types
Step 17  Add resources.go with embed + handlers
Step 18  Update server routing
Step 19  Attach structuredContent + resourceUri to list tools
Step 20  Update Dockerfile build stages
Step 21  Run Go tests

── Validation ────────────────────────────────────────────────────────────────
Step 22  Test with basic-host
Step 23  Verify resources/list and resources/read
Step 24  Verify tool UIs render correctly
```

---

## Testing

### Core validation

```bash
cd src
cd ui-apps
npm run build

cd ..
make dev
```

### Host validation with `basic-host`

```bash
git clone --branch "v$(npm view @modelcontextprotocol/ext-apps version)" --depth 1 \
  https://github.com/modelcontextprotocol/ext-apps.git /tmp/mcp-ext-apps

cd /tmp/mcp-ext-apps/examples/basic-host
npm install
SERVERS='["http://localhost:3282/mcp"]' npm run start
```

Verify:

1. Plain list tools still return text output
2. `resources/list` returns all UI resources
3. `resources/read` returns valid HTML for each URI
4. Hosts render the HTML apps in an iframe
5. Apps receive `structuredContent` in `ontoolinput`
6. Theme variables from the host are applied

---

## Version Impact

This is a product-relevant change, so a patch version bump is required when implementation starts.

Keep the repository's release version files synchronized according to the repository instructions.

---

## Final Scope Statement

This plan is intentionally limited to MCP Apps resource delivery and the required backend
integration. No additional UI surface is part of the task.
