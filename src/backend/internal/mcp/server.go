package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"docker-mcp/internal/docker"
)

const (
	contentTypeHeader = "Content-Type"
	jsonContentType   = "application/json"
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
		w.Header().Set(contentTypeHeader, jsonContentType)
		fmt.Fprintf(w, `{"status":"ok","server":"Docker Desktop MCP","version":"1.0.13"}`)
	case "/admin/restart":
		s.handleRestart(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set(contentTypeHeader, jsonContentType)
	if err := json.NewEncoder(w).Encode(map[string]any{"status": "restarting"}); err != nil {
		log.Printf("[MCP] restart response encode error: %v", err)
	}
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	go func() {
		log.Println("[MCP] restart requested via API")
		time.Sleep(1 * time.Second)
		os.Exit(0)
	}()
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
			Capabilities: ServerCaps{
				Tools:     &ToolsCap{ListChanged: false},
				Resources: &ResourcesCap{ListChanged: false},
			},
			ServerInfo: ServerInfo{Name: "Docker Desktop MCP", Version: "1.0.13"},
		})
	case "initialized":
		w.WriteHeader(http.StatusAccepted)
	case "ping":
		writeResult(w, req.ID, map[string]any{})
	case "tools/list":
		writeResult(w, req.ID, ListToolsResult{Tools: s.tools})
	case "tools/call":
		s.handleToolCall(w, ctx, req)
	case "resources/list":
		writeResult(w, req.ID, s.listResources())
	case "resources/read":
		var params ReadResourceParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			writeError(w, req.ID, ErrInvalidParams, "invalid params: "+err.Error())
			return
		}
		res, err := s.readResource(params)
		if err != nil {
			writeError(w, req.ID, ErrInvalidParams, err.Error())
			return
		}
		writeResult(w, req.ID, res)
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
	w.Header().Set(contentTypeHeader, jsonContentType)
	if err := json.NewEncoder(w).Encode(JSONRPCResponse{JSONRPC: "2.0", ID: id, Result: result}); err != nil {
		log.Printf("[MCP] encode error: %v", err)
	}
}

func writeError(w http.ResponseWriter, id json.RawMessage, code int, message string) {
	w.Header().Set(contentTypeHeader, jsonContentType)
	if err := json.NewEncoder(w).Encode(JSONRPCResponse{
		JSONRPC: "2.0", ID: id,
		Error: &RPCError{Code: code, Message: message},
	}); err != nil {
		log.Printf("[MCP] encode error: %v", err)
	}
}

// ─── Tool Dispatcher ─────────────────────────────────────────────────────────

func (s *Server) dispatchTool(ctx context.Context, name string, args map[string]any) *CallToolResult {
	result, err := s.dispatchContainerTool(ctx, name, args)
	if result == nil && err == nil {
		result, err = s.dispatchImageTool(ctx, name, args)
	}
	if result == nil && err == nil {
		result, err = s.dispatchVolumeTool(ctx, name, args)
	}
	if result == nil && err == nil {
		result, err = s.dispatchNetworkTool(ctx, name, args)
	}
	if result == nil && err == nil {
		result, err = s.dispatchComposeTool(ctx, name, args)
	}
	if result == nil && err == nil {
		result, err = s.dispatchSystemTool(ctx, name, args)
	}
	if result == nil && err == nil {
		return errorResult(fmt.Sprintf("unknown tool: %s", name))
	}

	if err != nil {
		return errorResult(fmt.Sprintf("tool %s failed: %v", name, err))
	}
	return result
}

func (s *Server) dispatchContainerTool(ctx context.Context, name string, args map[string]any) (*CallToolResult, error) {
	c := s.docker

	switch name {
	case "docker_container_list":
		return c.ContainerList(ctx, getBool(args, "all", false))
	case "docker_container_inspect":
		return c.ContainerInspect(ctx, getStr(args, "id", ""))
	case "docker_container_start":
		return c.ContainerStart(ctx, getStr(args, "id", ""))
	case "docker_container_stop":
		return c.ContainerStop(ctx, getStr(args, "id", ""), getInt(args, "timeout", 10))
	case "docker_container_restart":
		return c.ContainerRestart(ctx, getStr(args, "id", ""), getInt(args, "timeout", 10))
	case "docker_container_remove":
		return c.ContainerRemove(ctx, getStr(args, "id", ""), getBool(args, "force", false), getBool(args, "volumes", false))
	case "docker_container_logs":
		return c.ContainerLogs(ctx, getStr(args, "id", ""), getStr(args, "tail", "100"), getBool(args, "timestamps", false), getStr(args, "since", ""))
	case "docker_container_exec":
		return c.ContainerExec(ctx, getStr(args, "id", ""), getStr(args, "command", ""), getStr(args, "user", ""), getStr(args, "workdir", ""))
	case "docker_container_stats":
		return c.ContainerStats(ctx, getStr(args, "id", ""))
	case "docker_container_create":
		return c.ContainerCreate(ctx, args)
	default:
		return nil, nil
	}
}

