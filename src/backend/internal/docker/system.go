package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"docker-mcp/internal/result"
)

func buildPruneFilters(filterSpecs []string) (filters.Args, error) {
	args := filters.NewArgs()

	for _, spec := range filterSpecs {
		spec = strings.TrimSpace(spec)
		if spec == "" {
			continue
		}

		if key, value, ok := strings.Cut(spec, "!="); ok {
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			if key == "" || value == "" {
				return filters.Args{}, fmt.Errorf("invalid prune filter %q", spec)
			}
			args.Add(key+"!", value)
			continue
		}

		if key, value, ok := strings.Cut(spec, "="); ok {
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			if key == "" || value == "" {
				return filters.Args{}, fmt.Errorf("invalid prune filter %q", spec)
			}
			args.Add(key, value)
			continue
		}

		args.Add(spec, "true")
	}

	return args, nil
}

func marshalSummaryResult(summary map[string]any) (*result.CallToolResult, error) {
	out, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal prune summary: %w", err)
	}
	return result.TextStructured(string(out), summary), nil
}

func pruneUnusedImageFilters(filterSpecs []string) (filters.Args, error) {
	args, err := buildPruneFilters(filterSpecs)
	if err != nil {
		return filters.Args{}, err
	}
	args.Add("dangling", "false")
	return args, nil
}

func (c *Client) pruneContainers(ctx context.Context, filterSpecs []string) (map[string]any, error) {
	pruneFilters, err := buildPruneFilters(filterSpecs)
	if err != nil {
		return nil, err
	}

	report, err := c.cli.ContainersPrune(ctx, pruneFilters)
	if err != nil {
		return nil, fmt.Errorf("container prune: %w", err)
	}

	return map[string]any{
		"scope":             "containers",
		"filters":           filterSpecs,
		"containers_deleted": report.ContainersDeleted,
		"space_reclaimed":   formatBytes(report.SpaceReclaimed),
	}, nil
}

func (c *Client) pruneImages(ctx context.Context, filterSpecs []string) (map[string]any, error) {
	pruneFilters, err := pruneUnusedImageFilters(filterSpecs)
	if err != nil {
		return nil, err
	}

	report, err := c.cli.ImagesPrune(ctx, pruneFilters)
	if err != nil {
		return nil, fmt.Errorf("image prune: %w", err)
	}

	return map[string]any{
		"scope":           "images",
		"filters":         filterSpecs,
		"images_deleted":  len(report.ImagesDeleted),
		"space_reclaimed": formatBytes(report.SpaceReclaimed),
	}, nil
}

func (c *Client) pruneNetworks(ctx context.Context, filterSpecs []string) (map[string]any, error) {
	pruneFilters, err := buildPruneFilters(filterSpecs)
	if err != nil {
		return nil, err
	}

	report, err := c.cli.NetworksPrune(ctx, pruneFilters)
	if err != nil {
		return nil, fmt.Errorf("network prune: %w", err)
	}

	return map[string]any{
		"scope":            "networks",
		"filters":          filterSpecs,
		"networks_deleted": report.NetworksDeleted,
	}, nil
}

func (c *Client) pruneBuildCache(ctx context.Context, filterSpecs []string) (map[string]any, error) {
	pruneFilters, err := buildPruneFilters(filterSpecs)
	if err != nil {
		return nil, err
	}

	report, err := c.cli.BuildCachePrune(ctx, types.BuildCachePruneOptions{All: true, Filters: pruneFilters})
	if err != nil {
		return nil, fmt.Errorf("build cache prune: %w", err)
	}

	return map[string]any{
		"scope":               "build_cache",
		"filters":             filterSpecs,
		"build_cache_deleted": len(report.CachesDeleted),
		"space_reclaimed":     formatBytes(report.SpaceReclaimed),
	}, nil
}

func (c *Client) pruneVolumes(ctx context.Context, filterSpecs []string) (map[string]any, error) {
	pruneFilters, err := buildPruneFilters(filterSpecs)
	if err != nil {
		return nil, err
	}

	report, err := c.cli.VolumesPrune(ctx, pruneFilters)
	if err != nil {
		return nil, fmt.Errorf("volume prune: %w", err)
	}

	return map[string]any{
		"scope":            "volumes",
		"filters":          filterSpecs,
		"volumes_deleted":  report.VolumesDeleted,
		"space_reclaimed": formatBytes(report.SpaceReclaimed),
	}, nil
}

// SystemInfo returns system-wide Docker information.
func (c *Client) SystemInfo(ctx context.Context) (*result.CallToolResult, error) {
	info, err := c.cli.Info(ctx)
	if err != nil {
		return nil, err
	}

	summary := map[string]any{
		"id":                   info.ID,
		"hostname":             info.Name,
		"server_version":       info.ServerVersion,
		"os":                   info.OperatingSystem,
		"os_type":              info.OSType,
		"architecture":         info.Architecture,
		"kernel_version":       info.KernelVersion,
		"cpus":                 info.NCPU,
		"memory":               formatBytes(uint64(info.MemTotal)),
		"storage_driver":       info.Driver,
		"containers":           info.Containers,
		"containers_running":   info.ContainersRunning,
		"containers_paused":    info.ContainersPaused,
		"containers_stopped":   info.ContainersStopped,
		"images":               info.Images,
		"docker_root_dir":      info.DockerRootDir,
		"logging_driver":       info.LoggingDriver,
		"cgroup_driver":        info.CgroupDriver,
		"swarm":                info.Swarm.LocalNodeState,
		"experimental":         info.ExperimentalBuild,
		"live_restore_enabled": info.LiveRestoreEnabled,
	}

	out, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal system info: %w", err)
	}
	return result.TextStructuredUI(
		string(out),
		map[string]any{"info": summary},
		"ui://docker-desktop/system-info",
	), nil
}

