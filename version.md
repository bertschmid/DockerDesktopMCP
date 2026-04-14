# Version History

All notable changes and breaking changes are documented in this file.

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
