# Copilot Instructions for DockerDesktopMCP

## Project Overview

**DockerDesktopMCP** is a Docker Desktop Extension that runs a local HTTP server implementing the [Model Context Protocol (MCP)](https://spec.modelcontextprotocol.io/specification/2025-03-26/). It enables any MCP-compatible AI assistant (e.g. Claude) to manage Docker containers, images, volumes, networks, and Compose stacks through natural language.

- **MCP endpoint**: `http://localhost:3282/mcp`
- **Health endpoint**: `http://localhost:3282/health`
- **35 MCP tools** covering the full Docker Desktop API

## Repository Structure

```
DockerDesktopMCP/
├── .github/
│   └── copilot-instructions.md   # This file
├── src/
│   ├── Dockerfile                # Multi-stage build (Node → Go → Alpine runtime)
│   ├── docker-compose.yaml       # Extension service definition
│   ├── Makefile                  # Developer shortcuts
│   ├── metadata.json             # Docker Desktop extension manifest
│   ├── backend/                  # Go MCP server
│   │   ├── main.go
│   │   └── internal/
│   │       ├── docker/           # Docker API wrappers
│   │       └── mcp/              # MCP protocol & tool dispatch
│   └── ui/                       # React dashboard (TypeScript)
├── version.md                    # Version history and breaking changes
├── README.md
└── LICENSE
```

## Versioning Rules

- The version is defined in `src/Makefile` as `VERSION ?= <major>.<minor>.<patch>`
- **With every change, the build number (patch) must be incremented.**
- For breaking changes, increment the **minor** version and reset the patch to `0`.
- For major API/architecture overhauls, increment the **major** version.
- Always update `version.md` in the root folder when changing the version, documenting what changed and any breaking changes.

## Build & Development Commands

All commands are run from the `src/` directory:

| Command | Description |
|---|---|
| `make build` | Build the extension Docker image |
| `make install` | Install into Docker Desktop |
| `make update` | Update an already-installed extension |
| `make uninstall` | Uninstall the extension |
| `make dev` | Run the Go backend directly |
| `make dev-tls` | Run with auto-generated TLS |
| `make ui-dev` | Start React UI dev server (hot-reload) |
| `make test` | Run Go tests |
| `make tidy` | Tidy Go module dependencies |
| `make validate-local` | Run Docker's validator on local image |
| `make validate-release RELEASE_IMAGE=...` | Build, push, and validate multi-arch release |

## Technology Stack

- **Backend**: Go (MCP HTTP server, Docker API client)
- **Frontend**: React + TypeScript (Docker Desktop dashboard tab)
- **Runtime**: Alpine Linux (Docker container via Docker Desktop Extension)
- **Protocol**: Model Context Protocol (MCP) over HTTP

## Coding Conventions

- Go code lives in `src/backend/`; follow standard Go conventions and `gofmt` formatting.
- React/TypeScript code lives in `src/ui/`; use `npm run build` to compile.
- Do not commit secrets or credentials.
- Keep `README.md` and `version.md` in sync with any user-facing changes.
- The Docker image must support both `linux/amd64` and `linux/arm64` platforms.

## MCP Tool Categories

| Category | Count | Prefix |
|---|---|---|
| Containers | 10 | `docker_container_` |
| Images | 6 | `docker_image_` |
| Volumes | 4 | `docker_volume_` |
| Networks | 6 | `docker_network_` |
| Compose | 5 | `docker_compose_` |
| System | 4 | `docker_system_` |
