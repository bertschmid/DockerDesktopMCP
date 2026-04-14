# Docker Desktop MCP Server

> A **Docker Desktop Extension** that starts a local HTTP server implementing the [Model Context Protocol (MCP)](https://spec.modelcontextprotocol.io/specification/2025-03-26/). Connect any MCP-compatible AI assistant (Claude, etc.) to manage containers, images, volumes, networks, and Compose stacks through natural language.

---

## Features

- **35 MCP tools** covering the full Docker Desktop API
- Works as a **Docker Desktop Extension** — install in one command
- Dashboard tab in Docker Desktop UI
- Health endpoint at `/health`

---

## Prerequisites

| Requirement | Minimum version |
|---|---|
| [Docker Desktop](https://www.docker.com/products/docker-desktop/) | 4.8+ |
| Docker Desktop **Extensions** feature | enabled |

No other tools are required to install or run the extension.

---

## Installation

### One-liner (pre-built image)

```bash
docker extension install docker.io/bertschmid/docker-mcp-extension:latest
```

### Build from source

```bash
# 1. Clone the repository
git clone https://github.com/bertschmid/DockerDesktopMCP.git
cd DockerDesktopMCP/src

# 2. Build the extension image (React UI + Go backend)
make build

# 3. Install into Docker Desktop
make install
```

After installation the MCP server starts automatically and is reachable at:

- **MCP endpoint** — `http://localhost:3282/mcp`
- **Health check** — `http://localhost:3282/health`

### Update an existing installation

```bash
make update
```

### Validate before publishing

Docker Desktop's publication checks expect a pushed multi-arch image for both `linux/amd64` and `linux/arm64`.

For a local diagnostic run against the current image:

```bash
make validate-local
```

This still reports the multi-arch publication failure when the image only exists locally.

For a publish-grade validation:

```bash
make validate-release RELEASE_IMAGE=docker.io/<owner>/docker-mcp-extension:1.0.0
```

### Uninstall

```bash
make uninstall
```

---

## Connecting an AI assistant

Configure your MCP client to use the HTTP transport:

```json
{
  "mcpServers": {
    "docker": {
      "url": "http://localhost:3282/mcp"
    }
  }
}
```

> **Claude Desktop** example (`claude_desktop_config.json`):
> ```json
> {
>   "mcpServers": {
>     "docker": {
>       "command": "curl",
>       "args": ["-s", "-X", "POST", "http://localhost:3282/mcp"]
>     }
>   }
> }
> ```
> For best results use a native HTTP-transport MCP client.

---

## Available MCP Tools

### Containers (10)

| Tool | Description |
|---|---|
| `docker_container_list` | List all containers |
| `docker_container_inspect` | Detailed info about a container |
| `docker_container_create` | Create a new container |
| `docker_container_start` | Start a container |
| `docker_container_stop` | Stop a container |
| `docker_container_restart` | Restart a container |
| `docker_container_remove` | Remove a container |
| `docker_container_logs` | Fetch container logs |
| `docker_container_exec` | Execute a command inside a container |
| `docker_container_stats` | Live resource usage statistics |

### Images (6)

| Tool | Description |
|---|---|
| `docker_image_list` | List local images |
| `docker_image_pull` | Pull an image from a registry |
| `docker_image_build` | Build an image from a Dockerfile |
| `docker_image_tag` | Tag an image |
| `docker_image_inspect` | Detailed image metadata |
| `docker_image_remove` | Remove a local image |

### Volumes (4)

| Tool | Description |
|---|---|
| `docker_volume_list` | List volumes |
| `docker_volume_create` | Create a volume |
| `docker_volume_inspect` | Inspect a volume |
| `docker_volume_remove` | Remove a volume |

### Networks (6)

| Tool | Description |
|---|---|
| `docker_network_list` | List networks |
| `docker_network_create` | Create a network |
| `docker_network_inspect` | Inspect a network |
| `docker_network_connect` | Connect a container to a network |
| `docker_network_disconnect` | Disconnect a container from a network |
| `docker_network_remove` | Remove a network |

### Compose (5)

| Tool | Description |
|---|---|
| `docker_compose_up` | Start a Compose project |
| `docker_compose_down` | Stop and remove a Compose project |
| `docker_compose_ps` | List services in a Compose project |
| `docker_compose_logs` | Fetch logs from Compose services |
| `docker_compose_pull` | Pull images for a Compose project |

### System (4)

| Tool | Description |
|---|---|
| `docker_system_info` | Docker system-wide information |
| `docker_system_version` | Docker version details |
| `docker_system_df` | Disk usage summary |
| `docker_system_prune` | Remove unused Docker resources |

---

## Development

All commands are run from the `src/` directory.

```bash
# Run the Go backend directly against your host Docker socket
make dev

# Run with auto-generated TLS
make dev-tls

# Start the React UI dev server (hot-reload)
make ui-dev

# Run Go tests
make test

# Tidy Go module dependencies
make tidy

# Run Docker's validator against the local image
make validate-local

# Build, push, and validate a release image for linux/amd64 and linux/arm64
make validate-release RELEASE_IMAGE=docker.io/<owner>/docker-mcp-extension:1.0.0
```

### Project structure

```
src/
├── Dockerfile             # Multi-stage build (Node → Go → Alpine runtime)
├── docker-compose.yaml    # Extension service definition
├── Makefile               # Developer shortcuts
├── metadata.json          # Docker Desktop extension manifest
├── backend/               # Go MCP server
│   ├── main.go
│   └── internal/
│       ├── docker/        # Docker API wrappers
│       └── mcp/           # MCP protocol & tool dispatch
└── ui/                    # React dashboard (TypeScript)
```

---

## License

[MIT](LICENSE)

