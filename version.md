# Version History

All notable changes and breaking changes are documented in this file.

---

## [1.0.13] - 2026-04-15

### Changed
- Expanded MCP tool and parameter descriptions for container, image, volume, network, and compose tools.
- Added explicit boolean effects (`true` vs `false`) and concrete argument examples in tool schemas to improve AI prompt-to-tool precision.

### Breaking Changes
- None.

---

## [1.0.12] - 2026-04-15

### Changed
- Added concrete `filters` usage examples directly in all `docker_system_prune_*` tool descriptions to improve AI parameter selection quality.

### Breaking Changes
- None.

---

## [1.0.11] - 2026-04-15

### Changed
- Added an optional `filters` string array parameter to all `docker_system_prune_*` tools, allowing targeted cleanup with Docker prune filter expressions such as `label=...`, `label!=...`, or `until=...`.
- Refactored repeated MCP registry descriptions, UI metadata keys, and JSON header literals into constants/helpers to clear the remaining backend quality warnings.

### Breaking Changes
- None.

---

## [1.0.10] - 2026-04-15

### Changed
- Split the former `docker_system_prune` cleanup into dedicated tools for containers, images, networks, build cache, and volumes.
- Renamed the combined cleanup tool to `docker_system_prune_all` and removed the previous `all` and `volumes` parameters.
- Expanded system tool descriptions in the MCP tool registry so AI clients can better understand cleanup scope and diagnostic use cases.

### Breaking Changes
- `docker_system_prune` was renamed to `docker_system_prune_all`.

---

## [1.0.9] - 2026-04-15

### Changed
- Fixed disk usage chart rendering in MCP Apps UI by supporting IEC size units (`KiB`, `MiB`, `GiB`, `TiB`) in byte parsing.

### Breaking Changes
- None.

---

## [1.0.8] - 2026-04-15

### Changed
- Added tool-level MCP Apps metadata (`_meta.ui.resourceUri` and `ui/resourceUri`) to iframe-capable list and system tools in `tools/list`.
- Improved Claude Desktop compatibility so hosts can resolve and render MCP iframe resources directly from tool definitions.

### Breaking Changes
- None.

---

## [1.0.7] - 2026-04-15

### Changed
- Added MCP Apps resource support to the backend with `resources/list` and `resources/read` methods.
- Added a dedicated `src/ui-apps` Vite project that builds single-file HTML apps for list-oriented tools.
- Embedded generated UI app assets into the backend binary via `go:embed` and served them through MCP resources.
- Extended tool results with `structuredContent` and `_meta.ui.resourceUri` for UI-capable hosts.
- Added MCP Apps metadata and structured payloads for:
  - `docker_container_list`
  - `docker_image_list`
  - `docker_volume_list`
  - `docker_network_list`
  - `docker_compose_ps`
  - `docker_system_info`
  - `docker_system_df`
- Updated Dockerfile build pipeline to build and copy `ui-apps` artifacts before Go compilation.

### Breaking Changes
- None.

---

## [1.0.6] - 2026-04-14

### Changed
- **Fixed `docker_container_exec` output:** The exec attach stream is now properly demultiplexed using `stdcopy.StdCopy`, eliminating garbled binary frame headers in the output.
- **Added `workdir` parameter to `docker_container_exec`:** Callers can now specify the working directory inside the container for the executed command.

### Breaking Changes
- None.

---

## [1.0.5] - 2026-04-14

### Changed
- Updated Node.js build stage from `node:22-alpine` to `node:24-alpine`.

### Breaking Changes
- None.

---

## [1.0.4] - 2026-04-14

### Changed
- Added **Version Files** section to `.github/copilot-instructions.md`, listing all files that contain the project version number and must be kept in sync on every release.
- Aligned version number across all version-bearing files to `1.0.4`:
  - `src/Makefile` (authoritative)
  - `src/ui/package.json` (was `1.0.2`)
  - `src/backend/internal/mcp/server.go` health endpoint (was `1.0.0`)
  - `version.md`

### Breaking Changes
- None.

---

## [1.0.3] - 2026-04-14

