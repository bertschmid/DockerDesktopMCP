EXTENSION_IMAGE ?= docker-mcp-extension:latest
PORT            ?= 3282

.PHONY: help build install update uninstall dev dev-tls test tidy ui-install ui-build ui-dev validate lint

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "\033[36m%-20s\033[0m %s\n",$$1,$$2}'

# ── UI ────────────────────────────────────────────────────────────────────────

ui-install: ## Install React UI dependencies
	cd ui && npm install --legacy-peer-deps

ui-build: ui-install ## Build the React UI
	cd ui && npm run build

ui-dev: ui-install ## Start React UI dev server (hot-reload, outside Docker Desktop)
	cd ui && npm start

# ── Extension ─────────────────────────────────────────────────────────────────

build: ui-build ## Build the full extension Docker image (UI + Go backend)
	docker build -t $(EXTENSION_IMAGE) .

install: build ## Install the extension into Docker Desktop
	docker extension install $(EXTENSION_IMAGE)

update: build ## Update an already-installed extension
	docker extension update $(EXTENSION_IMAGE)

uninstall: ## Uninstall the extension from Docker Desktop
	docker extension uninstall $(EXTENSION_IMAGE)

validate: build ## Validate the extension image against Docker Desktop requirements
	docker extension validate $(EXTENSION_IMAGE)

# ── Backend (Go) ──────────────────────────────────────────────────────────────

dev: ## Run only the Go MCP server locally (uses host Docker socket)
	cd backend && go run . --port=$(PORT)

dev-tls: ## Run with auto-generated self-signed TLS certificate
	cd backend && go run . --port=$(PORT) --tls

test: ## Run Go tests
	cd backend && go test ./...

tidy: ## Tidy Go module dependencies
	cd backend && go mod tidy

lint: ## Run golangci-lint on the Go backend
	cd backend && golangci-lint run ./...

.DEFAULT_GOAL := help
