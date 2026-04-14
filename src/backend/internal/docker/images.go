package docker

import (
	"archive/tar"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"docker-mcp/internal/mcp"
)

// ImageList lists local images.
func (c *Client) ImageList(ctx context.Context, all bool) (*mcp.CallToolResult, error) {
	images, err := c.cli.ImageList(ctx, types.ImageListOptions{All: all})
	if err != nil {
		return nil, err
	}

	type row struct {
		ID         string   `json:"id"`
		Repository []string `json:"repository"`
		Tags       []string `json:"tags"`
		Size       string   `json:"size"`
		Created    string   `json:"created"`
	}

	rows := make([]row, 0, len(images))
	for _, img := range images {
		repos := make([]string, 0)
		tags := make([]string, 0)
		for _, tag := range img.RepoTags {
			parts := strings.SplitN(tag, ":", 2)
			if len(parts) == 2 {
				repos = append(repos, parts[0])
				tags = append(tags, parts[1])
			} else {
				repos = append(repos, tag)
			}
		}
		shortID := img.ID
		if strings.HasPrefix(img.ID, "sha256:") && len(img.ID) > 19 {
			shortID = img.ID[7:19]
		}
		rows = append(rows, row{
			ID:         shortID,
			Repository: repos,
			Tags:       tags,
			Size:       formatBytes(uint64(img.Size)),
			Created:    time.Unix(img.Created, 0).UTC().Format(time.RFC3339),
		})
	}

	out, _ := json.MarshalIndent(rows, "", "  ")
	return mcp.TextResult(string(out)), nil
}

// ImagePull pulls an image from a registry.
func (c *Client) ImagePull(ctx context.Context, ref, platform string) (*mcp.CallToolResult, error) {
	if ref == "" {
		return nil, fmt.Errorf("image is required")
	}

	opts := types.ImagePullOptions{}
	if platform != "" {
		opts.Platform = platform
	}

	reader, err := c.cli.ImagePull(ctx, ref, opts)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// Collect pull progress summary
	var sb strings.Builder
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		var event struct {
			Status string `json:"status"`
			Error  string `json:"error,omitempty"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &event); err == nil {
			if event.Error != "" {
				return nil, fmt.Errorf("pull error: %s", event.Error)
			}
			if event.Status != "" && !strings.HasPrefix(event.Status, "Pulling from") {
				// Skip noisy per-layer progress lines
				continue
			}
			if event.Status != "" {
				sb.WriteString(event.Status + "\n")
			}
		}
	}

	return mcp.TextResult(fmt.Sprintf("Pulled %s\n%s", ref, sb.String())), nil
}

// ImageRemove removes an image.
func (c *Client) ImageRemove(ctx context.Context, ref string, force bool) (*mcp.CallToolResult, error) {
	if ref == "" {
		return nil, fmt.Errorf("image is required")
	}

	items, err := c.cli.ImageRemove(ctx, ref, types.ImageRemoveOptions{
		Force:         force,
		PruneChildren: true,
	})
	if err != nil {
		return nil, err
	}

	type item struct {
		Untagged string `json:"untagged,omitempty"`
		Deleted  string `json:"deleted,omitempty"`
	}
	rows := make([]item, 0, len(items))
	for _, i := range items {
		rows = append(rows, item{Untagged: i.Untagged, Deleted: i.Deleted})
	}
	out, _ := json.MarshalIndent(rows, "", "  ")
	return mcp.TextResult(string(out)), nil
}

// ImageInspect returns detailed info about an image.
func (c *Client) ImageInspect(ctx context.Context, ref string) (*mcp.CallToolResult, error) {
	if ref == "" {
		return nil, fmt.Errorf("image is required")
	}
	info, _, err := c.cli.ImageInspectWithRaw(ctx, ref)
	if err != nil {
		return nil, err
	}
	out, _ := json.MarshalIndent(info, "", "  ")
	return mcp.TextResult(string(out)), nil
}

// ImageTag creates a new tag for an existing image.
func (c *Client) ImageTag(ctx context.Context, source, target string) (*mcp.CallToolResult, error) {
	if source == "" || target == "" {
		return nil, fmt.Errorf("source and target are required")
	}
	if err := c.cli.ImageTag(ctx, source, target); err != nil {
		return nil, err
	}
	return ok(fmt.Sprintf("Tagged %s as %s", source, target))
}

// ImageBuild builds an image from a Dockerfile context directory.
func (c *Client) ImageBuild(ctx context.Context, args map[string]any) (*mcp.CallToolResult, error) {
	contextPath := getStr(args, "context_path", "")
	if contextPath == "" {
		return nil, fmt.Errorf("context_path is required")
	}

	tarReader, err := buildContextTar(contextPath)
	if err != nil {
		return nil, fmt.Errorf("creating build context: %w", err)
	}

	dockerfile := getStr(args, "dockerfile", "Dockerfile")
	tag := getStr(args, "tag", "")
	noCache := getBool(args, "no_cache", false)
	buildArgs := parseBuildArgs(getStrSlice(args, "build_args"))

	tags := []string{}
	if tag != "" {
		tags = []string{tag}
	}

	resp, err := c.cli.ImageBuild(ctx, tarReader, types.ImageBuildOptions{
		Dockerfile: dockerfile,
		Tags:       tags,
		NoCache:    noCache,
		BuildArgs:  buildArgs,
		Remove:     true,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var sb strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var event struct {
			Stream string `json:"stream,omitempty"`
			Error  string `json:"error,omitempty"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &event); err == nil {
			if event.Error != "" {
				return mcp.TextResult("Build failed: " + event.Error), nil
			}
			sb.WriteString(event.Stream)
		}
	}

	return mcp.TextResult(sb.String()), nil
}

// ─── Build Context Helpers ────────────────────────────────────────────────────

func buildContextTar(contextPath string) (io.Reader, error) {
	pr, pw := io.Pipe()

	go func() {
		tw := tar.NewWriter(pw)
		err := filepath.Walk(contextPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			base := filepath.Base(path)
			if base == ".git" || base == ".DS_Store" {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if info.IsDir() {
				return nil
			}
			relPath, err := filepath.Rel(contextPath, path)
			if err != nil {
				return err
			}
			hdr := &tar.Header{
				Name:    filepath.ToSlash(relPath),
				Mode:    int64(info.Mode()),
				Size:    info.Size(),
				ModTime: info.ModTime(),
			}
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = io.Copy(tw, f)
			return err
		})
		tw.Close()
		pw.CloseWithError(err)
	}()

	return pr, nil
}

func parseBuildArgs(args []string) map[string]*string {
	result := map[string]*string{}
	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 2 {
			val := parts[1]
			result[parts[0]] = &val
		}
	}
	return result
}