### Changed
- Added **Agent Workflow** section to `.github/copilot-instructions.md`.
  Every agent must now follow a five-step process: Analyse → Plan → Execute (sub-agents, no nesting) → Verify (repeat if needed) → Version.

### Breaking Changes
- None.

---

## [1.0.2] - 2026-04-14

### Changed
- **UI build tool migrated from Create React App (`react-scripts`) to Vite 8.**
  Create React App is deprecated and produced build warnings; Vite provides a modern, warning-free build.
- **All UI npm dependencies updated to latest versions:**
  - `react` / `react-dom`: 17 → 19
  - `@mui/material` / `@mui/icons-material`: v5 → v9
  - `@emotion/react` / `@emotion/styled`: 11.11.x → 11.14.x
  - `typescript`: 4.9 → 6.0
  - `@types/react` / `@types/react-dom`: 17 → 19
  - Removed `react-scripts`, `ajv` (no longer needed with Vite)
  - Added `vite`, `@vitejs/plugin-react`
- **Docker builder image upgraded:** Node 18 (EOL) → Node 22 LTS.
- **UI build output directory changed:** `build/` → `dist/` (Vite default).
- **React 19 API:** `ReactDOM.render` replaced with `createRoot` in `index.tsx`.
- **MUI v9 breaking changes resolved in `App.tsx`:**
  - `Grid` `item`/`xs`/`md` props replaced with `size={{ xs, md }}` (new Grid2-style API).
  - `Typography paragraph` prop replaced with `sx={{ mb: 2 }}` (removed in v9).
  - Icon renames: `CheckCircleOutline` → `CheckCircleOutlined`, `ErrorOutline` → `ErrorOutlined`.
- Added `.gitignore` files to prevent accidental `node_modules` commits.
- Removed `--legacy-peer-deps` flag from npm install (no longer needed).

### Breaking Changes
- None for end users. The extension behaviour is identical.

---

## [1.0.1] - 2026-04-14

### Changed
- Internal version bump; no breaking changes.
- Added `.github/copilot-instructions.md` with project description, coding conventions, and versioning rules.
- Added `version.md` (this file) to track version history and breaking changes.

---

## [1.0.0] - Initial Release

### Added
- Docker Desktop Extension with a React UI dashboard tab.
- Local HTTP MCP server on port `3282` implementing the [Model Context Protocol](https://spec.modelcontextprotocol.io/specification/2025-03-26/).
- **35 MCP tools** across six categories:
  - **Containers (10):** `docker_container_list`, `docker_container_inspect`, `docker_container_create`, `docker_container_start`, `docker_container_stop`, `docker_container_restart`, `docker_container_remove`, `docker_container_logs`, `docker_container_exec`, `docker_container_stats`
  - **Images (6):** `docker_image_list`, `docker_image_pull`, `docker_image_build`, `docker_image_tag`, `docker_image_inspect`, `docker_image_remove`
  - **Volumes (4):** `docker_volume_list`, `docker_volume_create`, `docker_volume_inspect`, `docker_volume_remove`
  - **Networks (6):** `docker_network_list`, `docker_network_create`, `docker_network_inspect`, `docker_network_connect`, `docker_network_disconnect`, `docker_network_remove`
  - **Compose (5):** `docker_compose_up`, `docker_compose_down`, `docker_compose_ps`, `docker_compose_logs`, `docker_compose_pull`
  - **System (4):** `docker_system_info`, `docker_system_version`, `docker_system_df`, `docker_system_prune`
- Health endpoint at `http://localhost:3282/health`.
- Optional TLS support with auto-generated self-signed certificate.
- Multi-arch Docker image (`linux/amd64`, `linux/arm64`).
- Makefile with full developer workflow (`build`, `install`, `update`, `uninstall`, `dev`, `test`, `validate-local`, `validate-release`).

### Breaking Changes
- None (initial release).

---

> **Versioning policy:** The build number (patch) is incremented with every change. Minor version is incremented for breaking changes (patch resets to 0). Major version is incremented for major API or architecture overhauls.
