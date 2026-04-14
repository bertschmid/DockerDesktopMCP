package docker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"docker-mcp/internal/result"
)

// NetworkList lists all Docker networks.
func (c *Client) NetworkList(ctx context.Context) (*result.CallToolResult, error) {
	networks, err := c.cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return nil, err
	}

	type row struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Driver string `json:"driver"`
		Scope  string `json:"scope"`
		Subnet string `json:"subnet,omitempty"`
	}

	rows := make([]row, 0, len(networks))
	for _, n := range networks {
		subnet := ""
		if len(n.IPAM.Config) > 0 {
			subnet = n.IPAM.Config[0].Subnet
		}
		rows = append(rows, row{
			ID:     n.ID[:12],
			Name:   n.Name,
			Driver: n.Driver,
			Scope:  n.Scope,
			Subnet: subnet,
		})
	}

	out, _ := json.MarshalIndent(rows, "", "  ")
	return result.Text(string(out)), nil
}

// NetworkCreate creates a new Docker network.
func (c *Client) NetworkCreate(ctx context.Context, name, driver, subnet string) (*result.CallToolResult, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	opts := types.NetworkCreate{
		Driver:     driver,
		CheckDuplicate: true,
	}

	if subnet != "" {
		opts.IPAM = &network.IPAM{
			Driver: "default",
			Config: []network.IPAMConfig{{Subnet: subnet}},
		}
	}

	resp, err := c.cli.NetworkCreate(ctx, name, opts)
	if err != nil {
		return nil, err
	}

	out, _ := json.MarshalIndent(map[string]any{
		"id":      resp.ID[:12],
		"warning": resp.Warning,
	}, "", "  ")
	return result.Text(string(out)), nil
}

// NetworkRemove removes a network.
func (c *Client) NetworkRemove(ctx context.Context, name string) (*result.CallToolResult, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if err := c.cli.NetworkRemove(ctx, name); err != nil {
		return nil, err
	}
	return ok(fmt.Sprintf("Network %s removed", name))
}

// NetworkInspect returns detailed info about a network.
func (c *Client) NetworkInspect(ctx context.Context, name string) (*result.CallToolResult, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	n, err := c.cli.NetworkInspect(ctx, name, types.NetworkInspectOptions{Verbose: true})
	if err != nil {
		return nil, err
	}
	out, _ := json.MarshalIndent(n, "", "  ")
	return result.Text(string(out)), nil
}

// NetworkConnect connects a container to a network.
func (c *Client) NetworkConnect(ctx context.Context, networkName, containerID string) (*result.CallToolResult, error) {
	if networkName == "" || containerID == "" {
		return nil, fmt.Errorf("network and container are required")
	}
	if err := c.cli.NetworkConnect(ctx, networkName, containerID, nil); err != nil {
		return nil, err
	}
	return ok(fmt.Sprintf("Container %s connected to network %s", containerID, networkName))
}

// NetworkDisconnect disconnects a container from a network.
func (c *Client) NetworkDisconnect(ctx context.Context, networkName, containerID string, force bool) (*result.CallToolResult, error) {
	if networkName == "" || containerID == "" {
		return nil, fmt.Errorf("network and container are required")
	}
	if err := c.cli.NetworkDisconnect(ctx, networkName, containerID, force); err != nil {
		return nil, err
	}
	return ok(fmt.Sprintf("Container %s disconnected from network %s", containerID, networkName))
}
