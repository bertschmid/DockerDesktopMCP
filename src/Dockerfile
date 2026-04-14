# ── Stage 1: Build the React UI ──────────────────────────────────────────────
FROM --platform=$BUILDPLATFORM public.ecr.aws/docker/library/node:22-alpine AS ui-builder

WORKDIR /app/ui

# Cache node_modules layer separately from source changes
COPY ui/package.json ./
RUN npm install && npm dedupe

COPY ui/ .
RUN npm run build

# ── Stage 2: Build the Go MCP server ─────────────────────────────────────────
FROM --platform=$BUILDPLATFORM public.ecr.aws/docker/library/golang:1.25-alpine AS go-builder

WORKDIR /src
RUN apk add --no-cache git ca-certificates

# Copy entire backend — go mod tidy generates go.sum if it doesn't exist yet
COPY backend/ .
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-s -w" -o /mcp-server ./main.go

# ── Stage 3: Minimal runtime image ───────────────────────────────────────────
FROM public.ecr.aws/docker/library/alpine:3

# Docker CLI ships with docker compose (needed for Compose tools)
RUN apk add --no-cache docker-cli ca-certificates

# Go binary
COPY --from=go-builder /mcp-server /mcp-server

# React UI — served by Docker Desktop from /ui in the container filesystem
COPY --from=ui-builder /app/ui/dist /ui

# Extension metadata (required at image root)
COPY metadata.json .
COPY docker-compose.yaml .
COPY icon.svg .

# ── Docker Desktop Extension Labels ──────────────────────────────────────────
LABEL org.opencontainers.image.title="Docker Desktop MCP Server" \
      org.opencontainers.image.description="HTTP MCP server exposing the full Docker Desktop API to AI assistants." \
      org.opencontainers.image.vendor="bert.schmid@googlemail.com" \
      org.opencontainers.image.version="1.0.2" \
      com.docker.desktop.extension.api.version=">= 0.3.3" \
    com.docker.extension.categories="utility-tools" \
      com.docker.desktop.extension.icon="https://raw.githubusercontent.com/bertschmid/DockerDesktopMCP/main/src/icon.svg" \
    com.docker.extension.screenshots='[]' \
      com.docker.extension.detailed-description="<h1>Docker Desktop MCP Server</h1><p>Starts a local HTTP server implementing the Model Context Protocol (MCP). Connect any MCP-compatible AI assistant (Claude, etc.) to manage containers, images, volumes, networks, and Compose stacks from natural language.</p><h2>Features</h2><ul><li>35 MCP tools covering the full Docker Desktop API</li><li>Configurable port (default: 3282)</li><li>Optional TLS with auto-generated self-signed certificate</li></ul>" \
      com.docker.extension.publisher-url="https://github.com/bertschmid/DockerDesktopMCP" \
    com.docker.extension.additional-urls='[{"title":"Project Repository","url":"https://github.com/bertschmid/DockerDesktopMCP"}]' \
      com.docker.extension.changelog="<ul><li>v1.0.1 — Dark mode support (follows system preference).</li><li>v1.0.0 — Initial release — 35 MCP tools for containers, images, volumes, networks, Compose, and system operations.</li></ul>"

EXPOSE 3282

ENTRYPOINT ["/mcp-server"]
CMD ["--port=3282", "--host=0.0.0.0"]
