// Package result defines the shared MCP tool result types used by both the
// docker and mcp packages. Keeping them here breaks the import cycle.
package result

// UIMeta holds MCP Apps UI metadata for a tool result.
type UIMeta struct {
	ResourceURI string `json:"resourceUri"`
}

// ResultMeta is the top-level metadata container for CallToolResult.
type ResultMeta struct {
	UI *UIMeta `json:"ui,omitempty"`
}

// CallToolResult is the MCP tools/call response payload.
type CallToolResult struct {
	Content           []Content      `json:"content"`
	StructuredContent map[string]any `json:"structuredContent,omitempty"`
	Meta              *ResultMeta    `json:"_meta,omitempty"`
	IsError           bool           `json:"isError,omitempty"`
}

// Content is a single piece of content inside a CallToolResult.
type Content struct {
	Type string `json:"type"` // "text" | "image" | "resource"
	Text string `json:"text,omitempty"`
}

// Text creates a successful text tool result.
func Text(text string) *CallToolResult {
	return &CallToolResult{Content: []Content{{Type: "text", Text: text}}}
}

// TextStructured creates a successful text tool result with structuredContent.
func TextStructured(text string, structured map[string]any) *CallToolResult {
	return &CallToolResult{
		Content:           []Content{{Type: "text", Text: text}},
		StructuredContent: structured,
	}
}

// TextStructuredUI creates a successful text result with structured data and UI metadata.
func TextStructuredUI(text string, structured map[string]any, resourceURI string) *CallToolResult {
	res := TextStructured(text, structured)
	if resourceURI != "" {
		res.Meta = &ResultMeta{UI: &UIMeta{ResourceURI: resourceURI}}
	}
	return res
}

// Error creates an error tool result.
func Error(msg string) *CallToolResult {
	return &CallToolResult{
		IsError: true,
		Content: []Content{{Type: "text", Text: msg}},
	}
}
