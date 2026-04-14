package mcp

// registerTools populates s.tools with all 35 Docker MCP tool definitions.
func (s *Server) registerTools() {
	s.tools = []Tool{
		// ── Containers ────────────────────────────────────────────────────────
		{
			Name:        "docker_container_list",
			Description: "List Docker containers. Set all=true to include stopped containers.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"all": {Type: "boolean", Description: "Include stopped containers (default: false)"},
				},
			},
		},
		{
			Name:        "docker_container_inspect",
			Description: "Return detailed information about a container.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"id": {Type: "string", Description: "Container ID or name"},
				},
				Required: []string{"id"},
			},
		},
		{
			Name:        "docker_container_create",
			Description: "Create a new container without starting it.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"image":   {Type: "string", Description: "Image name and optional tag"},
					"name":    {Type: "string", Description: "Container name (optional)"},
					"command": {Type: "string", Description: "Command to run (optional)"},
					"env":     {Type: "array", Description: "Environment variables in KEY=VALUE format", Items: &Items{Type: "string"}},
					"ports":   {Type: "array", Description: "Port bindings in HOST:CONTAINER format", Items: &Items{Type: "string"}},
					"volumes": {Type: "array", Description: "Volume bindings in HOST:CONTAINER format", Items: &Items{Type: "string"}},
					"network": {Type: "string", Description: "Network to attach the container to"},
					"restart": {Type: "string", Description: "Restart policy: no, always, unless-stopped, on-failure"},
				},
				Required: []string{"image"},
			},
		},
		{
			Name:        "docker_container_start",
			Description: "Start a stopped container.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"id": {Type: "string", Description: "Container ID or name"},
				},
				Required: []string{"id"},
			},
		},
		{
			Name:        "docker_container_stop",
			Description: "Stop a running container.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"id":      {Type: "string", Description: "Container ID or name"},
					"timeout": {Type: "integer", Description: "Seconds to wait before killing (default: 10)"},
				},
				Required: []string{"id"},
			},
		},
		{
			Name:        "docker_container_restart",
			Description: "Restart a container.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"id":      {Type: "string", Description: "Container ID or name"},
					"timeout": {Type: "integer", Description: "Seconds to wait before killing (default: 10)"},
				},
				Required: []string{"id"},
			},
		},
		{
			Name:        "docker_container_remove",
			Description: "Remove a container.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"id":      {Type: "string", Description: "Container ID or name"},
					"force":   {Type: "boolean", Description: "Force removal of a running container"},
					"volumes": {Type: "boolean", Description: "Remove associated anonymous volumes"},
				},
				Required: []string{"id"},
			},
		},
		{
			Name:        "docker_container_logs",
			Description: "Fetch logs from a container.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"id":         {Type: "string", Description: "Container ID or name"},
					"tail":       {Type: "string", Description: "Number of lines from the end (default: 100)"},
					"timestamps": {Type: "boolean", Description: "Show timestamps"},
					"since":      {Type: "string", Description: "Show logs since this timestamp or duration (e.g. 1h)"},
				},
				Required: []string{"id"},
			},
		},
		{
			Name:        "docker_container_exec",
			Description: "Execute a command inside a running container and return stdout+stderr.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"id":      {Type: "string", Description: "Container ID or name"},
					"command": {Type: "string", Description: "Command to execute"},
					"user":    {Type: "string", Description: "User to run the command as (optional)"},
				},
				Required: []string{"id", "command"},
			},
		},
		{
			Name:        "docker_container_stats",
			Description: "Return live CPU, memory, network and I/O statistics for a container.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"id": {Type: "string", Description: "Container ID or name"},
				},
				Required: []string{"id"},
			},
		},

		// ── Images ────────────────────────────────────────────────────────────
		{
			Name:        "docker_image_list",
			Description: "List local Docker images.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"all": {Type: "boolean", Description: "Include intermediate images"},
				},
			},
		},
		{
			Name:        "docker_image_pull",
			Description: "Pull an image from a registry.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"image":    {Type: "string", Description: "Image reference (e.g. nginx:latest)"},
					"platform": {Type: "string", Description: "Target platform (e.g. linux/arm64)"},
				},
				Required: []string{"image"},
			},
		},
		{
			Name:        "docker_image_build",
			Description: "Build an image from a Dockerfile context directory.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"context_path": {Type: "string", Description: "Path to the build context directory"},
					"dockerfile":   {Type: "string", Description: "Dockerfile name (default: Dockerfile)"},
					"tag":          {Type: "string", Description: "Image tag (e.g. myapp:latest)"},
					"no_cache":     {Type: "boolean", Description: "Do not use cache"},
					"build_args":   {Type: "array", Description: "Build arguments in KEY=VALUE format", Items: &Items{Type: "string"}},
				},
				Required: []string{"context_path"},
			},
		},
		{
			Name:        "docker_image_tag",
			Description: "Create a new tag for an existing image.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"source": {Type: "string", Description: "Source image reference"},
					"target": {Type: "string", Description: "Target image reference"},
				},
				Required: []string{"source", "target"},
			},
		},
		{
			Name:        "docker_image_inspect",
			Description: "Return detailed metadata about a local image.",
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
			Description: "Remove a local image.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"image": {Type: "string", Description: "Image reference or ID"},
					"force": {Type: "boolean", Description: "Force removal"},
				},
				Required: []string{"image"},
			},
		},

		// ── Volumes ───────────────────────────────────────────────────────────
		{
			Name:        "docker_volume_list",
			Description: "List all Docker volumes.",
			InputSchema: InputSchema{Type: "object"},
		},
		{
			Name:        "docker_volume_create",
			Description: "Create a new named Docker volume.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"name":   {Type: "string", Description: "Volume name"},
					"driver": {Type: "string", Description: "Volume driver (default: local)"},
				},
			},
		},
		{
			Name:        "docker_volume_inspect",
			Description: "Return detailed information about a volume.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"name": {Type: "string", Description: "Volume name"},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "docker_volume_remove",
			Description: "Remove a Docker volume.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"name":  {Type: "string", Description: "Volume name"},
					"force": {Type: "boolean", Description: "Force removal"},
				},
				Required: []string{"name"},
			},
		},

		// ── Networks ──────────────────────────────────────────────────────────
		{
			Name:        "docker_network_list",
			Description: "List all Docker networks.",
			InputSchema: InputSchema{Type: "object"},
		},
		{
			Name:        "docker_network_create",
			Description: "Create a new Docker network.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"name":   {Type: "string", Description: "Network name"},
					"driver": {Type: "string", Description: "Network driver (default: bridge)"},
					"subnet": {Type: "string", Description: "Subnet in CIDR format (optional)"},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "docker_network_inspect",
			Description: "Return detailed information about a network.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"name": {Type: "string", Description: "Network name or ID"},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "docker_network_connect",
			Description: "Connect a container to a network.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"network":   {Type: "string", Description: "Network name or ID"},
					"container": {Type: "string", Description: "Container name or ID"},
				},
				Required: []string{"network", "container"},
			},
		},
		{
			Name:        "docker_network_disconnect",
			Description: "Disconnect a container from a network.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"network":   {Type: "string", Description: "Network name or ID"},
					"container": {Type: "string", Description: "Container name or ID"},
					"force":     {Type: "boolean", Description: "Force disconnection"},
				},
				Required: []string{"network", "container"},
			},
		},
		{
			Name:        "docker_network_remove",
			Description: "Remove a Docker network.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"name": {Type: "string", Description: "Network name or ID"},
				},
				Required: []string{"name"},
			},
		},

		// ── Compose ───────────────────────────────────────────────────────────
		{
			Name:        "docker_compose_up",
			Description: "Start services defined in a docker-compose.yml file.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_dir":    {Type: "string", Description: "Path to the directory containing docker-compose.yml"},
					"services":       {Type: "array", Description: "Specific services to start (empty = all)", Items: &Items{Type: "string"}},
					"detach":         {Type: "boolean", Description: "Run in background (default: true)"},
					"build":          {Type: "boolean", Description: "Build images before starting"},
					"force_recreate": {Type: "boolean", Description: "Recreate containers even if nothing changed"},
				},
				Required: []string{"project_dir"},
			},
		},
		{
			Name:        "docker_compose_down",
			Description: "Stop and remove containers and networks for a Compose project.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_dir":    {Type: "string", Description: "Path to the directory containing docker-compose.yml"},
					"volumes":        {Type: "boolean", Description: "Remove named and anonymous volumes"},
					"remove_orphans": {Type: "boolean", Description: "Remove containers for services not in the Compose file"},
				},
				Required: []string{"project_dir"},
			},
		},
		{
			Name:        "docker_compose_ps",
			Description: "List containers for a Compose project.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_dir": {Type: "string", Description: "Path to the directory containing docker-compose.yml"},
				},
				Required: []string{"project_dir"},
			},
		},
		{
			Name:        "docker_compose_logs",
			Description: "Fetch logs from Compose services.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_dir": {Type: "string", Description: "Path to the directory containing docker-compose.yml"},
					"services":    {Type: "array", Description: "Specific services to get logs from (empty = all)", Items: &Items{Type: "string"}},
					"tail":        {Type: "string", Description: "Number of lines from the end (default: 100)"},
				},
				Required: []string{"project_dir"},
			},
		},
		{
			Name:        "docker_compose_pull",
			Description: "Pull images for all or specified Compose services.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_dir": {Type: "string", Description: "Path to the directory containing docker-compose.yml"},
					"services":    {Type: "array", Description: "Specific services to pull (empty = all)", Items: &Items{Type: "string"}},
				},
				Required: []string{"project_dir"},
			},
		},

		// ── System ────────────────────────────────────────────────────────────
		{
			Name:        "docker_system_info",
			Description: "Return Docker system-wide information (daemon version, resources, etc.).",
			InputSchema: InputSchema{Type: "object"},
		},
		{
			Name:        "docker_system_version",
			Description: "Return Docker client and server version details.",
			InputSchema: InputSchema{Type: "object"},
		},
		{
			Name:        "docker_system_df",
			Description: "Return disk usage summary for images, containers, volumes, and build cache.",
			InputSchema: InputSchema{Type: "object"},
		},
		{
			Name:        "docker_system_prune",
			Description: "Remove unused Docker resources (stopped containers, dangling images, unused networks, build cache).",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"all":     {Type: "boolean", Description: "Remove all unused images, not just dangling ones"},
					"volumes": {Type: "boolean", Description: "Also prune unused volumes"},
				},
			},
		},
	}
}
