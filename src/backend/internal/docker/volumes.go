package docker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/docker/docker/api/types/volume"
	"docker-mcp/internal/mcp"
)

// VolumeList lists all volumes.
func (c *Client) VolumeList(ctx context.Context) (*mcp.CallToolResult, error) {
	resp, err := c.cli.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, err
	}

	type row struct {
		Name       string `json:"name"`
		Driver     string `json:"driver"`
		Mountpoint string `json:"mountpoint"`
		Created    string `json:"created,omitempty"`
		Scope      string `json:"scope"`
	}

	rows := make([]row, 0, len(resp.Volumes))
	for _, v := range resp.Volumes {
		rows = append(rows, row{
			Name:       v.Name,
			Driver:     v.Driver,
			Mountpoint: v.Mountpoint,
			Created:    v.CreatedAt,
			Scope:      v.Scope,
		})
	}

	out, _ := json.MarshalIndent(map[string]any{
		"volumes":  rows,
		"warnings": resp.Warnings,
	}, "", "  ")
	return mcp.TextResult(string(out)), nil
}

// VolumeCreate creates a new named volume.
func (c *Client) VolumeCreate(ctx context.Context, name, driver string) (*mcp.CallToolResult, error) {
	v, err := c.cli.VolumeCreate(ctx, volume.CreateOptions{
		Name:   name,
		Driver: driver,
	})
	if err != nil {
		return nil, err
	}
	out, _ := json.MarshalIndent(map[string]any{
		"name":       v.Name,
		"driver":     v.Driver,
		"mountpoint": v.Mountpoint,
		"created":    v.CreatedAt,
	}, "", "  ")
	return mcp.TextResult(string(out)), nil
}

// VolumeRemove removes a volume.
func (c *Client) VolumeRemove(ctx context.Context, name string, force bool) (*mcp.CallToolResult, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if err := c.cli.VolumeRemove(ctx, name, force); err != nil {
		return nil, err
	}
	return ok(fmt.Sprintf("Volume %s removed", name))
}

// VolumeInspect returns detailed info about a volume.
func (c *Client) VolumeInspect(ctx context.Context, name string) (*mcp.CallToolResult, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	v, err := c.cli.VolumeInspect(ctx, name)
	if err != nil {
		return nil, err
	}
	out, _ := json.MarshalIndent(v, "", "  ")
	return mcp.TextResult(string(out)), nil
}
