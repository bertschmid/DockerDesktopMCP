package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"docker-mcp/internal/mcp"
)

// ContainerList lists containers.
func (c *Client) ContainerList(ctx context.Context, all bool) (*mcp.CallToolResult, error) {
	containers, err := c.cli.ContainerList(ctx, types.ContainerListOptions{All: all})
	if err != nil {
		return nil, err
	}

	type row struct {
		ID      string   `json:"id"`
		Names   []string `json:"names"`
		Image   string   `json:"image"`
		Status  string   `json:"status"`
		State   string   `json:"state"`
		Ports   []string `json:"ports"`
		Created string   `json:"created"`
	}

	rows := make([]row, 0, len(containers))
	for _, ct := range containers {
		ports := make([]string, 0)
		for _, p := range ct.Ports {
			if p.IP != "" {
				ports = append(ports, fmt.Sprintf("%s:%d->%d/%s", p.IP, p.PublicPort, p.PrivatePort, p.Type))
			} else {
				ports = append(ports, fmt.Sprintf("%d/%s", p.PrivatePort, p.Type))
			}
		}
		rows = append(rows, row{
			ID:      ct.ID[:12],
			Names:   ct.Names,
			Image:   ct.Image,
			Status:  ct.Status,
			State:   ct.State,
			Ports:   ports,
			Created: time.Unix(ct.Created, 0).UTC().Format(time.RFC3339),
		})
	}

	out, _ := json.MarshalIndent(rows, "", "  ")
	return mcp.TextResult(string(out)), nil
}

// ContainerInspect returns detailed info about a container.
func (c *Client) ContainerInspect(ctx context.Context, id string) (*mcp.CallToolResult, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	info, _, err := c.cli.ContainerInspectWithRaw(ctx, id, false)
	if err != nil {
		return nil, err
	}
	out, _ := json.MarshalIndent(info, "", "  ")
	return mcp.TextResult(string(out)), nil
}

// ContainerStart starts a container.
func (c *Client) ContainerStart(ctx context.Context, id string) (*mcp.CallToolResult, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	if err := c.cli.ContainerStart(ctx, id, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}
	return ok(fmt.Sprintf("Container %s started", id))
}

// ContainerStop stops a container.
func (c *Client) ContainerStop(ctx context.Context, id string, timeout int) (*mcp.CallToolResult, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	stopOptions := container.StopOptions{Timeout: &timeout}
	if err := c.cli.ContainerStop(ctx, id, stopOptions); err != nil {
		return nil, err
	}
	return ok(fmt.Sprintf("Container %s stopped", id))
}

// ContainerRestart restarts a container.
func (c *Client) ContainerRestart(ctx context.Context, id string, timeout int) (*mcp.CallToolResult, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	stopOptions := container.StopOptions{Timeout: &timeout}
	if err := c.cli.ContainerRestart(ctx, id, stopOptions); err != nil {
		return nil, err
	}
	return ok(fmt.Sprintf("Container %s restarted", id))
}

// ContainerRemove removes a container.
func (c *Client) ContainerRemove(ctx context.Context, id string, force, removeVolumes bool) (*mcp.CallToolResult, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	err := c.cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{
		Force:         force,
		RemoveVolumes: removeVolumes,
	})
	if err != nil {
		return nil, err
	}
	return ok(fmt.Sprintf("Container %s removed", id))
}

// ContainerLogs returns logs for a container.
func (c *Client) ContainerLogs(ctx context.Context, id, tail string, timestamps bool, since string) (*mcp.CallToolResult, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	opts := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tail,
		Timestamps: timestamps,
		Since:      since,
	}
	reader, err := c.cli.ContainerLogs(ctx, id, opts)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// Strip Docker multiplexing header (8 bytes per log line)
	var buf bytes.Buffer
	raw, _ := io.ReadAll(reader)
	i := 0
	for i < len(raw) {
		if i+8 > len(raw) {
			break
		}
		size := int(raw[i+4])<<24 | int(raw[i+5])<<16 | int(raw[i+6])<<8 | int(raw[i+7])
		i += 8
		if i+size > len(raw) {
			buf.Write(raw[i:])
			break
		}
		buf.Write(raw[i : i+size])
		i += size
	}
	return mcp.TextResult(buf.String()), nil
}

// ContainerExec runs a command in a container and returns stdout+stderr.
func (c *Client) ContainerExec(ctx context.Context, id, command, user string) (*mcp.CallToolResult, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	if command == "" {
		return nil, fmt.Errorf("command is required")
	}

	// Split command string into args
	args := splitCommand(command)

	execConfig := types.ExecConfig{
		Cmd:          args,
		AttachStdout: true,
		AttachStderr: true,
	}
	if user != "" {
		execConfig.User = user
	}

	execID, err := c.cli.ContainerExecCreate(ctx, id, execConfig)
	if err != nil {
		return nil, fmt.Errorf("exec create: %w", err)
	}

	resp, err := c.cli.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{})
	if err != nil {
		return nil, fmt.Errorf("exec attach: %w", err)
	}
	defer resp.Close()

	var buf bytes.Buffer
	io.Copy(&buf, resp.Reader)

	// Get exit code
	inspect, _ := c.cli.ContainerExecInspect(ctx, execID.ID)
	result := buf.String()
	if inspect.ExitCode != 0 {
		return mcp.TextResult(fmt.Sprintf("exit code %d\n%s", inspect.ExitCode, result)), nil
	}
	return mcp.TextResult(result), nil
}

