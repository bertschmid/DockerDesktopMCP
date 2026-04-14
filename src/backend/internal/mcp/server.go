package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"docker-mcp/internal/docker"
)

// Server is an HTTP handler implementing the MCP Streamable HTTP transport.
type Server struct {
	docker *docker.Client
	tools  []Tool
}

// NewServer constructs an MCP server with all Docker tools registered.
func NewServer(dc *docker.Client) *Server {
	s := &Server{docker: dc}
	s.registerTools()
	return s
}

// ServeHTTP dispatches HTTP requests.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS for local AI clients
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	switch r.URL.Path {
	case "/mcp":
		s.handleMCP(w, r)
	case "/health":
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","server":"Docker Desktop MCP","version":"1.0.0"}`)
	default:
		http.NotFound(w, r)
	}
}

// handleMCP processes all JSON-RPC 2.0 MCP messages.
func (s *Server) handleMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, nil, ErrParse, "Parse error: "+err.Error())
		return
	}

	if req.JSONRPC != "2.0" {
		writeError(w, req.ID, ErrInvalidRequest, `jsonrpc must be "2.0"`)
		return
	}

	log.Printf("[MCP] method=%s id=%s", req.Method, string(req.ID))

	ctx := r.Context()

	switch req.Method {
	case "initialize":
		writeResult(w, req.ID, InitializeResult{
			ProtocolVersion: ProtocolVersion,
			Capabilities:    ServerCaps{Tools: &ToolsCap{ListChanged: false}},
			ServerInfo:      ServerInfo{Name: "Docker Desktop MCP", Version: "1.0.0"},
		})
	case "initialized":
		w.WriteHeader(http.StatusAccepted)
	case "ping":
		writeResult(w, req.ID, map[string]any{})
	case "tools/list":
		writeResult(w, req.ID, ListToolsResult{Tools: s.tools})
	case "tools/call":
		s.handleToolCall(w, ctx, req)
	default:
		writeError(w, req.ID, ErrMethodNotFound, fmt.Sprintf("method not found: %s", req.Method))
	}
}

func (s *Server) handleToolCall(w http.ResponseWriter, ctx context.Context, req JSONRPCRequest) {
	var params CallToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		writeError(w, req.ID, ErrInvalidParams, "invalid params: "+err.Error())
		return
	}
	if params.Arguments == nil {
		params.Arguments = map[string]any{}
	}
	result := s.dispatchTool(ctx, params.Name, params.Arguments)
	writeResult(w, req.ID, result)
}

// ─── JSON-RPC Helpers ─────────────────────────────────────────────────────────

func writeResult(w http.ResponseWriter, id json.RawMessage, result any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(JSONRPCResponse{JSONRPC: "2.0", ID: id, Result: result}); err != nil {
		log.Printf("[MCP] encode error: %v", err)
	}
}

func writeError(w http.ResponseWriter, id json.RawMessage, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(JSONRPCResponse{
		JSONRPC: "2.0", ID: id,
		Error: &RPCError{Code: code, Message: message},
	}); err != nil {
		log.Printf("[MCP] encode error: %v", err)
	}
}

// ─── Tool Dispatcher ─────────────────────────────────────────────────────────

