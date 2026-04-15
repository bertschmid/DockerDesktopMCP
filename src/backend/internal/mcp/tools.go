package mcp

const (
	containerIDOrNameDescription   = "Container ID or name"
	volumeNameDescription          = "Volume name"
	networkNameOrIDDescription     = "Network name or ID"
	composeProjectDirDescription   = "Path to the directory containing docker-compose.yml"
	uiResourceURIKey               = "ui/resourceUri"
	systemPruneFiltersDescription  = "Optional prune filters as strings such as label=team=platform, label!=keep, or until=24h"
)

func uiResourceMeta(resourceURI string) map[string]any {
	return map[string]any{
		"ui": map[string]any{"resourceUri": resourceURI},
		uiResourceURIKey: resourceURI,
	}
}

// registerTools populates s.tools with all registered Docker MCP tool definitions.
func (s *Server) registerTools() {
	s.tools = []Tool{
		// ── Containers ────────────────────────────────────────────────────────
		{
			Name:        "docker_container_list",
			Description: "List Docker containers for inventory and status checks. Use this as first step before inspect/logs/exec actions. Example: {\"all\":true}.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"all": {Type: "boolean", Description: "Include non-running containers. false: only running containers. true: running + stopped/exited containers (default: false)"},
				},
			},
			Meta: uiResourceMeta("ui://docker-desktop/containers"),
		},
		{
			Name:        "docker_container_inspect",
			Description: "Return full low-level metadata for one container, including mounts, networking, labels, and runtime config. Example: {\"id\":\"web-1\"}.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"id": {Type: "string", Description: containerIDOrNameDescription},
				},
				Required: []string{"id"},
			},
		},
		{
			Name:        "docker_container_create",
			Description: "Create a new container without starting it. Use this to stage configuration before start. Example: {\"image\":\"nginx:latest\",\"name\":\"web-1\",\"ports\":[\"8080:80\"]}.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"image":   {Type: "string", Description: "Image reference, usually name:tag. Example: nginx:latest"},
					"name":    {Type: "string", Description: "Optional container name. Example: web-1"},
					"command": {Type: "string", Description: "Optional override command. Example: sleep infinity"},
					"env":     {Type: "array", Description: "Optional environment variables as KEY=VALUE entries. Example: [\"APP_ENV=prod\",\"DEBUG=false\"]", Items: &Items{Type: "string"}},
					"ports":   {Type: "array", Description: "Optional port bindings as HOST:CONTAINER. Example: [\"8080:80\"]", Items: &Items{Type: "string"}},
					"volumes": {Type: "array", Description: "Optional volume bindings as HOST:CONTAINER. Example: [\"/data:/var/lib/app\"]", Items: &Items{Type: "string"}},
					"network": {Type: "string", Description: "Optional network name for initial attachment. Example: bridge"},
					"restart": {Type: "string", Description: "Optional restart policy. Allowed values: no, always, unless-stopped, on-failure"},
				},
				Required: []string{"image"},
			},
		},
		{
			Name:        "docker_container_start",
			Description: "Start a stopped container. Example: {\"id\":\"web-1\"}.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"id": {Type: "string", Description: containerIDOrNameDescription},
				},
				Required: []string{"id"},
			},
		},
		{
			Name:        "docker_container_stop",
			Description: "Stop a running container gracefully, then force kill after timeout if needed. Example: {\"id\":\"web-1\",\"timeout\":20}.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"id":      {Type: "string", Description: containerIDOrNameDescription},
					"timeout": {Type: "integer", Description: "Grace period in seconds before forced termination. Example: 20 (default: 10)"},
				},
				Required: []string{"id"},
			},
		},
		{
			Name:        "docker_container_restart",
			Description: "Restart a container (stop then start). Example: {\"id\":\"web-1\",\"timeout\":10}.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"id":      {Type: "string", Description: containerIDOrNameDescription},
					"timeout": {Type: "integer", Description: "Grace period in seconds before forced stop during restart (default: 10)"},
				},
				Required: []string{"id"},
			},
		},
		{
			Name:        "docker_container_remove",
			Description: "Remove a container and optionally force removal and anonymous volume cleanup. Example: {\"id\":\"web-1\",\"force\":true,\"volumes\":true}.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"id":      {Type: "string", Description: containerIDOrNameDescription},
					"force":   {Type: "boolean", Description: "Force removal. false: remove only safely removable container. true: also remove running container"},
					"volumes": {Type: "boolean", Description: "Also remove anonymous volumes. false: keep anonymous volumes. true: delete anonymous volumes with container"},
				},
				Required: []string{"id"},
			},
		},
		{
			Name:        "docker_container_logs",
			Description: "Fetch container logs with optional line limit, timestamps, and since-filter. Example: {\"id\":\"web-1\",\"tail\":\"200\",\"timestamps\":true,\"since\":\"1h\"}.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"id":         {Type: "string", Description: containerIDOrNameDescription},
					"tail":       {Type: "string", Description: "Tail line count from end. Example: \"300\" (default: 100)"},
					"timestamps": {Type: "boolean", Description: "Include timestamps. false: plain log lines. true: each line prefixed with timestamp"},
					"since":      {Type: "string", Description: "Only return logs newer than this timestamp or duration. Examples: \"1h\", \"2026-04-15T10:00:00Z\""},
				},
				Required: []string{"id"},
			},
		},
		{
			Name:        "docker_container_exec",
			Description: "Execute a command inside a running container and return stdout/stderr output. Example: {\"id\":\"web-1\",\"command\":\"ls -la /\",\"user\":\"root\"}.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"id":      {Type: "string", Description: containerIDOrNameDescription},
					"command": {Type: "string", Description: "Shell command to execute inside container. Example: cat /etc/os-release"},
					"user":    {Type: "string", Description: "Optional user identity. Examples: root, 1000, 1000:1000"},
					"workdir": {Type: "string", Description: "Optional working directory. Example: /app"},
				},
				Required: []string{"id", "command"},
			},
		},
		{
			Name:        "docker_container_stats",
			Description: "Return live CPU, memory, network and I/O statistics for one running container. Example: {\"id\":\"web-1\"}.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"id": {Type: "string", Description: containerIDOrNameDescription},
				},
				Required: []string{"id"},
			},
		},

		// ── Images ────────────────────────────────────────────────────────────
		{
			Name:        "docker_image_list",
			Description: "List local Docker images for cleanup and deployment planning. Example: {\"all\":true}.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"all": {Type: "boolean", Description: "Include intermediate images. false: omit intermediates. true: include intermediates"},
				},
			},
			Meta: uiResourceMeta("ui://docker-desktop/images"),
		},
		{
			Name:        "docker_image_pull",
			Description: "Pull an image from a registry, optionally for a specific platform. Example: {\"image\":\"redis:7\",\"platform\":\"linux/amd64\"}.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"image":    {Type: "string", Description: "Image reference to pull. Example: nginx:latest"},
					"platform": {Type: "string", Description: "Optional target platform. Example: linux/arm64"},
				},
				Required: []string{"image"},
			},
		},
		{
			Name:        "docker_image_build",
			Description: "Build an image from a Dockerfile context directory with optional Dockerfile override, tag, cache toggle, and build args. Example: {\"context_path\":\"C:/repo/app\",\"tag\":\"myapp:prod\",\"no_cache\":true}.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"context_path": {Type: "string", Description: "Path to build context directory. Example: C:/repo/app"},
					"dockerfile":   {Type: "string", Description: "Dockerfile filename/path relative to context. Example: Dockerfile.prod (default: Dockerfile)"},
					"tag":          {Type: "string", Description: "Optional image tag for build output. Example: myapp:latest"},
					"no_cache":     {Type: "boolean", Description: "Disable build cache. false: allow cached layers. true: force clean rebuild"},
					"build_args":   {Type: "array", Description: "Optional build arguments as KEY=VALUE. Example: [\"APP_ENV=prod\",\"COMMIT_SHA=abc123\"]", Items: &Items{Type: "string"}},
				},
				Required: []string{"context_path"},
			},
		},
		{
			Name:        "docker_image_tag",
			Description: "Create a new tag for an existing local image reference. Example: {\"source\":\"myapp:latest\",\"target\":\"registry.example.com/myapp:1.2.3\"}.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"source": {Type: "string", Description: "Existing local source image reference"},
					"target": {Type: "string", Description: "New target image reference (name:tag)"},
				},
				Required: []string{"source", "target"},
			},
		},
		{
			Name:        "docker_image_inspect",
			Description: "Return detailed metadata about a local image, including layers, config, labels, and architecture. Example: {\"image\":\"redis:7\"}.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"image": {Type: "string", Description: "Image reference or ID"},
				},
				Required: []string{"image"},
			},
		},
		{
			Name:        "docker_image_remove",
			Description: "Remove a local image reference and related tags where possible. Example: {\"image\":\"redis:7\",\"force\":true}.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"image": {Type: "string", Description: "Image reference or ID"},
					"force": {Type: "boolean", Description: "Force removal. false: remove only without conflicts. true: force delete even when referenced"},
				},
				Required: []string{"image"},
			},
		},

		// ── Volumes ───────────────────────────────────────────────────────────
		{
			Name:        "docker_volume_list",
			Description: "List all Docker volumes to inspect persistent storage usage and cleanup targets.",
			InputSchema: InputSchema{Type: "object"},
			Meta: uiResourceMeta("ui://docker-desktop/volumes"),
		},
		{
			Name:        "docker_volume_create",
			Description: "Create a new named Docker volume. Example: {\"name\":\"pgdata\",\"driver\":\"local\"}.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"name":   {Type: "string", Description: "Optional volume name. Example: pgdata"},
					"driver": {Type: "string", Description: "Volume driver plugin. Example: local (default: local)"},
				},
			},
		},
		{
			Name:        "docker_volume_inspect",
			Description: "Return detailed information about one volume, including mountpoint, labels, and usage metadata.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"name": {Type: "string", Description: volumeNameDescription},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "docker_volume_remove",
			Description: "Remove a Docker volume by name. Example: {\"name\":\"pgdata\",\"force\":true}.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"name":  {Type: "string", Description: volumeNameDescription},
					"force": {Type: "boolean", Description: "Force removal. false: remove only when unused. true: force delete"},
				},
				Required: []string{"name"},
			},
		},

		// ── Networks ──────────────────────────────────────────────────────────
		{
			Name:        "docker_network_list",
			Description: "List all Docker networks for connectivity checks and cleanup planning.",
			InputSchema: InputSchema{Type: "object"},
			Meta: uiResourceMeta("ui://docker-desktop/networks"),
		},
		{
			Name:        "docker_network_create",
			Description: "Create a new Docker network with optional driver and subnet. Example: {\"name\":\"app-net\",\"driver\":\"bridge\",\"subnet\":\"172.28.0.0/16\"}.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"name":   {Type: "string", Description: "Network name. Example: app-net"},
					"driver": {Type: "string", Description: "Network driver. Example: bridge (default: bridge)"},
					"subnet": {Type: "string", Description: "Optional subnet in CIDR format. Example: 172.28.0.0/16"},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "docker_network_inspect",
			Description: "Return detailed information about a network, including connected containers, IPAM, labels, and options.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"name": {Type: "string", Description: networkNameOrIDDescription},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "docker_network_connect",
			Description: "Connect a container to a network. Example: {\"network\":\"app-net\",\"container\":\"web-1\"}.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"network":   {Type: "string", Description: networkNameOrIDDescription},
					"container": {Type: "string", Description: "Container name or ID to attach"},
				},
				Required: []string{"network", "container"},
			},
		},
		{
			Name:        "docker_network_disconnect",
			Description: "Disconnect a container from a network. Example: {\"network\":\"app-net\",\"container\":\"web-1\",\"force\":true}.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"network":   {Type: "string", Description: networkNameOrIDDescription},
					"container": {Type: "string", Description: "Container name or ID to detach"},
					"force":     {Type: "boolean", Description: "Force disconnection. false: normal detach, may fail with conflicts. true: force detach"},
				},
				Required: []string{"network", "container"},
			},
		},
		{
			Name:        "docker_network_remove",
			Description: "Remove a Docker network by name or ID.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"name": {Type: "string", Description: networkNameOrIDDescription},
				},
				Required: []string{"name"},
			},
		},

		// ── Compose ───────────────────────────────────────────────────────────
		{
			Name:        "docker_compose_up",
			Description: "Start services from a Compose project with optional subset, image build, detach mode, and force recreate. Example: {\"project_dir\":\"C:/repo/stack\",\"services\":[\"api\"],\"detach\":true,\"build\":true}.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_dir":    {Type: "string", Description: composeProjectDirDescription},
					"services":       {Type: "array", Description: "Optional service names. Empty/omitted means all services. Example: [\"api\",\"worker\"]", Items: &Items{Type: "string"}},
					"detach":         {Type: "boolean", Description: "Run mode. true: start in background. false: run attached/foreground (default: true)"},
					"build":          {Type: "boolean", Description: "Build policy. false: use existing images. true: build images before start"},
					"force_recreate": {Type: "boolean", Description: "Recreation policy. false: recreate only when needed. true: always recreate containers"},
				},
				Required: []string{"project_dir"},
			},
		},
		{
			Name:        "docker_compose_down",
			Description: "Stop and remove Compose project resources. Optionally remove volumes and orphan containers. Example: {\"project_dir\":\"C:/repo/stack\",\"volumes\":true,\"remove_orphans\":true}.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_dir":    {Type: "string", Description: composeProjectDirDescription},
					"volumes":        {Type: "boolean", Description: "Volume cleanup. false: keep named/anonymous volumes. true: remove named and anonymous volumes"},
					"remove_orphans": {Type: "boolean", Description: "Orphan cleanup. false: keep orphan service containers. true: remove containers not defined in current Compose file"},
				},
				Required: []string{"project_dir"},
			},
		},
		{
			Name:        "docker_compose_ps",
			Description: "List containers and states for a Compose project. Useful for service health and rollout checks.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_dir": {Type: "string", Description: composeProjectDirDescription},
				},
				Required: []string{"project_dir"},
			},
			Meta: uiResourceMeta("ui://docker-desktop/compose-services"),
		},
		{
			Name:        "docker_compose_logs",
			Description: "Fetch logs from one or more Compose services. Example: {\"project_dir\":\"C:/repo/stack\",\"services\":[\"api\"],\"tail\":\"300\"}.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_dir": {Type: "string", Description: composeProjectDirDescription},
					"services":    {Type: "array", Description: "Optional service names for log scope. Empty/omitted means all services", Items: &Items{Type: "string"}},
					"tail":        {Type: "string", Description: "Tail line count from end. Example: \"1000\" (default: 100)"},
				},
				Required: []string{"project_dir"},
			},
		},
		{
			Name:        "docker_compose_pull",
			Description: "Pull images for all or selected Compose services before deployment/update.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_dir": {Type: "string", Description: composeProjectDirDescription},
					"services":    {Type: "array", Description: "Optional service names to pull. Empty/omitted means all services", Items: &Items{Type: "string"}},
				},
				Required: []string{"project_dir"},
			},
		},

		// ── System ────────────────────────────────────────────────────────────
		{
			Name:        "docker_system_info",
			Description: "Inspect host and daemon metadata (OS, kernel, CPU, memory, storage driver, counts, runtime settings). Use this before orchestration or cleanup decisions.",
			InputSchema: InputSchema{Type: "object"},
			Meta: uiResourceMeta("ui://docker-desktop/system-info"),
		},
		{
			Name:        "docker_system_version",
			Description: "Return detailed Docker client/server version metadata for compatibility checks and version-specific planning.",
			InputSchema: InputSchema{Type: "object"},
		},
		{
			Name:        "docker_system_df",
			Description: "Summarize Docker disk usage across images, writable container layers, volumes, and build cache, including reclaimable space for cleanup planning.",
			InputSchema: InputSchema{Type: "object"},
			Meta: uiResourceMeta("ui://docker-desktop/disk-usage"),
		},
		{
			Name:        "docker_system_prune_all",
			Description: "Aggressively clear all unused Docker resources in one operation: stopped containers, unused images, unused networks, all unused build cache, and unused volumes. Use this when an AI should reclaim as much local Docker space as possible without specifying individual cleanup scopes. Filter examples: filters=[\"until=24h\"], filters=[\"label=cleanup=true\"], filters=[\"label!=keep\"].",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"filters": {Type: "array", Description: systemPruneFiltersDescription, Items: &Items{Type: "string"}},
			}},
		},
		{
			Name:        "docker_system_prune_containers",
			Description: "Remove stopped containers and other container resources that are no longer used. Use this when cleanup should target containers only and leave images, networks, volumes, and build cache untouched. Filter examples: filters=[\"until=168h\"], filters=[\"label=env=dev\"], filters=[\"label!=critical\"].",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"filters": {Type: "array", Description: systemPruneFiltersDescription, Items: &Items{Type: "string"}},
			}},
		},
		{
			Name:        "docker_system_prune_images",
			Description: "Remove unused Docker images, including images that are not dangling but are no longer referenced by any container. Use this to reclaim image storage without touching containers, networks, volumes, or build cache. Filter examples: filters=[\"until=720h\"], filters=[\"label=stage=ci\"], filters=[\"label!=retain\"].",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"filters": {Type: "array", Description: systemPruneFiltersDescription, Items: &Items{Type: "string"}},
			}},
		},
		{
			Name:        "docker_system_prune_networks",
			Description: "Remove unused Docker networks that are not currently attached to active containers. Use this to clean up stale networking artifacts without affecting images, containers, volumes, or build cache. Filter examples: filters=[\"until=48h\"], filters=[\"label=scope=temp\"], filters=[\"label!=shared\"].",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"filters": {Type: "array", Description: systemPruneFiltersDescription, Items: &Items{Type: "string"}},
			}},
		},
		{
			Name:        "docker_system_prune_build_cache",
			Description: "Remove all unused Docker build cache records created by image builds and BuildKit. Use this when disk pressure is caused by cached layers rather than images or volumes. Filter examples: filters=[\"until=72h\"], filters=[\"id=sha256:...\"], filters=[\"parent=sha256:...\"].",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"filters": {Type: "array", Description: systemPruneFiltersDescription, Items: &Items{Type: "string"}},
			}},
		},
		{
			Name:        "docker_system_prune_volumes",
			Description: "Remove unused Docker volumes that are no longer referenced by any container. Use this carefully when an AI should reclaim persistent storage without deleting images, networks, or containers. Filter examples: filters=[\"label=cleanup=true\"], filters=[\"label=project=tmp\"], filters=[\"label!=keep\"].",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"filters": {Type: "array", Description: systemPruneFiltersDescription, Items: &Items{Type: "string"}},
			}},
		},
	}
}