// SystemVersion returns Docker client and server version details.
func (c *Client) SystemVersion(ctx context.Context) (*result.CallToolResult, error) {
	v, err := c.cli.ServerVersion(ctx)
	if err != nil {
		return nil, err
	}
	out, _ := json.MarshalIndent(v, "", "  ")
	return result.Text(string(out)), nil
}

// SystemDf returns disk usage for images, containers, volumes, and build cache.
func (c *Client) SystemDf(ctx context.Context) (*result.CallToolResult, error) {
	du, err := c.cli.DiskUsage(ctx, types.DiskUsageOptions{})
	if err != nil {
		return nil, err
	}

	// Images summary
	var imageSize int64
	for _, img := range du.Images {
		imageSize += img.Size
	}

	// Containers summary
	var containerRWSize int64
	for _, ct := range du.Containers {
		containerRWSize += ct.SizeRw
	}

	// Volumes summary
	var volumeSize int64
	for _, v := range du.Volumes {
		if v.UsageData != nil {
			volumeSize += v.UsageData.Size
		}
	}

	// Build cache summary
	var cacheSize int64
	for _, bc := range du.BuildCache {
		cacheSize += bc.Size
	}

	summary := map[string]any{
		"images": map[string]any{
			"count":      len(du.Images),
			"total_size": formatBytes(uint64(imageSize)),
		},
		"containers": map[string]any{
			"count":   len(du.Containers),
			"rw_size": formatBytes(uint64(containerRWSize)),
		},
		"volumes": map[string]any{
			"count":      len(du.Volumes),
			"total_size": formatBytes(uint64(volumeSize)),
		},
		"build_cache": map[string]any{
			"count":      len(du.BuildCache),
			"total_size": formatBytes(uint64(cacheSize)),
		},
		"total_reclaimable": formatBytes(uint64(imageSize + containerRWSize + volumeSize + cacheSize)),
	}

	out, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal system df: %w", err)
	}
	return result.TextStructuredUI(
		string(out),
		map[string]any{"disk_usage": summary},
		"ui://docker-desktop/disk-usage",
	), nil
}

// SystemPruneAll removes unused containers, images, networks, build cache, and volumes in one operation.
func (c *Client) SystemPruneAll(ctx context.Context, filterSpecs []string) (*result.CallToolResult, error) {
	containerSummary, err := c.pruneContainers(ctx, filterSpecs)
	if err != nil {
		return nil, err
	}

	imageSummary, err := c.pruneImages(ctx, filterSpecs)
	if err != nil {
		return nil, err
	}

	networkSummary, err := c.pruneNetworks(ctx, filterSpecs)
	if err != nil {
		return nil, err
	}

	buildCacheSummary, err := c.pruneBuildCache(ctx, filterSpecs)
	if err != nil {
		return nil, err
	}

	volumeSummary, err := c.pruneVolumes(ctx, filterSpecs)
	if err != nil {
		return nil, err
	}

	summary := map[string]any{
		"scope":       "all",
		"filters":     filterSpecs,
		"containers":  containerSummary,
		"images":      imageSummary,
		"networks":    networkSummary,
		"build_cache": buildCacheSummary,
		"volumes":     volumeSummary,
	}

	return marshalSummaryResult(summary)
}

// SystemPruneContainers removes stopped containers and container resources that are no longer used.
func (c *Client) SystemPruneContainers(ctx context.Context, filterSpecs []string) (*result.CallToolResult, error) {
	summary, err := c.pruneContainers(ctx, filterSpecs)
	if err != nil {
		return nil, err
	}
	return marshalSummaryResult(summary)
}

// SystemPruneImages removes unused images, not just dangling ones.
func (c *Client) SystemPruneImages(ctx context.Context, filterSpecs []string) (*result.CallToolResult, error) {
	summary, err := c.pruneImages(ctx, filterSpecs)
	if err != nil {
		return nil, err
	}
	return marshalSummaryResult(summary)
}

// SystemPruneNetworks removes unused Docker networks that have no active attachments.
func (c *Client) SystemPruneNetworks(ctx context.Context, filterSpecs []string) (*result.CallToolResult, error) {
	summary, err := c.pruneNetworks(ctx, filterSpecs)
	if err != nil {
		return nil, err
	}
	return marshalSummaryResult(summary)
}

// SystemPruneBuildCache removes all unused Docker build cache records.
func (c *Client) SystemPruneBuildCache(ctx context.Context, filterSpecs []string) (*result.CallToolResult, error) {
	summary, err := c.pruneBuildCache(ctx, filterSpecs)
	if err != nil {
		return nil, err
	}
	return marshalSummaryResult(summary)
}

// SystemPruneVolumes removes unused Docker volumes that are not referenced by containers.
func (c *Client) SystemPruneVolumes(ctx context.Context, filterSpecs []string) (*result.CallToolResult, error) {
	summary, err := c.pruneVolumes(ctx, filterSpecs)
	if err != nil {
		return nil, err
	}
	return marshalSummaryResult(summary)
}