func (s *Server) dispatchImageTool(ctx context.Context, name string, args map[string]any) (*CallToolResult, error) {
	c := s.docker

	switch name {
	case "docker_image_list":
		return c.ImageList(ctx, getBool(args, "all", false))
	case "docker_image_pull":
		return c.ImagePull(ctx, getStr(args, "image", ""), getStr(args, "platform", ""))
	case "docker_image_remove":
		return c.ImageRemove(ctx, getStr(args, "image", ""), getBool(args, "force", false))
	case "docker_image_inspect":
		return c.ImageInspect(ctx, getStr(args, "image", ""))
	case "docker_image_tag":
		return c.ImageTag(ctx, getStr(args, "source", ""), getStr(args, "target", ""))
	case "docker_image_build":
		return c.ImageBuild(ctx, args)
	default:
		return nil, nil
	}
}

func (s *Server) dispatchVolumeTool(ctx context.Context, name string, args map[string]any) (*CallToolResult, error) {
	c := s.docker

	switch name {
	case "docker_volume_list":
		return c.VolumeList(ctx)
	case "docker_volume_create":
		return c.VolumeCreate(ctx, getStr(args, "name", ""), getStr(args, "driver", "local"))
	case "docker_volume_remove":
		return c.VolumeRemove(ctx, getStr(args, "name", ""), getBool(args, "force", false))
	case "docker_volume_inspect":
		return c.VolumeInspect(ctx, getStr(args, "name", ""))
	default:
		return nil, nil
	}
}

func (s *Server) dispatchNetworkTool(ctx context.Context, name string, args map[string]any) (*CallToolResult, error) {
	c := s.docker

	switch name {
	case "docker_network_list":
		return c.NetworkList(ctx)
	case "docker_network_create":
		return c.NetworkCreate(ctx, getStr(args, "name", ""), getStr(args, "driver", "bridge"), getStr(args, "subnet", ""))
	case "docker_network_remove":
		return c.NetworkRemove(ctx, getStr(args, "name", ""))
	case "docker_network_inspect":
		return c.NetworkInspect(ctx, getStr(args, "name", ""))
	case "docker_network_connect":
		return c.NetworkConnect(ctx, getStr(args, "network", ""), getStr(args, "container", ""))
	case "docker_network_disconnect":
		return c.NetworkDisconnect(ctx, getStr(args, "network", ""), getStr(args, "container", ""), getBool(args, "force", false))
	default:
		return nil, nil
	}
}

func (s *Server) dispatchComposeTool(ctx context.Context, name string, args map[string]any) (*CallToolResult, error) {
	c := s.docker

	switch name {
	case "docker_compose_up":
		return c.ComposeUp(ctx, getStr(args, "project_dir", ""), getStrSlice(args, "services"), getBool(args, "detach", true), getBool(args, "build", false), getBool(args, "force_recreate", false))
	case "docker_compose_down":
		return c.ComposeDown(ctx, getStr(args, "project_dir", ""), getBool(args, "volumes", false), getBool(args, "remove_orphans", false))
	case "docker_compose_ps":
		return c.ComposePs(ctx, getStr(args, "project_dir", ""))
	case "docker_compose_logs":
		return c.ComposeLogs(ctx, getStr(args, "project_dir", ""), getStrSlice(args, "services"), getStr(args, "tail", "100"))
	case "docker_compose_pull":
		return c.ComposePull(ctx, getStr(args, "project_dir", ""), getStrSlice(args, "services"))
	default:
		return nil, nil
	}
}

func (s *Server) dispatchSystemTool(ctx context.Context, name string, args map[string]any) (*CallToolResult, error) {
	c := s.docker

	switch name {
	case "docker_system_info":
		return c.SystemInfo(ctx)
	case "docker_system_version":
		return c.SystemVersion(ctx)
	case "docker_system_df":
		return c.SystemDf(ctx)
	case "docker_system_prune_all":
		return c.SystemPruneAll(ctx, getStrSlice(args, "filters"))
	case "docker_system_prune_containers":
		return c.SystemPruneContainers(ctx, getStrSlice(args, "filters"))
	case "docker_system_prune_images":
		return c.SystemPruneImages(ctx, getStrSlice(args, "filters"))
	case "docker_system_prune_networks":
		return c.SystemPruneNetworks(ctx, getStrSlice(args, "filters"))
	case "docker_system_prune_build_cache":
		return c.SystemPruneBuildCache(ctx, getStrSlice(args, "filters"))
	case "docker_system_prune_volumes":
		return c.SystemPruneVolumes(ctx, getStrSlice(args, "filters"))
	default:
		return nil, nil
	}
}
