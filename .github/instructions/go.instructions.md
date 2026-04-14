---
applyTo: "**/*.go"
---

# Go Best Practices — DockerDesktopMCP Backend

## Environment

- **Go version**: 1.25.0 (see `src/backend/go.mod`)
- **Module**: `docker-mcp`
- **Key dependencies**:
  - `github.com/docker/docker v26.1.4+incompatible` — Docker Engine API client
  - `github.com/docker/go-connections v0.5.0` — Docker connection helpers

## Code Style

- Format all code with `gofmt` before committing. No unformatted code is acceptable.
- Follow the [Effective Go](https://go.dev/doc/effective_go) guidelines.
- Use `goimports` to manage import blocks (stdlib → external → internal, each separated by a blank line).
- Keep lines under 120 characters where practical.
- Exported identifiers must have a doc comment (`// FunctionName …`).

## Error Handling

- Always check and handle errors explicitly — never use `_` to discard an error silently.
- Wrap errors with context using `fmt.Errorf("doing X: %w", err)` so callers can inspect the chain with `errors.Is` / `errors.As`.
- Use `log.Fatalf` only in `main` for unrecoverable startup errors; return errors from all other functions.
- Never panic in library or handler code; convert panics to errors at package boundaries when necessary.

## Package & File Organisation

- One package per directory; package name matches the directory name.
- Keep `main.go` thin — only flag parsing, wiring, and server lifecycle.
- Place Docker API wrappers under `internal/docker/` and MCP protocol logic under `internal/mcp/`.
- Use the `internal/` visibility boundary to prevent accidental external imports.

## Naming Conventions

- Use `camelCase` for unexported and `PascalCase` for exported identifiers.
- Acronyms remain upper-case when exported: `MCP`, `URL`, `HTTP`, `TLS`, `ID`.
- Receiver names are short (one or two letters) and consistent across all methods of a type.
- Constructor functions follow the `New<Type>` pattern (e.g. `NewServer`, `NewClient`).

## Concurrency

- Prefer `context.Context` propagation for cancellation and timeouts; always accept `ctx` as the first parameter of functions that perform I/O.
- Use `sync.Mutex` or `sync/atomic` for shared state; document which fields are guarded.
- Avoid spawning goroutines without a clear ownership and shutdown path.
- Use `signal.Notify` + graceful shutdown as shown in `main.go` — honour `SIGTERM`.

## HTTP & MCP Server

- Register handlers through the `mcp.Server` type; do not add routes directly in `main`.
- Set explicit `ReadTimeout`, `WriteTimeout`, and `IdleTimeout` on every `http.Server`.
- Return structured JSON error responses; never expose raw Go error strings to callers.
- Use `http.Error` only for non-JSON endpoints; for MCP endpoints write a proper JSON error body.

## TLS

- Minimum TLS version: `tls.VersionTLS12` (already enforced in `buildTLSConfig`).
- Prefer ECDSA P-256 keys for auto-generated certificates.
- Load external certificates only via `tls.LoadX509KeyPair`.

## Testing

- Run tests with `make test` from the `src/` directory.
- Table-driven tests are preferred for functions with multiple cases.
- Use `testify` assertions only if the dependency is already present; otherwise use `gotest.tools/v3` (already a transitive dependency).
- Do not modify or delete existing tests without a documented reason.

## Docker API Usage

- Always pass a `context.Context` to Docker client calls.
- Close the Docker client with `defer dockerClient.Close()` in `main`.
- Perform a lightweight `Ping` on startup to detect connectivity issues early, but treat failure as a warning rather than fatal (the daemon may start after the extension).

## Module Management

- Run `make tidy` (`go mod tidy`) after adding or removing dependencies.
- Pin direct dependencies to explicit versions in `go.mod`; keep indirect dependencies in the `require` block marked `// indirect`.
- Never vendor the module; rely on the Go module cache.

## Security

- Do not log or expose secrets, tokens, or credentials.
- Validate and sanitise all input that arrives via HTTP before passing it to Docker API calls.
- Bind the server to `127.0.0.1` by default; only expose externally when explicitly configured.
