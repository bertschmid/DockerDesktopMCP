// Package docker wraps the Docker SDK and exposes operations as MCP tool results.
package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
	"docker-mcp/internal/mcp"
)

// Client wraps the Docker SDK client.
type Client struct {
	cli *client.Client
}

// NewClient creates a new Docker client using environment / default socket.
func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("creating docker client: %w", err)
	}
	return &Client{cli: cli}, nil
}

// Ping verifies the Docker daemon is reachable.
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.cli.Ping(ctx)
	return err
}

// Close releases resources held by the client.
func (c *Client) Close() error {
	return c.cli.Close()
}

// ok returns a simple success result.
func ok(msg string) (*mcp.CallToolResult, error) {
	return mcp.TextResult(msg), nil
}
