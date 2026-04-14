package docker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"docker-mcp/internal/mcp"
)

// SystemInfo returns system-wide Docker information.
func (c *Client) SystemInfo(ctx context.Context) (*mcp.CallToolResult, error) {
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

	out, _ := json.MarshalIndent(summary, "", "  ")
	return mcp.TextResult(string(out)), nil
}

// SystemVersion returns Docker client and server version details.
func (c *Client) SystemVersion(ctx context.Context) (*mcp.CallToolResult, error) {
	v, err := c.cli.ServerVersion(ctx)
	if err != nil {
		return nil, err
	}
	out, _ := json.MarshalIndent(v, "", "  ")
	return mcp.TextResult(string(out)), nil
}

// SystemDf returns disk usage for images, containers, volumes, and build cache.
func (c *Client) SystemDf(ctx context.Context) (*mcp.CallToolResult, error) {
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

	out, _ := json.MarshalIndent(summary, "", "  ")
	return mcp.TextResult(string(out)), nil
}

// SystemPrune removes unused Docker resources.
func (c *Client) SystemPrune(ctx context.Context, all, pruneVolumes bool) (*mcp.CallToolResult, error) {
	pruneFilters := filters.NewArgs()

	// Prune containers
	containerReport, err := c.cli.ContainersPrune(ctx, pruneFilters)
	if err != nil {
		return nil, fmt.Errorf("container prune: %w", err)
	}

	// Prune images
	imageReport, err := c.cli.ImagesPrune(ctx, pruneFilters)
	if err != nil {
		return nil, fmt.Errorf("image prune: %w", err)
	}

	// Prune networks
	networkReport, err := c.cli.NetworksPrune(ctx, pruneFilters)
	if err != nil {
		return nil, fmt.Errorf("network prune: %w", err)
	}

	// Prune build cache
	buildCacheReport, err := c.cli.BuildCachePrune(ctx, types.BuildCachePruneOptions{All: all})
	if err != nil {
		return nil, fmt.Errorf("build cache prune: %w", err)
	}

	result := map[string]any{
		"containers_deleted": containerReport.ContainersDeleted,
		"space_reclaimed_containers": formatBytes(containerReport.SpaceReclaimed),
		"images_deleted":     len(imageReport.ImagesDeleted),
		"space_reclaimed_images": formatBytes(imageReport.SpaceReclaimed),
		"networks_deleted":   networkReport.NetworksDeleted,
		"build_cache_deleted": len(buildCacheReport.CachesDeleted),
		"space_reclaimed_cache": formatBytes(buildCacheReport.SpaceReclaimed),
	}

	// Optionally prune volumes
	if pruneVolumes {
		volReport, err := c.cli.VolumesPrune(ctx, pruneFilters)
		if err != nil {
			return nil, fmt.Errorf("volume prune: %w", err)
		}
		result["volumes_deleted"] = volReport.VolumesDeleted
		result["space_reclaimed_volumes"] = formatBytes(volReport.SpaceReclaimed)
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return mcp.TextResult(string(out)), nil
}