func (s *Server) dispatchTool(ctx context.Context, name string, args map[string]any) *CallToolResult {
	c := s.docker
	var result *CallToolResult
	var err error

	switch name {
	// ── Containers ──────────────────────────────────────────────────────────
	case "docker_container_list":
		result, err = c.ContainerList(ctx, getBool(args, "all", false))
	case "docker_container_inspect":
		result, err = c.ContainerInspect(ctx, getStr(args, "id", ""))
	case "docker_container_start":
		result, err = c.ContainerStart(ctx, getStr(args, "id", ""))
	case "docker_container_stop":
		result, err = c.ContainerStop(ctx, getStr(args, "id", ""), getInt(args, "timeout", 10))
	case "docker_container_restart":
		result, err = c.ContainerRestart(ctx, getStr(args, "id", ""), getInt(args, "timeout", 10))
	case "docker_container_remove":
		result, err = c.ContainerRemove(ctx, getStr(args, "id", ""), getBool(args, "force", false), getBool(args, "volumes", false))
	case "docker_container_logs":
		result, err = c.ContainerLogs(ctx, getStr(args, "id", ""), getStr(args, "tail", "100"), getBool(args, "timestamps", false), getStr(args, "since", ""))
	case "docker_container_exec":
		result, err = c.ContainerExec(ctx, getStr(args, "id", ""), getStr(args, "command", ""), getStr(args, "user", ""))
	case "docker_container_stats":
		result, err = c.ContainerStats(ctx, getStr(args, "id", ""))
	case "docker_container_create":
		result, err = c.ContainerCreate(ctx, args)

	// ── Images ──────────────────────────────────────────────────────────────
	case "docker_image_list":
		result, err = c.ImageList(ctx, getBool(args, "all", false))
	case "docker_image_pull":
		result, err = c.ImagePull(ctx, getStr(args, "image", ""), getStr(args, "platform", ""))
	case "docker_image_remove":
		result, err = c.ImageRemove(ctx, getStr(args, "image", ""), getBool(args, "force", false))
	case "docker_image_inspect":
		result, err = c.ImageInspect(ctx, getStr(args, "image", ""))
	case "docker_image_tag":
		result, err = c.ImageTag(ctx, getStr(args, "source", ""), getStr(args, "target", ""))
	case "docker_image_build":
		result, err = c.ImageBuild(ctx, args)

	// ── Volumes ─────────────────────────────────────────────────────────────
	case "docker_volume_list":
		result, err = c.VolumeList(ctx)
	case "docker_volume_create":
		result, err = c.VolumeCreate(ctx, getStr(args, "name", ""), getStr(args, "driver", "local"))
	case "docker_volume_remove":
		result, err = c.VolumeRemove(ctx, getStr(args, "name", ""), getBool(args, "force", false))
	case "docker_volume_inspect":
		result, err = c.VolumeInspect(ctx, getStr(args, "name", ""))

	// ── Networks ────────────────────────────────────────────────────────────
	case "docker_network_list":
		result, err = c.NetworkList(ctx)
	case "docker_network_create":
		result, err = c.NetworkCreate(ctx, getStr(args, "name", ""), getStr(args, "driver", "bridge"), getStr(args, "subnet", ""))
	case "docker_network_remove":
		result, err = c.NetworkRemove(ctx, getStr(args, "name", ""))
	case "docker_network_inspect":
		result, err = c.NetworkInspect(ctx, getStr(args, "name", ""))
	case "docker_network_connect":
		result, err = c.NetworkConnect(ctx, getStr(args, "network", ""), getStr(args, "container", ""))
	case "docker_network_disconnect":
		result, err = c.NetworkDisconnect(ctx, getStr(args, "network", ""), getStr(args, "container", ""), getBool(args, "force", false))

	// ── Compose ─────────────────────────────────────────────────────────────
	case "docker_compose_up":
		result, err = c.ComposeUp(ctx, getStr(args, "project_dir", ""), getStrSlice(args, "services"), getBool(args, "detach", true), getBool(args, "build", false), getBool(args, "force_recreate", false))
	case "docker_compose_down":
		result, err = c.ComposeDown(ctx, getStr(args, "project_dir", ""), getBool(args, "volumes", false), getBool(args, "remove_orphans", false))
	case "docker_compose_ps":
		result, err = c.ComposePs(ctx, getStr(args, "project_dir", ""))
	case "docker_compose_logs":
		result, err = c.ComposeLogs(ctx, getStr(args, "project_dir", ""), getStrSlice(args, "services"), getStr(args, "tail", "100"))
	case "docker_compose_pull":
		result, err = c.ComposePull(ctx, getStr(args, "project_dir", ""), getStrSlice(args, "services"))

	// ── System ──────────────────────────────────────────────────────────────
	case "docker_system_info":
		result, err = c.SystemInfo(ctx)
	case "docker_system_version":
		result, err = c.SystemVersion(ctx)
	case "docker_system_df":
		result, err = c.SystemDf(ctx)
	case "docker_system_prune":
		result, err = c.SystemPrune(ctx, getBool(args, "all", false), getBool(args, "volumes", false))

	default:
		return errorResult(fmt.Sprintf("unknown tool: %s", name))
	}

	if err != nil {
		return errorResult(fmt.Sprintf("tool %s failed: %v", name, err))
	}
	return result
}
