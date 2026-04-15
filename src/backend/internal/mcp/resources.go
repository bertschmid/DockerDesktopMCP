package mcp

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

const resourceMIMEType = "text/html;profile=mcp-app"

//go:embed ui-apps-dist/*.html
var uiAppsFS embed.FS

type registeredResource struct {
	URI  string
	Name string
	File string
}

var uiResources = []registeredResource{
	{URI: "ui://docker-desktop/containers", Name: "Containers List", File: "ui-apps-dist/containers.html"},
	{URI: "ui://docker-desktop/images", Name: "Images List", File: "ui-apps-dist/images.html"},
	{URI: "ui://docker-desktop/volumes", Name: "Volumes List", File: "ui-apps-dist/volumes.html"},
	{URI: "ui://docker-desktop/networks", Name: "Networks List", File: "ui-apps-dist/networks.html"},
	{URI: "ui://docker-desktop/compose-services", Name: "Compose Services", File: "ui-apps-dist/compose.html"},
	{URI: "ui://docker-desktop/disk-usage", Name: "Disk Usage", File: "ui-apps-dist/disk-usage.html"},
	{URI: "ui://docker-desktop/system-info", Name: "System Info", File: "ui-apps-dist/system-info.html"},
}

func (s *Server) listResources() ListResourcesResult {
	resources := make([]Resource, 0, len(uiResources))
	for _, r := range uiResources {
		resources = append(resources, Resource{
			URI:      r.URI,
			Name:     r.Name,
			MimeType: resourceMIMEType,
		})
	}
	return ListResourcesResult{Resources: resources}
}

func (s *Server) readResource(params ReadResourceParams) (*ReadResourceResult, error) {
	if params.URI == "" {
		return nil, fmt.Errorf("uri is required")
	}

	for _, r := range uiResources {
		if r.URI == params.URI {
			htmlBytes, err := readUIResourceFile(r.File)
			if err != nil {
				return nil, fmt.Errorf("reading resource %s: %w", r.File, err)
			}
			return &ReadResourceResult{
				Contents: []ResourceContent{{
					URI:      r.URI,
					MimeType: resourceMIMEType,
					Text:     string(htmlBytes),
				}},
			}, nil
		}
	}

	return nil, fmt.Errorf("unknown resource uri: %s", params.URI)
}

func readUIResourceFile(embedPath string) ([]byte, error) {
	htmlBytes, err := uiAppsFS.ReadFile(embedPath)
	if err == nil {
		return htmlBytes, nil
	}

	base := filepath.Base(embedPath)
	devCandidates := []string{
		filepath.Join("..", "ui-apps", "dist", base),
		filepath.Join("src", "ui-apps", "dist", base),
	}

	for _, p := range devCandidates {
		if b, readErr := os.ReadFile(p); readErr == nil {
			return b, nil
		}
	}

	return nil, err
}
