package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"docker-mcp/internal/result"
)

const composeProjectDirRequiredError = "project_dir is required"

// ComposeUp starts services defined in a docker-compose.yml file.
func (c *Client) ComposeUp(ctx context.Context, projectDir string, services []string, detach, build, forceRecreate bool) (*result.CallToolResult, error) {
	if projectDir == "" {
		return nil, fmt.Errorf(composeProjectDirRequiredError)
	}

	args := []string{"compose", "up"}
	if detach {
		args = append(args, "-d")
	}
	if build {
		args = append(args, "--build")
	}
	if forceRecreate {
		args = append(args, "--force-recreate")
	}
	args = append(args, services...)

	out, err := runDockerCLI(ctx, projectDir, args...)
	if err != nil {
		return result.Text(fmt.Sprintf("compose up failed:\n%s", out)), nil
	}
	if out == "" {
		out = "Services started successfully"
	}
	return result.Text(out), nil
}

// ComposeDown stops and removes containers and networks.
func (c *Client) ComposeDown(ctx context.Context, projectDir string, volumes, removeOrphans bool) (*result.CallToolResult, error) {
	if projectDir == "" {
		return nil, fmt.Errorf(composeProjectDirRequiredError)
	}

	args := []string{"compose", "down"}
	if volumes {
		args = append(args, "-v")
	}
	if removeOrphans {
		args = append(args, "--remove-orphans")
	}

	out, err := runDockerCLI(ctx, projectDir, args...)
	if err != nil {
		return result.Text(fmt.Sprintf("compose down failed:\n%s", out)), nil
	}
	if out == "" {
		out = "Services stopped and removed"
	}
	return result.Text(out), nil
}

// ComposePs lists containers for a compose project.
func (c *Client) ComposePs(ctx context.Context, projectDir string) (*result.CallToolResult, error) {
	if projectDir == "" {
		return nil, fmt.Errorf(composeProjectDirRequiredError)
	}

	rawJSON, err := runDockerCLI(ctx, projectDir, "compose", "ps", "--format", "json")
	if err != nil {
		return result.Text(fmt.Sprintf("compose ps failed:\n%s", rawJSON)), nil
	}

	var services []map[string]any
	if err := json.Unmarshal([]byte(rawJSON), &services); err != nil {
		// Fallback if Compose version does not support --format json output.
		fallbackOut, fallbackErr := runDockerCLI(ctx, projectDir, "compose", "ps")
		if fallbackErr != nil {
			return result.Text(fmt.Sprintf("compose ps failed:\n%s", fallbackOut)), nil
		}
		return result.TextStructuredUI(
			fallbackOut,
			map[string]any{"services": []map[string]any{}, "raw": fallbackOut, "project_dir": projectDir},
			"ui://docker-desktop/compose-services",
		), nil
	}

	out, err := json.MarshalIndent(services, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal compose ps: %w", err)
	}

	return result.TextStructuredUI(
		string(out),
		map[string]any{"services": services, "project_dir": projectDir},
		"ui://docker-desktop/compose-services",
	), nil
}

// ComposeLogs fetches logs from compose services.
func (c *Client) ComposeLogs(ctx context.Context, projectDir string, services []string, tail string) (*result.CallToolResult, error) {
	if projectDir == "" {
		return nil, fmt.Errorf(composeProjectDirRequiredError)
	}

	args := []string{"compose", "logs", "--no-color", "--tail", tail}
	args = append(args, services...)

	out, err := runDockerCLI(ctx, projectDir, args...)
	if err != nil {
		return result.Text(fmt.Sprintf("compose logs failed:\n%s", out)), nil
	}
	return result.Text(out), nil
}

// ComposePull pulls images for all or specified services.
func (c *Client) ComposePull(ctx context.Context, projectDir string, services []string) (*result.CallToolResult, error) {
	if projectDir == "" {
		return nil, fmt.Errorf(composeProjectDirRequiredError)
	}

	args := []string{"compose", "pull"}
	args = append(args, services...)

	out, err := runDockerCLI(ctx, projectDir, args...)
	if err != nil {
		return result.Text(fmt.Sprintf("compose pull failed:\n%s", out)), nil
	}
	if out == "" {
		out = "Images pulled successfully"
	}
	return result.Text(out), nil
}

// runDockerCLI executes a docker CLI command in the given working directory.
// It captures both stdout and stderr, returning combined output.
func runDockerCLI(ctx context.Context, workDir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	combined := strings.TrimSpace(stdout.String() + stderr.String())
	return combined, err
}