// ContainerStats returns current resource usage for a container.
func (c *Client) ContainerStats(ctx context.Context, id string) (*mcp.CallToolResult, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	resp, err := c.cli.ContainerStats(ctx, id, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var stats types.StatsJSON
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, err
	}

	// Calculate CPU %
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
	numCPUs := float64(stats.CPUStats.OnlineCPUs)
	if numCPUs == 0 {
		numCPUs = float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
	}
	cpuPercent := 0.0
	if systemDelta > 0 {
		cpuPercent = (cpuDelta / systemDelta) * numCPUs * 100.0
	}

	// Memory
	memUsage := stats.MemoryStats.Usage - stats.MemoryStats.Stats["cache"]
	memLimit := stats.MemoryStats.Limit
	memPercent := 0.0
	if memLimit > 0 {
		memPercent = float64(memUsage) / float64(memLimit) * 100.0
	}

	summary := map[string]any{
		"container_id": stats.ID[:12],
		"name":         strings.TrimPrefix(stats.Name, "/"),
		"cpu_percent":  fmt.Sprintf("%.2f%%", cpuPercent),
		"memory": map[string]any{
			"usage":   formatBytes(memUsage),
			"limit":   formatBytes(memLimit),
			"percent": fmt.Sprintf("%.2f%%", memPercent),
		},
		"network": buildNetworkStats(stats.Networks),
		"block_io": map[string]any{
			"read":  formatBytes(blkioRead(stats.BlkioStats)),
			"write": formatBytes(blkioWrite(stats.BlkioStats)),
		},
		"pids": stats.PidsStats.Current,
	}

	out, _ := json.MarshalIndent(summary, "", "  ")
	return mcp.TextResult(string(out)), nil
}

// ContainerCreate creates a new container without starting it.
func (c *Client) ContainerCreate(ctx context.Context, args map[string]any) (*mcp.CallToolResult, error) {
	image := getStr(args, "image", "")
	if image == "" {
		return nil, fmt.Errorf("image is required")
	}

	cfg := &container.Config{
		Image: image,
	}
	if cmd := getStr(args, "command", ""); cmd != "" {
		cfg.Cmd = splitCommand(cmd)
	}
	if envSlice := getStrSlice(args, "env"); len(envSlice) > 0 {
		cfg.Env = envSlice
	}

	// Port bindings
	hostCfg := &container.HostConfig{}
	hostCfg.RestartPolicy = container.RestartPolicy{Name: container.RestartPolicyMode(getStr(args, "restart", "no"))}

	if ports := getStrSlice(args, "ports"); len(ports) > 0 {
		portBindings := nat.PortMap{}
		exposedPorts := nat.PortSet{}
		for _, p := range ports {
			parts := strings.SplitN(p, ":", 2)
			if len(parts) == 2 {
				port := nat.Port(parts[1] + "/tcp")
				portBindings[port] = []nat.PortBinding{{HostPort: parts[0]}}
				exposedPorts[port] = struct{}{}
			}
		}
		hostCfg.PortBindings = portBindings
		cfg.ExposedPorts = exposedPorts
	}

	if vols := getStrSlice(args, "volumes"); len(vols) > 0 {
		hostCfg.Binds = vols
	}

	netCfg := &network.NetworkingConfig{}
	if netName := getStr(args, "network", ""); netName != "" {
		netCfg.EndpointsConfig = map[string]*network.EndpointSettings{
			netName: {},
		}
	}

	resp, err := c.cli.ContainerCreate(ctx, cfg, hostCfg, netCfg, nil, getStr(args, "name", ""))
	if err != nil {
		return nil, err
	}

	out, _ := json.MarshalIndent(map[string]any{
		"id":       resp.ID[:12],
		"warnings": resp.Warnings,
	}, "", "  ")
	return mcp.TextResult(string(out)), nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func splitCommand(cmd string) []string {
	// Simple shell-like split (handles quoted strings)
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := byte(0)

	for i := 0; i < len(cmd); i++ {
		ch := cmd[i]
		if inQuote {
			if ch == quoteChar {
				inQuote = false
			} else {
				current.WriteByte(ch)
			}
		} else if ch == '"' || ch == '\'' {
			inQuote = true
			quoteChar = ch
		} else if ch == ' ' || ch == '\t' {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(ch)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func buildNetworkStats(networks map[string]types.NetworkStats) map[string]any {
	if len(networks) == 0 {
		return nil
	}
	result := map[string]any{}
	for name, n := range networks {
		result[name] = map[string]any{
			"rx": formatBytes(n.RxBytes),
			"tx": formatBytes(n.TxBytes),
		}
	}
	return result
}

func blkioRead(s types.BlkioStats) uint64 {
	var total uint64
	for _, e := range s.IoServiceBytesRecursive {
		if strings.EqualFold(e.Op, "read") {
			total += e.Value
		}
	}
	return total
}

func blkioWrite(s types.BlkioStats) uint64 {
	var total uint64
	for _, e := range s.IoServiceBytesRecursive {
		if strings.EqualFold(e.Op, "write") {
			total += e.Value
		}
	}
	return total
}

// Re-export arg helpers so they can be used in this package
func getStr(args map[string]any, key, def string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func getBool(args map[string]any, key string, def bool) bool {
	if v, ok := args[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return def
}

func getStrSlice(args map[string]any, key string) []string {
	if v, ok := args[key]; ok {
		if arr, ok := v.([]any); ok {
			result := make([]string, 0, len(arr))
			for _, item := range arr {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
	}
	return nil
}

